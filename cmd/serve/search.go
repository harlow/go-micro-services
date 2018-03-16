package main

import (
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/search"
	"github.com/harlow/go-micro-services/tracing"
)

func runSearch(port int, registry *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("search", jaegeraddr)
	if err != nil {
		panic(err)
	}

	srv := &search.Server{
		Tracer:   tracer,
		Port:     port,
		Registry: registry,
	}
	return srv.Run()
}
