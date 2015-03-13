# HTTP up front, Protobufs in the rear.

An experiment with Golang micro-services that accept external HTTP requests and then
leverage [Protocol Buffers][3] for inter-service tcp communication.

![likes_sequence](https://cloud.githubusercontent.com/assets/739782/6634233/3046c1ec-c912-11e4-96cd-84cf359aa6dc.png)

The API Service accepts HTTP requests on port `8080` and then dials a tcp connection
to the User Service and authenticates the token with the Auth Service.

The applications use Consul for service discovery.

### Installation

Clone the repository:

    git clone git@github.com:harlow/go-micro-services.git

### Prerequisites

Consul is used for the discovery mechanism.

#### Install Consul

    https://www.consul.io/intro/getting-started/install.html

#### Run Consul

    consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul

### Set up Auth Service

Create a Postgres database for test and development environments:

    CREATE DATABASE auth_service_development;
    CREATE DATABASE auth_service_test;

Use [goose][1] to run the database migrations:

    cd auth_service
    go get bitbucket.org/liamstask/goose/cmd/goose
    goose up

Add a new user to the database with `auth_token=VALID_TOKEN`.

Create a `.env` file with the database url:

    DATABASE_URL=postgres://localhost/auth_service_development?sslmode=disable

### Boot each of the services

Use [foreman][2] to bring up the user service:

    cd www
    foreman start

    cd user_service
    foreman start

Curl the service with a valid auth token:

    $ curl http://localhost:8080 -H "Authorization: Bearer VALID_TOKEN"
    Hello world!

Curl the service with an invalid auth token:

    $ curl http://localhost:8080 -H "Authorization: Bearer INVALID_TOKEN"
    Unauthorized

### Protocol Buffers

When changes are made to the Protocol Buffers we'll need to regenerate them:

    make

[1]: https://bitbucket.org/liamstask/goose
[2]: https://github.com/ddollar/foreman
[3]: https://github.com/golang/protobuf
