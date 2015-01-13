# HTTP up front, Protobufs in the rear.

An experiment with Golang micro-services that accept external HTTP requests and then
leverage [Protocol Buffers][3] for inter-service tcp communication.

![auth_service](https://cloud.githubusercontent.com/assets/739782/5716158/ecc8b7ee-9a9b-11e4-8821-e48d5fdc5838.png)

The Web Service accepts HTTP requests on port `8000` and then dials a tcp connection
to port `1984` and authenticates the token with the Auth Service.

### Installation

Clone the repository:

    git clone git@github.com:harlow/go-micro-services.git

Create a `.env` file with a database and service details:

    AUTH_SERVICE_PORT=1984
    DATABASE_URL=postgres://localhost/auth_service_development?sslmode=disable
    WEB_SERVICE_PORT=8000

Use [goose][1] to run the database migrations:

    goose up

### Run the Application

Add a new user to the database with `auth_token=VALID_TOKEN`.

Use [foreman][2] to bring up the services:

    foreman start

Curl the service with a valid auth token:

    $ curl http://localhost:8000 -H "Authorization: Bearer VALID_TOKEN"
    Hello world!

Curl the service with an invalid auth token:

    $ curl http://localhost:8000 -H "Authorization: Bearer INVALID_TOKEN"
    Unauthorized

### Protocol Buffers

When changes are made to the Protocol Buffers we'll need to regenerate them:

    make

[1]: https://bitbucket.org/liamstask/goose
[2]: https://github.com/ddollar/foreman
[3]: https://github.com/golang/protobuf
