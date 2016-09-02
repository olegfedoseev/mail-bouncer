FROM golang:1.7-alpine

RUN apk add --update git

RUN go get -v -d github.com/Sirupsen/logrus && \
	go get -v -d github.com/docopt/docopt-go

WORKDIR /go/src/github.com/olegfedoseev/mail-bouncer
ADD . /go/src/github.com/olegfedoseev/mail-bouncer
RUN go build -o /mail-bouncer
