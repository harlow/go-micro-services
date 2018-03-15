package main

import (
	"flag"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/frontend"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port       = flag.String("port", "5000", "The server port")
		jaegeraddr = flag.String("jaegeraddr", "jaeger:6831", "Jaeger address")
		consuladdr = flag.String("consuladdr", "consul:8500", "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("frontend", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &frontend.Server{
		Registry: registry,
		Tracer:   tracer,
		Port:     *port,
	}
	srv.Run()
}
