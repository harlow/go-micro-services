# HTTP up front, TCP in the rear.

An experiment with Golang micro-services that accept external HTTP requests and then
leverage [RPC][3] for inter-service tcp communication.

![sequence](https://cloud.githubusercontent.com/assets/739782/6883107/ac49593a-d55b-11e4-8f3e-9c9675db0002.png)

The API Service accepts HTTP requests on port `8000` and then dials a tcp connection
to the User Service. When the user is verified another tcp request is sent to the Like Service.

[gob][3] is the default data encoding format for Golang RPC calls.

### Installation

Clone the repository:

    git clone git@github.com:harlow/go-micro-services.git

Install [goose][1] for managing database migrations:

    go get bitbucket.org/liamstask/goose/cmd/goose

Run the installation script

    ./bin/setup

Add a new user to the development database with `auth_token=VALID_TOKEN`.

A new `.env` file was created in the project root. Make changes if needed:

    API_PORT=8000
    LIKE_SERVICE_PORT=5001
    LIKE_SERVICE_URL=localhost:5001
    USER_SERVICE_DATABASE_URL=postgres://localhost/auth_service_development?sslmode=disable
    USER_SERVICE_PORT=5002
    USER_SERVICE_URL=localhost:5002

### Boot the Services

To make the demo as straigforward we'll use Foreman to boot all the services at once.

Use [foreman][2] to bring up the services:

    foreman start

_Note:_ Typically each application would be run as stand-alone service.

Curl the endpoint with an invalid auth token:

    $ curl http://localhost:8000 -H "Authorization: Bearer INVALID_TOKEN"
    Unauthorized

Curl the API endpoint with a valid auth token:

    $ curl http://localhost:8000 -H "Authorization: Bearer VALID_TOKEN"
    Hello world!

[1]: https://bitbucket.org/liamstask/goose
[2]: https://github.com/ddollar/foreman
[3]: http://golang.org/pkg/net/rpc/
[4]: http://golang.org/pkg/encoding/gob/
