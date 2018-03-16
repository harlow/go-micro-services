package main

import (
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/geo"
	"github.com/harlow/go-micro-services/tracing"
)

func runGeo(port int, registry *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("geo", jaegeraddr)
	if err != nil {
		panic(err)
	}

	srv := &geo.Server{
		Port:     port,
		Tracer:   tracer,
		Registry: registry,
	}
	return srv.Run()
}
