package main

import (
	"fmt"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/rate"
	"github.com/harlow/go-micro-services/tracing"
)

const rateSrvName = "srv-rate"

func runRate(port int, consul *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("rate", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// service registry
	id, err := consul.Register(rateSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer consul.Deregister(id)

	srv := rate.NewServer(tracer)
	return srv.Run(port)
}
