package main

import (
	"flag"
	"log"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/geo"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port       = flag.Int("port", 8080, "Server port")
		jaegeraddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger address")
		consuladdr = flag.String("consuladdr", "consul:8500", "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("geo", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &geo.Server{
		Port:     *port,
		Tracer:   tracer,
		Registry: registry,
	}
	log.Fatal(srv.Run())
}
