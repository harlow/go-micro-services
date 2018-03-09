FROM golang:1.9
COPY . /go/src/github.com/harlow/go-micro-services
WORKDIR /go/src/github.com/harlow/go-micro-services
RUN go install -ldflags="-s -w" ./cmd/...
