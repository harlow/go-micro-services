package main

import (
	"fmt"
	"log"

	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/search"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type config struct {
	Port        string `default:"8080" envconfig:"PORT"`
	SearchAddr  string `default:"search:8080" envconfig:"SEARCH_ADDR"`
	ProfileAddr string `default:"profile:8080" envconfig:"PROFILE_ADDR"`
}

func newEnv() *env {
	var cfg config
	envconfig.MustProcess("", &cfg)

	return &env{
		cfg:           cfg,
		SearchClient:  search.NewSearchClient(mustDial(cfg.SearchAddr)),
		ProfileClient: profile.NewProfileClient(mustDial(cfg.ProfileAddr)),
	}
}

type env struct {
	cfg config

	SearchClient  search.SearchClient
	ProfileClient profile.ProfileClient
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
