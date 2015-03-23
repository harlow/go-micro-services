# HTTP up front, TCP in the rear.

An experiment with Golang micro-services that accept external HTTP requests and then
leverage [RPC][3] for inter-service tcp communication.

![like_sequence_animated](https://cloud.githubusercontent.com/assets/739782/6776496/81a69638-d0fc-11e4-915a-c0b37108c641.gif)

The API Service accepts HTTP requests on port `8000` and then dials a tcp connection
to the User Service. When the user is verified another tcp request is sent to the Like Service.

[gob][3] is the default data encoding format for Golang RPC calls.

### Installation

Clone the repository:

    git clone git@github.com:harlow/go-micro-services.git

### WWW Service

Create a `.env` file with the port and user service url:

    PORT=8000
    USER_SERVICE_URL=localhost:1234

### User Service

Create a Postgres database for test and development environments:

    CREATE DATABASE auth_service_development;
    CREATE DATABASE auth_service_test;

Use [goose][1] to manage database migrations:

    go get bitbucket.org/liamstask/goose/cmd/goose

Run the migrations:

    cd user_srevice
    goose up

Create a `.env` file with the database url:

    DATABASE_URL=postgres://localhost/auth_service_development?sslmode=disable
    PORT=1234

Add a new user to the database with `auth_token=VALID_TOKEN`.

### Run the Services

Use [foreman][2] to bring up the www and user service:

    cd www
    foreman start

    cd user_service
    foreman start

Curl the WWW service with a valid auth token:

    $ curl http://localhost:8080 -H "Authorization: Bearer VALID_TOKEN"
    Hello world!

Curl the service with an invalid auth token:

    $ curl http://localhost:8080 -H "Authorization: Bearer INVALID_TOKEN"
    Unauthorized

[1]: https://bitbucket.org/liamstask/goose
[2]: https://github.com/ddollar/foreman
[3]: http://golang.org/pkg/net/rpc/
[4]: http://golang.org/pkg/encoding/gob/
