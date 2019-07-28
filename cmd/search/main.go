package main

import (
	"flag"
	"log"

	"github.com/harlow/go-micro-services/dialer"
	geo "github.com/harlow/go-micro-services/services/geo/proto"
	rate "github.com/harlow/go-micro-services/services/rate/proto"
	"github.com/harlow/go-micro-services/services/search"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port       = flag.Int("port", 8080, "The server port")
		geoAddr    = flag.String("geoaddr", "geo:8080", "Geo server addr")
		rateAddr   = flag.String("rateaddr", "rate:8080", "Rate server addr")
		jaegeraddr = flag.String("jaeger_addr", "jaeger:6831", "Jaeger address")
	)
	flag.Parse()

	tracer, err := tracing.Init("search", *jaegeraddr)
	if err != nil {
		log.Fatalf("tracing init error: %v", err)
	}

	// dial geo srv
	gc, err := dialer.Dial(*geoAddr, dialer.WithTracer(tracer))
	if err != nil {
		log.Fatalf("dialer error: %v", err)
	}

	// dial rate srv
	rc, err := dialer.Dial(*rateAddr, dialer.WithTracer(tracer))
	if err != nil {
		log.Fatalf("dialer error: %v", err)
	}

	srv := search.NewServer(
		geo.NewGeoClient(gc),
		rate.NewRateClient(rc),
		tracer,
	)
	srv.Run(*port)
}
