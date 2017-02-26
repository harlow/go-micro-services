package lib

import (
	"context"
	"log"

	"cloud.google.com/go/trace"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// NewTraceClient returns an initialzed client
func NewTraceClient(projectID, jsonConfig string) *trace.Client {
	ctx := context.Background()
	tokenSource := getTokenSource(ctx, jsonConfig)
	return getTraceClient(ctx, projectID, tokenSource)
}

// getTraceClient returns an initialized tracing client
func getTraceClient(ctx context.Context, projectID string, tokenSource oauth2.TokenSource) *trace.Client {
	traceClient, err := trace.NewClient(
		ctx,
		projectID,
		option.WithTokenSource(tokenSource),
	)
	if err != nil {
		log.Fatalf("cannot init noop tracer: %v", err)
	}

	p, err := trace.NewLimitedSampler(1, 99) // sample every request.
	if err != nil {
		log.Fatalf("NewLimitedSampler: %v", err)
	}
	traceClient.SetSamplingPolicy(p)

	return traceClient
}

// getTokenSource creates token source from json config
// https://developers.google.com/identity/protocols/googlescopes#cloudtracev1
func getTokenSource(ctx context.Context, jsonConfig string) oauth2.TokenSource {
	cfg, err := google.JWTConfigFromJSON(
		[]byte(jsonConfig),
		"https://www.googleapis.com/auth/trace.append",
	)
	if err != nil {
		log.Fatalf("google config from json error: %v", err)
	}
	return cfg.TokenSource(ctx)
}
