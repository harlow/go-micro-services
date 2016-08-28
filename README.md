# HTTP up front, Protobufs in the rear.

An demonstration of Golang micro-services that accept HTTP/JSON requests at API level and then
leverage [gRPC][1] for inter-service communication.

![new_sequence](https://cloud.githubusercontent.com/assets/739782/7439604/d1f324c2-f036-11e4-958a-6f6913049946.png)

The API Endpoint accepts HTTP requests at `localhost:8080` and then spawns a number of RPC requests to the backend services.

_Note:_ Data for each of the services is stored in JSON flat files under the `/data/` directory. In reality each of the services could choose their own specialty datastore. The Geo service for example could use PostGis or any other database specializing in geospacial queries.

### Setup

Docker is used for bootstrapping the services:

    https://docs.docker.com/engine/installation/

Protobuf v3 are required:

    $ brew install protobuf --devel

Install the protoc-gen libraries:

    $ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

Clone the repository:

    $ git clone git@github.com:harlow/go-micro-services.git

### Protobufs

If changes are made to the Protocol Buffer files use the Makefile to regenerate:

    $ make pb

### Run

To make the demo as straigforward as possible; [Docker Compose](https://docs.docker.com/compose/) is used to run all the services at once (In a production environment each of the services would be run (and scaled) independently).

    $ make build
    $ make run

Curl the endpoint with an invalid auth token:

    $ curl http://localhost:8080 -H "Authorization: Bearer INVALID_TOKEN"
    Unauthorized

Curl the endpoint without checkin or checkout dates:

    $ curl "http://localhost:8080?inDate=2015-04-09" -H "Authorization: Bearer VALID_TOKEN"
    Please specify outDate

Curl the API endpoint with a valid auth token:

    $ curl "http://localhost:8080?inDate=2015-04-09&outDate=2015-04-10" -H "Authorization: Bearer VALID_TOKEN"

The JSON response:

```json
{
    "hotels": [
        {
            "id": 1,
            "name": "Clift Hotel",
            "phoneNumber": "(415) 775-4700",
            "description": "A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
            "address": {
                "streetNumber": "495",
                "streetName": "Geary St",
                "city": "San Francisco",
                "state": "CA",
                "country": "United States",
                "postalCode": "94102"
            }
        }
    ],
    "ratePlans": [
        {
            "hotelId": 1,
            "code": "RACK",
            "inDate": "2015-04-09",
            "outDate": "2015-04-10",
            "roomType": {
                "bookableRate": 109,
                "totalRate": 109,
                "totalRateInclusive": 123.17,
                "code": "KNG"
            }
        }
    ]
}
```

### Tracing Information

<img width="885" alt="apitrace" src="https://cloud.githubusercontent.com/assets/739782/11326691/7ffb5a92-9124-11e5-8818-1d5b3c0b1e51.png">

gRPC tracing information:

    http://micro.demo:8080/debug/requests
    http://micro.demo:8080/debug/events

### Credits

Thanks to all the [contributors][6]. This codebase was heavily inspired by the following talks and repositories:

* [Scaling microservices in Go][3]
* [gRPC Example Service][4]
* [go-kit][5]

[1]: http://www.grpc.io/
[3]: https://speakerdeck.com/mattheath/scaling-microservices-in-go-high-load-strategy-2015
[4]: https://github.com/grpc/grpc-go/tree/master/examples/route_guide
[5]: https://github.com/go-kit/kit
[6]: https://github.com/harlow/go-micro-services/graphs/contributors
