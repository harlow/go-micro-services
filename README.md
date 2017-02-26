# Golang Microservices Example

An demonstration of Golang micro-services that accept HTTP/JSON requests at API level and then
leverage [gRPC][1] for inter-service communication. The example application plots Hotel locations on a Google map:

<img width="865" alt="screen shot 2016-11-07 at 9 31 12 pm" src="https://cloud.githubusercontent.com/assets/739782/20087958/de0ef9b4-a531-11e6-953a-4425fe445883.png">

The web page makes an HTTP request to the API Endpoint which in turn spawns a number of RPC requests to the backend services.
 
![new_sequence](https://cloud.githubusercontent.com/assets/739782/7439604/d1f324c2-f036-11e4-958a-6f6913049946.png)

_Note:_ Data for each of the services is stored in JSON flat files under the `/data/` directory. In reality each of the services could choose their own specialty datastore. The Geo service for example could use PostGis or any other database specializing in geospacial queries.

### Setup

Docker is required for running the services https://docs.docker.com/engine/installation.

Protobuf v3 are required:

    $ brew install protobuf

Install the protoc-gen libraries:

    $ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

Clone the repository:

    $ git clone git@github.com:harlow/go-micro-services.git

### Protobufs

If changes are made to the Protocol Buffer files use the Makefile to regenerate:

    $ make pb

### Run

To make the demo as straigforward as possible; [Docker Compose](https://docs.docker.com/compose/) is used to run all the services at once (In a production environment each of the services would be run (and scaled) independently).

    $ make run

Vist the web page in a browser:

[http://localhost:5000/](http://localhost:5000/)

cURL the API endpoint and receive GeoJSON response:

    $ curl "http://localhost:8080/inventory?inDate=2015-04-09&outDate=2015-04-10" 

The JSON response:

```json
{
	"type": "FeatureCollection",
	"features": [{
		"id": "5",
		"type": "Feature",
		"properties": {
			"name": "Phoenix Hotel",
			"phone_number": "(415) 776-1380"
		},
		"geometry": {
			"type": "Point",
			"coordinates": [-122.4181, 37.7831]
		}
	}, {
		"id": "3",
		"type": "Feature",
		"properties": {
			"name": "Hotel Zetta",
			"phone_number": "(415) 543-8555"
		},
		"geometry": {
			"type": "Point",
			"coordinates": [-122.4071, 37.7834]
		}
	}]
}
```

### Tracing

Tracing data is pushed to Google Stackdriver. See [Tracing gRPC calls in Golang with Google Stackdriver](https://medium.com/@harlow/tracing-grpc-calls-in-golang-with-google-stackdriver-b22495763a06#.6skzmkwqy) for more detailed information on how tracing was implemented.

<img width="987" alt="screen shot 2017-02-25 at 11 42 02 pm" src="https://cloud.githubusercontent.com/assets/739782/23338017/49e00402-fbb4-11e6-9d89-7b4d7af54315.png">

To enable tracing add the following to `.env`:

```
TRACE_PROJECT_ID=
TRACE_JSON_CONFIG=
```

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
