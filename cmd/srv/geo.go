package main

import (
	"fmt"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/geo"
	"github.com/harlow/go-micro-services/tracing"
)

const geoSrvName = "srv-geo"

func runGeo(port int, consul *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("geo", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// service registry
	id, err := consul.Register(geoSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer consul.Deregister(id)

	srv := geo.NewServer(tracer)
	return srv.Run(port)
}
