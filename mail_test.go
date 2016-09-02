package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	handler := http.Handler(NewHandler("localhost", "tester@example.com"))

	var tests = []struct {
		url  string
		code int
		body string
	}{
		{"/?email=oleg.fedoseev@me.com", 200, ""},
		{"/?invalid=reqeust", 400, ""},
		{"/?email=invalid", 200, "invalid email: mail: missing phrase"},
		{"/?email=invalid@example.com", 200, "MX lookup failed for example.com"},
	}
	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.url, nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, test.code, recorder.Code)
		assert.Contains(t, recorder.Body.String(), test.body)
	}
}

func TestValidationWithCallback(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	handler := http.Handler(NewHandler("localhost", "tester@example.com"))
	callbackBody := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		callbackBody <- string(body)
	}))
	defer ts.Close()

	var tests = []struct {
		url  string
		code int
		body string
	}{
		{
			"/?email=oleg.fedoseev@me.com&callback=" + ts.URL,
			201,
			`"is_valid":true`,
		},
		{
			fmt.Sprintf("/?email=invalid-email%d@gmail.com&callback=%s", time.Now().Unix(), ts.URL),
			201,
			`"is_valid":false`,
		},
	}
	for _, test := range tests {
		req, _ := http.NewRequest("POST", test.url, nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, test.code, recorder.Code)

		// Waiting for callback
		body := <-callbackBody
		assert.Contains(t, body, test.body)
	}
}
