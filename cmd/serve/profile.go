package main

import (
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/profile"
	"github.com/harlow/go-micro-services/tracing"
)

func runProfile(port int, registry *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("profile", jaegeraddr)
	if err != nil {
		panic(err)
	}

	srv := profile.Server{
		Tracer:   tracer,
		Port:     port,
		Registry: registry,
	}
	return srv.Run()
}
