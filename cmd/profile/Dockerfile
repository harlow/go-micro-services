FROM golang:1.7
EXPOSE 8080
ADD . /go/src/github.com/harlow/go-micro-services/cmd/profile
WORKDIR /go/src/github.com/harlow/go-micro-services/cmd/profile
RUN go build .
CMD ["/go/src/github.com/harlow/go-micro-services/cmd/profile/profile"]