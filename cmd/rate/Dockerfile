FROM golang:1.7
EXPOSE 8080
ADD . /go/src/github.com/harlow/go-micro-services/cmd/rate
WORKDIR /go/src/github.com/harlow/go-micro-services/cmd/rate
RUN go build .
CMD ["/go/src/github.com/harlow/go-micro-services/cmd/rate/rate"]