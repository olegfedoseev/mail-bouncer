all: build test

build:
	docker build --rm -t mail-bouncer-builder --file Dockerfile.build .
	docker create --name mail-bouncer-builder mail-bouncer-builder /bin/true
	docker cp mail-bouncer-builder:/mail-bouncer mail-bouncer
	docker rm -f mail-bouncer-builder || true
	docker build --rm -t olegfedoseev/mail-bouncer .

push: build
	docker push olegfedoseev/mail-bouncer

test: build
	docker rm -f mail-bouncer || true
	docker run -d --name mail-bouncer -p 8080:80 olegfedoseev/mail-bouncer
	docker logs -f mail-bouncer

update:
	git fetch && git pull --rebase
