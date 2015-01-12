# HTTP in the front, Protobufs in the rear.

An example of Golang micro-services that accept HTTP requests and then
leverage [Protocol Buffers][3] for inter-service tcp communications.

![auth_service_flow](https://cloud.githubusercontent.com/assets/739782/5699710/2ffb37e4-99e3-11e4-9fec-4c0dd52a98c3.png)

Service1 accepts HTTP requests on port `8080` and then dials a tcp connection
on `127.0.0.1:1984` to authenticate the token with the AuthService.

### Installation

Clone the repository:

    git clone git@github.com:harlow/go-micro-services.git

Create a `.env` file with a database URL for the Auth Application:

    DATABASE_URL=postgres://localhost/auth_service_development?sslmode=disable

Use [goose][1] to run the database migrations:

    goose up

Add a user to the database with an `auth_token`.

### Run the Application

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
