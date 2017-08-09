package main

import (
	"fmt"
	"log"

	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/rate"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type config struct {
	Port        string `default:"8080" envconfig:"PORT"`
	GeoAddr     string `default:"geo:8080" envconfig:"GEO_ADDR"`
	ProfileAddr string `default:"profile:8080" envconfig:"PROFILE_ADDR"`
	RateAddr    string `default:"rate:8080" envconfig:"RATE_ADDR"`
}

func newEnv() *env {
	var cfg config
	envconfig.MustProcess("", &cfg)

	return &env{
		cfg:           cfg,
		GeoClient:     geo.NewGeoClient(mustDial(cfg.GeoAddr)),
		ProfileClient: profile.NewProfileClient(mustDial(cfg.ProfileAddr)),
		RateClient:    rate.NewRateClient(mustDial(cfg.RateAddr)),
	}
}

type env struct {
	cfg config

	GeoClient     geo.GeoClient
	ProfileClient profile.ProfileClient
	RateClient    rate.RateClient
}

func (e *env) serviceAddr() string {
	return fmt.Sprintf(":%s", e.cfg.Port)
}

// mustDial ensures a tcp connection to specified address.
func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		panic(err)
	}
	return conn
}
