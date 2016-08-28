package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/harlow/go-micro-services/data"
	"github.com/harlow/go-micro-services/pb/auth"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authServer struct {
	customers map[string]*auth.Customer
}

// VerifyToken returns a customer from authentication token.
func (s *authServer) VerifyToken(ctx context.Context, req *auth.Request) (*auth.Result, error) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
		tr.LazyPrintf("traceID %s", traceID)
	}

	customer := s.customers[req.AuthToken]
	if customer == nil {
		return &auth.Result{}, errors.New("Invalid Token")
	}

	reply := new(auth.Result)
	reply.Customer = customer
	return reply, nil
}

// loadCustomers loads customers from a JSON file.
func loadCustomerData(path string) map[string]*auth.Customer {
	file := data.MustAsset(path)
	customers := []*auth.Customer{}

	// unmarshal JSON
	if err := json.Unmarshal(file, &customers); err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}

	// create customer lookup map
	cache := make(map[string]*auth.Customer)
	for _, c := range customers {
		cache[c.AuthToken] = c
	}
	return cache
}

func main() {
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	// listen on port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// grpc server
	srv := grpc.NewServer()
	auth.RegisterAuthServer(srv, &authServer{
		customers: loadCustomerData("data/customers.json"),
	})
	srv.Serve(lis)
}
