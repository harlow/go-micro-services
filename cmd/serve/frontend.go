package main

import (
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/frontend"
	"github.com/harlow/go-micro-services/tracing"
)

func runFrontend(port int, registry *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("frontend", jaegeraddr)
	if err != nil {
		panic(err)
	}

	srv := &frontend.Server{
		Registry: registry,
		Tracer:   tracer,
		Port:     port,
	}
	return srv.Run()
}
