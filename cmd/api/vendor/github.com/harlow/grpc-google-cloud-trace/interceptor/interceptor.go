package interceptor

import (
	"fmt"
	"strings"

	"cloud.google.com/go/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	headerKey  = "google_cloud_trace_context"
	headerTmpl = "%s/%d;o=1"
)

// Server returns a new unary server interceptor used for parsing
// the google trace header from metadata and creating new tracing span.
func Server(traceClient *trace.Client) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		header := strings.Join(md[headerKey], "")
		span := traceClient.SpanFromHeader(info.FullMethod, header)
		defer span.Finish()

		ctx = trace.NewContext(ctx, span)
		return handler(ctx, req)
	}
}

// Client returns a new unary client interceptor used for injecting
// the google trace header into the context metadata.
func Client(traceClient *trace.Client) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		span := trace.FromContext(ctx).NewChild(
			fmt.Sprintf("/grpc.Sent%s", method),
		)
		defer span.Finish()

		md, ok := metadata.FromContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		md[headerKey] = append(md[headerKey], traceHeader(span.TraceID()))

		ctx = metadata.NewContext(ctx, md)
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

// force google trace header specification
// https://cloud.google.com/trace/docs/faq
func traceHeader(traceID string) string {
	return fmt.Sprintf(headerTmpl, traceID, 0)
}
