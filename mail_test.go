package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	handler := http.Handler(NewHandler("localhost", "tester@example.com"))

	var tests = []struct {
		url  string
		code int
		body string
	}{
		{"/?email=o.fedoseev@office.ngs.ru", 200, ""},
		{"/?invalid=reqeust", 400, ""},
		{"/?email=invalid", 417, "invalid email: mail: missing phrase\n"},
		{"/?email=invalid@example.com", 417, "MX lookup failed for example.com"},
		{"/?email=qwekfksla@gmail.com", 417, "RCPT failed for qwekfksla@gmail.com: 550 5.1.1"},
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
			"/?email=o.fedoseev@office.ngs.ru&callback=" + ts.URL,
			201,
			`{"email":"o.fedoseev@office.ngs.ru","error":null,"valid":true}` + "\n",
		},
		{
			"/?email=rtrty@rambler.ru&callback=" + ts.URL,
			201,
			`{"email":"rtrty@rambler.ru","error":"RCPT failed for rtrty@rambler.ru: 554 5.7.1 Helo command rejected","valid":false}` + "\n",
		},
	}
	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.url, nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, test.code, recorder.Code)

		// Waiting for callback
		body := <-callbackBody
		assert.Contains(t, body, test.body)
	}
}
