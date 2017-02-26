FROM golang:1.7
EXPOSE 8080
ADD . /go/src/github.com/harlow/go-micro-services/cmd/geo
WORKDIR /go/src/github.com/harlow/go-micro-services/cmd/geo
RUN go build .
CMD ["/go/src/github.com/harlow/go-micro-services/cmd/geo/geo"]