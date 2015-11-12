package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/protos/auth"
	"github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	port       = flag.Int("port", 8080, "The server port")
	serverName = "service.auth"
)

type authServer struct {
	customers map[string]*auth.Customer
}

// VerifyToken finds a customer by authentication token.
func (s *authServer) VerifyToken(ctx context.Context, args *auth.Args) (*auth.Customer, error) {
	md, _ := metadata.FromContext(ctx)

	t := trace.Tracer{TraceID: md["traceID"]}
	t.In(serverName, args.From)
	defer t.Out(args.From, serverName, time.Now())

	customer := s.customers[args.AuthToken]
	if customer == nil {
		return &auth.Customer{}, errors.New("Invalid Token")
	}

	return customer, nil
}

// loadCustomers loads customers from a JSON file.
func (s *authServer) loadCustomers(file []byte) {
	// unmarshal JSON
	customers := []*auth.Customer{}
	if err := json.Unmarshal(file, &customers); err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}

	// create customer lookup map
	s.customers = make(map[string]*auth.Customer)
	for _, c := range customers {
		s.customers[c.AuthToken] = c
	}
}

func newServer() *authServer {
	s := new(authServer)
	s.loadCustomers(data.MustAsset("data/customers.json"))
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	auth.RegisterAuthServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
