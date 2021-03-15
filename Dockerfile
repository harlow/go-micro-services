FROM golang:1.12
COPY . /go/src/github.com/harlow/go-micro-services
WORKDIR /go/src/github.com/harlow/go-micro-services

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io

RUN go install -ldflags="-s -w" ./cmd/...
