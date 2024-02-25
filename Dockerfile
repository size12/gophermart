FROM golang:alpine

COPY . /service
WORKDIR /service

RUN go build cmd/gophermart/main.go