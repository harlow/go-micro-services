package main

import (
	"fmt"
	"log"

	"cloud.google.com/go/trace"
	"github.com/harlow/go-micro-services/lib"
	"github.com/harlow/go-micro-services/pb/geo"
	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/rate"
	"github.com/harlow/grpc-google-cloud-trace/intercept"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type config struct {
	TraceProjectID  string `envconfig:"TRACE_PROJECT_ID"`
	TraceJSONConfig string `envconfig:"TRACE_JSON_CONFIG"`
	Port            string `default:"8080" envconfig:"PORT"`
	GeoAddr         string `default:"geo:8080" envconfig:"GEO_ADDR"`
	ProfileAddr     string `default:"profile:8080" envconfig:"PROFILE_ADDR"`
	RateAddr        string `default:"rate:8080" envconfig:"RATE_ADDR"`
}

func newEnv() *env {
	var cfg config
	envconfig.MustProcess("", &cfg)

	tc := lib.NewTraceClient(
		cfg.TraceProjectID, cfg.TraceJSONConfig,
	)

	return &env{
		cfg:           cfg,
		Tracer:        tc,
		GeoClient:     geo.NewGeoClient(mustDial(cfg.GeoAddr, tc)),
		ProfileClient: profile.NewProfileClient(mustDial(cfg.ProfileAddr, tc)),
		RateClient:    rate.NewRateClient(mustDial(cfg.RateAddr, tc)),
	}
}

type env struct {
	cfg config

	Tracer        *trace.Client
	GeoClient     geo.GeoClient
	ProfileClient profile.ProfileClient
	RateClient    rate.RateClient
}

func (e *env) serviceAddr() string {
	return fmt.Sprintf(":%s", e.cfg.Port)
}

// mustDial ensures a tcp connection to specified address.
func mustDial(addr string, traceClient *trace.Client) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(intercept.ClientTrace(traceClient)),
	)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		panic(err)
	}
	return conn
}
