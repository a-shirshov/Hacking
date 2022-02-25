# syntax=docker/dockerfile:1

FROM golang:1.17-alpine
WORKDIR /app
COPY * ./

RUN go mod download

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache openssl

RUN go build -o /proxy 
RUN ./gen_ca.sh 
CMD ["/proxy"]