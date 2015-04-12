# HTTP up front, Protobufs in the rear.

An demonstration of Golang micro-services that accept HTTP/JSON requests at API level and then
leverage [gRPC][1] for inter-service communication.

![flow_sequential](https://cloud.githubusercontent.com/assets/739782/7106819/9cc00ec4-e103-11e4-8718-851b92b913cc.png)

The API Endpoint accepts HTTP requests on port `5000` and then spawns a number of RPC requests to the backend services.

_Note:_ Data for each of the services is stored in JSON flat files under the `/data/` directory. In reality each of the services could choose their own specialty datastore. The Geo service for example could use PostGis or any other database specializing in geospacial queries.

### Installation

Clone the repository:

    git clone git@github.com:harlow/go-micro-services.git

If changes are made to the Protocol Buffers a Make file can be used to regenerate:

    make

### Bootstrap the Services

To make the demo as straigforward as possible; Foreman is used to launch all the services.

Use [foreman][2] to bring up the services:

    foreman start

_Note:_ Typically each service would be run independently.

Curl the endpoint with an invalid auth token:

    $ curl http://localhost:5000 -H "Authorization: Bearer INVALID_TOKEN"
    Unauthorized

Curl the API endpoint with a valid auth token:

    $ curl http://localhost:5000 -H "Authorization: Bearer VALID_TOKEN"
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
        "rates": [
            {
                "hotelId": 1,
                "code": "RACK",
                "inDate": "2015-04-09",
                "outDate": "2015-04-10",
                "roomType": {
                    "bookableRate": 109,
                    "totalRate": 109,
                    "TotalRateInclusive": 123.17,
                    "code": "KNG"
                }
            }
        ]
    }

[1]: http://www.grpc.io/
[2]: https://github.com/ddollar/foreman
