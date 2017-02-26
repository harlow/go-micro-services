FROM golang:1.7
EXPOSE 8080
ADD . /go/src/github.com/harlow/go-micro-services/cmd/api
WORKDIR /go/src/github.com/harlow/go-micro-services/cmd/api
RUN go build .
CMD ["/go/src/github.com/harlow/go-micro-services/cmd/api/api"]