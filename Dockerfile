FROM golang:1.7
COPY . /go/src/github.com/harlow/go-micro-services
WORKDIR /go/src/github.com/harlow/go-micro-services
RUN cd cmd/api && go build .
RUN cd cmd/search && go build .
RUN cd cmd/geo && go build .
RUN cd cmd/profile && go build .
RUN cd cmd/rate && go build .
RUN cd cmd/www && go build .
