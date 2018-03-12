package main

import (
	"flag"

	geo "github.com/harlow/go-micro-services/services/geo/proto"
	rate "github.com/harlow/go-micro-services/services/rate/proto"
	"github.com/harlow/go-micro-services/services/search"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port       = flag.String("port", "8080", "The server port")
		geoAddr    = flag.String("geoaddr", "geo:8080", "Geo server addr")
		rateAddr   = flag.String("rateaddr", "rate:8080", "Rate server addr")
		jaegerAddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger server addr")
	)
	flag.Parse()

	var (
		tracer     = tracing.Init("search", *jaegerAddr)
		geoClient  = geo.NewGeoClient(tracing.MustDial(*geoAddr, tracer))
		rateClient = rate.NewRateClient(tracing.MustDial(*rateAddr, tracer))
	)

	srv := &search.Server{
		Tracer:     tracer,
		Port:       *port,
		GeoClient:  geoClient,
		RateClient: rateClient,
	}
	srv.Run()
}
