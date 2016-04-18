package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// MailHandler implements simple RESTish API and holds host and from address
// for validation proccess
type MailHandler struct {
	Hostname string
	From     string
}

// NewHandler creates new handler with given host and from address
func NewHandler(host, from string) *MailHandler {
	handler := MailHandler{
		Hostname: host,
		From:     from,
	}
	return &handler
}

// ServerHTTP is handler for RESTish API
// It understands to get arguments: email and callback
// Example: GET /?email=test@email.com
// Example: GET /?email=test@email.com&callback=http://test.com/emailCallback
// If email valid it'll return 200 Ok, if not 417 Expectation Failed with
// detailed error message
func (handler *MailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	callback := r.URL.Query().Get("callback")
	if callback != "" {
		log.WithFields(log.Fields{
			"email":    email,
			"callback": callback,
		}).Info("Callback requested")

		w.WriteHeader(http.StatusCreated)
		go handler.handleCallback(email, callback)
		return
	}

	if err := handler.validateEmail(email); err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		fmt.Fprintln(w, err)
		log.WithFields(log.Fields{
			"email": email,
			"error": err,
		}).Info("Email is invalid")
		return
	}

	log.WithFields(log.Fields{"email": email}).Info("Email is valid")
	w.WriteHeader(http.StatusOK)
}

func (handler *MailHandler) handleCallback(email, callback string) {
	validateError := handler.validateEmail(email)
	isValid := (validateError == nil)
	// body for POST
	callbackData := map[string]interface{}{
		"email": email,
		"valid": isValid,
		"error": validateError,
	}
	if !isValid {
		callbackData["error"] = validateError.Error()
	}

	logFields := log.Fields{
		"email":    email,
		"valid":    isValid,
		"callback": callback,
		"error":    validateError,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(callbackData); err != nil {
		log.WithFields(logFields).Errorf("Failed to encode callback data: %v", err)

		// Fallback
		buf.WriteString(fmt.Sprintf("{\"email\": \"%s\", \"valid\": %v, \"error\": \"\"}",
			email, isValid))
	}

	resp, err := http.Post(callback, "application/json", &buf)
	if err != nil {
		log.WithFields(logFields).Errorf("Failed to POST: %v", err)
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Errorf("resp.Body.Close() failed: %v", err)
	}

	log.WithFields(logFields).Info("Send POST to callback")
}

func (handler *MailHandler) validateEmail(emailAddr string) error {
	// Step 1: validate format
	email, err := mail.ParseAddress(emailAddr)
	if err != nil {
		return fmt.Errorf("invalid email: %v", err)
	}

	parts := strings.SplitN(email.Address, "@", 2)
	domain := parts[1]

	// Step 2: validate MX record
	nss, err := net.LookupMX(domain)
	if err != nil {
		return fmt.Errorf("MX lookup failed for %v: %v", domain, err)
	}

	if len(nss) == 0 {
		return fmt.Errorf("no MX records found for %v", domain)
	}

	// Step 3: try to "send" email
	// TODO: add connection pooling?
	// TODO: add timeout?
	client, err := smtp.Dial(nss[0].Host + ":25")
	if err != nil {
		return fmt.Errorf("can't connect to %s: %v", domain, err)
	}
	if err := client.Hello(handler.Hostname); err != nil {
		return fmt.Errorf("HELO failed: %v", err)
	}
	if err := client.Mail(handler.From); err != nil {
		return fmt.Errorf("MAIL failed: %v", err)
	}
	if err := client.Rcpt(email.Address); err != nil {
		return fmt.Errorf("RCPT failed for %s: %v", email.Address, err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("QUIT failed: %v", err)
	}

	return nil
}
