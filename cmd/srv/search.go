package main

import (
	"fmt"

	"github.com/harlow/go-micro-services/dialer"
	"github.com/harlow/go-micro-services/registry"
	geo "github.com/harlow/go-micro-services/services/geo/proto"
	rate "github.com/harlow/go-micro-services/services/rate/proto"
	"github.com/harlow/go-micro-services/services/search"
	"github.com/harlow/go-micro-services/tracing"
)

const searchSrvName = "srv-search"

func runSearch(port int, consul *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("search", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// dial geo srv
	gc, err := dialer.Dial(
		geoSrvName,
		dialer.WithTracer(tracer),
		dialer.WithBalancer(consul.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}

	// dial rate srv
	rc, err := dialer.Dial(
		rateSrvName,
		dialer.WithTracer(tracer),
		dialer.WithBalancer(consul.Client),
	)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}

	// service registry
	id, err := consul.Register(searchSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer consul.Deregister(id)

	srv := search.NewServer(
		geo.NewGeoClient(gc),
		rate.NewRateClient(rc),
		tracer,
	)
	return srv.Run(port)
}
