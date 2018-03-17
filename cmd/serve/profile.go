package main

import (
	"fmt"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/profile"
	"github.com/harlow/go-micro-services/tracing"
)

const profileSrvName = "srv-profile"

func runProfile(port int, consul *registry.Client, jaegeraddr string) error {
	tracer, err := tracing.Init("profile", jaegeraddr)
	if err != nil {
		return fmt.Errorf("tracing init error: %v", err)
	}

	// service registry
	id, err := consul.Register(profileSrvName, port)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	defer consul.Deregister(id)

	srv := profile.NewServer(tracer)
	return srv.Run(port)
}
