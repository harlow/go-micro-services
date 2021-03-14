# Go-bindata Usage

## Installation

### Setup

Install the go-bindata libraries:

	$ go get -u github.com/go-bindata/go-bindata/...


### Change data

1. inventory.json
Remove the "hotel2" section

2. Gen bindata.go
```bash
$ cd ~/go/src/github.com/xiaobin80/go-micro-services/
$ go-bindata -pkg=data data/
$ mv bindata.go ./data/
```


### Run
    $ make run

cURL the API endpoint and receive GeoJSON response:

    $ curl "http://localhost:5000/hotels?inDate=2015-04-09&outDate=2015-04-10" 

The JSON response:

```json
{
	"features":[{
		"geometry":{
			"coordinates":[-122.4071,37.7834],"type":"Point"
		},
		"id":"3",
		"properties":{
			"name":"Hotel Zetta",
			"phone_number":"(415) 543-8555"
		},
		"type":"Feature"
		},{
		"geometry":{
			"coordinates":[-122.4112,37.7867],
			"type":"Point"
		},
		"id":"1",
		"properties":{
			"name":"Clift Hotel",
			"phone_number":"(415) 775-4700"},
			"type":"Feature"}
		],
	"type":"FeatureCollection"
}
```

## ISSUES
1. add json data
https://github.com/go-bindata/go-bindata/issues/34
