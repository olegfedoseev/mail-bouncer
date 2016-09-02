FROM alpine:3.1

RUN ln -sf /usr/share/zoneinfo/Asia/Novosibirsk /etc/localtime
ADD mail-bouncer /mail-bouncer

ENTRYPOINT ["/mail-bouncer"]
CMD ["--listen=0.0.0.0:80", "--host=localhost", "--from=tester@example.com"]
EXPOSE 80
