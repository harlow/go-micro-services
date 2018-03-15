package tracing

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	consul "github.com/hashicorp/consul/api"
	lb "github.com/olivere/grpc/lb/consul"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// Dialer returns a load balanced grpc client conn with tracing interceptor
func Dialer(name string, tracer opentracing.Tracer, registry *consul.Client) (*grpc.ClientConn, error) {
	r, err := lb.NewResolver(registry, name, "")
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		name,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
		grpc.WithBalancer(grpc.RoundRobin(r)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %v", name, err)
	}
	return conn, nil
}
