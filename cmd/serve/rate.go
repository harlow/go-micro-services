package main

import (
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/rate"
	"github.com/harlow/go-micro-services/tracing"
)

func runRate(port int, registry *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("rate", jaegeraddr)
	if err != nil {
		panic(err)
	}

	srv := &rate.Server{
		Tracer:   tracer,
		Port:     port,
		Registry: registry,
	}
	return srv.Run()
}
