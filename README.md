# mail-bouncer

Email validation service with RESTish API

[![Go Report Card](https://goreportcard.com/badge/github.com/olegfedoseev/mail-bouncer)](https://goreportcard.com/report/github.com/olegfedoseev/mail-bouncer)
[![](https://images.microbadger.com/badges/image/olegfedoseev/mail-bouncer.svg)](http://microbadger.com/images/olegfedoseev/mail-bouncer "Get your own image badge on microbadger.com")

## Goal

Often it's not enough to just validate a format of an email.
Even if a format is Ok, you may fail to send a message to it.
You can have an invalid domain, without MX record, a user can be over quota. Or already deleted.

And mail-bouncer helps you to find out about it before you send a message.

## Usage
Get it with:

	go get github.com/olegfedoseev/mail-bouncer

Start it with:

	mail-bouncer --listen=<listen> --host=<host> --from=<from>

You have to specify valid hostname for "host" (see HELO command in SMTP) and valid email for "from" (see MAIL command in SMTP)

Or using Docker:

	docker run -d -p 80 olegfedoseev/mail-bouncer

Then all you need is simple GET:

	> curl -i ":8080/?email=invalid@email.com"                                                                                                   HTTP/1.1 200 OK
	Date: Fri, 02 Sep 2016 05:02:46 GMT
	Content-Length: 241
	Content-Type: text/plain; charset=utf-8

	{
		"email":"invalid@email.com",
		"is_valid":false,
		"description":"MX server is unreachable",
		"error":"can't connect to email.com: 421 mail.com (mxgmxus004) Nemesis ESMTP Service not available\nRequested action aborted: local error in processing"
	}

	> curl -i "http://127.0.0.1:8080/?email=valid@gmail.com"
	HTTP/1.1 200 OK

	{
		"email":"valid@gmail.com",
		"is_valid":true,
		"description":"Ok",
		"error":""
	}

Or with callback:

	> curl -i -XPOST "http://127.0.0.1:8080/?email=invalid@email.com&callback=$URL"
	HTTP/1.1 201 Created

And you will get POST to $URL with JSON data, ex.:

	{
		"email": "invalid@email.com",
		"is_valid": false,
		"error": "RCPT failed for invalid@email.com: 554 5.7.1 Helo command rejected"
	}
