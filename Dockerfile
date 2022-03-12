# syntax=docker/dockerfile:1

FROM golang:latest as build
LABEL maintainer="Artyom <artyomsh01@yandex.ru>"
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . /app
RUN make build

FROM golang:latest as proxy
LABEL maintainer="Artyom <artyomsh01@yandex.ru>"
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . /app
RUN make proxy
CMD ["./proxy"]

FROM golang:latest as web
LABEL maintainer="Artyom <artyomsh01@yandex.ru>"
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . /app
RUN make web
CMD ["./web"]