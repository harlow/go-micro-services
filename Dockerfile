FROM golang:1.11.2
COPY . /go/src/github.com/harlow/go-micro-services
WORKDIR /go/src/github.com/harlow/go-micro-services
RUN go install -ldflags="-s -w" ./cmd/...
