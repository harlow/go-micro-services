package main

import (
	"flag"
	"log"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/profile"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port       = flag.Int("port", 8080, "The server port")
		jaegeraddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger server addr")
		consuladdr = flag.String("consuladdr", "consul:8500", "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("profile", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := profile.Server{
		Tracer:   tracer,
		Port:     *port,
		Registry: registry,
	}
	log.Fatal(srv.Run())
}
