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
// for validation process
type MailHandler struct {
	Hostname string
	From     string
}

// NewHandler creates new handler with given host and from address
func NewHandler(host, from string) *MailHandler {
	return &MailHandler{
		Hostname: host,
		From:     from,
	}
}

type apiResponse struct {
	Email       string `json:"email"`
	IsValid     bool   `json:"is_valid"`
	Description string `json:"description"`
	Error       string `json:"error"`
}

// ServerHTTP is handler for RESTish API
// It understands two arguments: email and callback
// Example: GET /?email=test@email.com
// Example: POST /?email=test@email.com&callback=http://test.com/emailCallback
// Result wiil be in bool field "is_valid"
func (handler *MailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cache-Control")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	callback := r.URL.Query().Get("callback")
	if callback != "" && r.Method == "POST" {
		log.WithFields(log.Fields{
			"email":    email,
			"callback": callback,
		}).Info("Callback requested")

		w.WriteHeader(http.StatusCreated)
		go handler.handleCallback(email, callback)
		return
	}

	resp := handler.validateEmail(email)
	logFields := log.Fields{"email": email, "valid": resp.IsValid}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.WithFields(logFields).Errorf("Failed to encode callback data: %v", err)
	}
	log.WithFields(logFields).Info("Email validated")
}

func (handler *MailHandler) handleCallback(email, callback string) {
	validateResp := handler.validateEmail(email)
	logFields := log.Fields{
		"email":    email,
		"valid":    validateResp.IsValid,
		"callback": callback,
		"error":    validateResp.Error,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(validateResp); err != nil {
		log.WithFields(logFields).Errorf("Failed to encode callback data: %v", err)

		// Fallback
		buf.WriteString(fmt.Sprintf("{\"email\": \"%s\", \"valid\": %v, \"error\": \"\"}",
			email, validateResp.IsValid))
	}

	resp, err := http.Post(callback, "application/json", &buf)
	if err != nil {
		log.WithFields(logFields).Errorf("Failed to POST: %v", err)
		return
	}

	if err := resp.Body.Close(); err != nil {
		log.Errorf("resp.Body.Close() failed: %v", err)
	}
}

func (handler *MailHandler) validateEmail(emailAddr string) *apiResponse {
	response := &apiResponse{
		Email:       emailAddr,
		IsValid:     false,
		Description: "Unknown",
		Error:       "",
	}

	logFields := log.Fields{
		"email": emailAddr,
	}

	// Step 1: validate format
	email, err := mail.ParseAddress(emailAddr)
	if err != nil {
		response.Error = fmt.Sprintf("invalid email: %v", err)
		response.Description = "Format validation failed"
		return response
	}
	log.WithFields(logFields).Debug("Format is valid")

	parts := strings.SplitN(email.Address, "@", 2)
	domain := parts[1]

	// Step 2: validate MX record
	// TODO: add cache
	nss, err := net.LookupMX(domain)
	if err != nil {
		response.Error = fmt.Sprintf("MX lookup failed for %v: %v", domain, err)
		response.Description = "Coudn't find MX server for this address"
		return response
	}

	if len(nss) == 0 {
		response.Error = fmt.Sprintf("no MX records found for %v", domain)
		response.Description = "Coudn't find MX server for this address"
		return response
	}
	log.WithFields(logFields).Debugf("Found MX servers: %v", nss)

	// Step 3: try to "send" email
	// TODO: add connection pooling?
	// TODO: add timeout?
	client, err := smtp.Dial(nss[0].Host + ":25")
	if err != nil {
		response.Error = fmt.Sprintf("can't connect to %s: %v", domain, err)
		response.Description = "MX server is unreachable"
		return response
	}
	log.WithFields(logFields).Debugf("Connected to: %v", nss[0].Host+":25")
	if err := client.Hello(handler.Hostname); err != nil {
		response.Error = fmt.Sprintf("HELO failed: %v", err)
		return response
	}
	log.WithFields(logFields).Debug("HELO is ok")
	if err := client.Mail(handler.From); err != nil {
		response.Error = fmt.Sprintf("MAIL failed: %v", err)
		return response
	}
	log.WithFields(logFields).Debug("MAIL is ok")
	if err := client.Rcpt(email.Address); err != nil {
		response.Error = fmt.Sprintf("RCPT failed for %s: %v", email.Address, err)
		response.Description = err.Error()
		return response
	}
	log.WithFields(logFields).Debug("RCPT is ok")
	if err := client.Quit(); err != nil {
		response.Error = fmt.Sprintf("QUIT failed: %v", err)
		return response
	}
	log.WithFields(logFields).Debug("QUIT is ok")

	response.Description = "Ok"
	response.IsValid = true
	return response
}
