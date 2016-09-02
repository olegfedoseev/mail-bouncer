package main

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
)

func main() {
	usage := `mail-bouncer 1.0

mail-bouncer is email validation service with simple RESTish API.

Example:
	> curl -i ":8080/?email=invalid@email.com"
	HTTP/1.1 200 OK

	{
		"email":"invalid@email.com",
		"is_valid":false,
		"description":"MX server is unreachable",
		"error":"can't connect to email.com: ...."
	}

	> curl -i "http://127.0.0.1:8080/?email=valid@gamil.com"
	HTTP/1.1 200 OK

	> curl -i "http://127.0.0.1:8080/?email=invalid@email.com&callback=$URL"
	HTTP/1.1 201 Created

	And POST to $URL with json data, ex.:
	{
		"email": "invalid@email.com",
		"valid": false,
		"error": "RCPT failed for invalid@email.com: 554 5.7.1 Helo command rejected"
	}

Usage:
	mail-bouncer [options]
	mail-bouncer -v | -h

Options:
  -h --help         Show this screen.
  -v --version      Show version.
  --listen=<listen> Host and port for http interface
  --host=<host>     Hostname for HELO command
  --from=<from>     Email address for MAIL command
`
	args, err := docopt.Parse(usage, nil, true, "mail-bouncer 1.0", false)
	if err != nil {
		log.Fatal(err)
	}

	listen := ":8080"
	if args["--listen"] != nil {
		listen = args["--listen"].(string)
	}

	mailHostname := "localhost"
	if args["--host"] != nil {
		mailHostname = args["--host"].(string)
	}

	fromAddress := "tester@localhost"
	if args["--from"] != nil {
		fromAddress = args["--from"].(string)
	}

	http.Handle("/", NewHandler(mailHostname, fromAddress))

	log.WithFields(log.Fields{
		"pid":    os.Getpid(),
		"listen": listen,
	}).Info("Ready")
	log.Fatal(http.ListenAndServe(listen, nil))
}
