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
	"github.com/harlow/go-micro-services/proto/auth"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// newServer returns a server with initialization data loaded.
func newServer() *authServer {
	s := new(authServer)
	s.loadCustomers(data.MustAsset("data/customers.json"))
	return s
}

func logRequest(ctx context.Context) {
	md, _ := metadata.FromContext(ctx)
	traceID := strings.Join(md["traceID"], ",")

	if tr, ok := trace.FromContext(ctx); ok {
		tr.LazyPrintf("traceID %s", traceID)
	}

	log.Printf("traceID %s", traceID)
}

type authServer struct {
	customers map[string]*auth.Customer
}

// VerifyToken finds a customer by authentication token.
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

func main() {
	var port = flag.Int("port", 8080, "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	g := grpc.NewServer()
	auth.RegisterAuthServer(g, newServer())
	g.Serve(lis)
}
