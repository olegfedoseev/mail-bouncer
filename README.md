# Pechkin

Email validation service with RESTish API

## Goal

Often it's not enough to just validate format of email.
Even if format is Ok, you may fail to send a message to it.
You can have an invalid domain, without MX record, a user can be over quota. Or already deleted.

And Pechkin helps you to find out about it before you send a message.

## Usage
Get it with:
`go get github.com/olegfedoseev/pechkin`

Start it with:
`pechkin --listen=<listen> --host=<host> --from=<from>`
You have to specify valid hostname for "host" (see HELO command in SMTP) and valid email for "from" (see MAIL command in SMTP)

Then all you need is simple GET:

	> curl -i "http://127.0.0.1:8080/?email=invalid@email.com"
	HTTP/1.1 417 Expectation Failed

	RCPT failed for invalid@email.com: 554 5.7.1 Helo command rejected


	> curl -i "http://127.0.0.1:8080/?email=valid@gamil.com"
	HTTP/1.1 200 OK

Or with callback:

	> curl -i "http://127.0.0.1:8080/?email=invalid@email.com&callback=$URL"
	HTTP/1.1 201 Created

And you will get POST to $URL with JSON data, ex.:

	{
		"email": "invalid@email.com",
		"valid": false,
		"error": "RCPT failed for invalid@email.com: 554 5.7.1 Helo command rejected"
	}
