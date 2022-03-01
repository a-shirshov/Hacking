.PHONY: build
build:
	make proxy
	make web


.PHONY: proxy
proxy:
	go build -o ./proxy -v ./cmd/proxy

.PHONY: web
web:
	go build -o ./web -v ./cmd/web