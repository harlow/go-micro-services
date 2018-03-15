# Golang Microservices Example

An demonstration of Golang micro-services that expose a HTTP/JSON frontend and then
leverages [gRPC][1] for inter-service communication.

* Services written in Golang
* Consul for service discovery
* gRPC for service communication
* Jaeger for request tracing

The example application plots Hotel locations on a Google map:

<img width="865" alt="screen shot 2016-11-07 at 9 31 12 pm" src="https://cloud.githubusercontent.com/assets/739782/20087958/de0ef9b4-a531-11e6-953a-4425fe445883.png">

The web page makes an HTTP request to the API Endpoint which in turn spawns a number of RPC requests to the backend services.
 
![reqseq](https://user-images.githubusercontent.com/739782/29102656-8328b66e-7c6f-11e7-8fe5-c9d5f90e213f.png)

_Note:_ Data for each of the services is stored in JSON flat files under the `/data/` directory. In reality each of the services could choose their own specialty datastore. The Geo service for example could use PostGis or any other database specializing in geospacial queries.

### Setup

Docker is required for running the services https://docs.docker.com/engine/installation.

Protobuf v3 are required:

    $ brew install protobuf

Install the protoc-gen libraries:

    $ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

Clone the repository:

    $ git clone git@github.com:harlow/go-micro-services.git

### Run

To make the demo as straigforward as possible; [Docker Compose](https://docs.docker.com/compose/) is used to run all the services at once (In a production environment each of the services would be run (and scaled) independently).

    $ make run

Vist the web page in a browser:

[http://localhost:5000/](http://localhost:5000/)

cURL the API endpoint and receive GeoJSON response:

    $ curl "http://localhost:5000/hotels?inDate=2015-04-09&outDate=2015-04-10" 

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

### Service Discovery

[Consul](https://www.consul.io/) is used for service discovery. This allows each service to register with the registry and then discovery the IP addresses of the services they need to comunicate with.

<img width="1072" alt="consul service discovery" src="https://user-images.githubusercontent.com/739782/37442561-23444504-285b-11e8-9d10-1c971c44a720.png">

### Tracing

The [Jaeger Tracing](https://github.com/jaegertracing/jaeger) project is used for tracing inter-service requests.

<img width="1068" alt="jaeger trace diagram" src="https://user-images.githubusercontent.com/739782/37238917-b710b734-2484-11e8-8148-50fc7fe5e366.png">

Once docker-compose has launched all services visit http://localhost:16686/search

### Protobufs

If changes are made to the Protocol Buffer files use the Makefile to regenerate:

    $ make proto

### Credits

Thanks to all the [contributors][6]. This codebase was heavily inspired by the following talks and repositories:

* [Scaling microservices in Go][3]
* [gRPC Example Service][4]
* [go-kit][5]

[1]: http://www.grpc.io/
[2]: https://github.com/docker/compose/issues/3560
[3]: https://speakerdeck.com/mattheath/scaling-microservices-in-go-high-load-strategy-2015
[4]: https://github.com/grpc/grpc-go/tree/master/examples/route_guide
[5]: https://github.com/go-kit/kit
[6]: https://github.com/harlow/go-micro-services/graphs/contributors
