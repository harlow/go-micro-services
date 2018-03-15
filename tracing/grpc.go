package tracing

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// Dialer returns a grpc client connection with tracing interceptor
func Dialer(addr string, tracer opentracing.Tracer) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %v", addr, err)
	}
	return conn, nil
}
