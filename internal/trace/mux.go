package trace

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewServeMux creates a new TracedServeMux.
func NewServeMux() *TracedServeMux {
	return &TracedServeMux{mux: http.NewServeMux()}
}

// TracedServeMux is a wrapper around http.ServeMux that instruments handlers for tracing.
type TracedServeMux struct {
	mux *http.ServeMux
}

// Handle implements http.ServeMux#Handle.
func (tm *TracedServeMux) Handle(pattern string, handler http.Handler) {
	tm.mux.Handle(pattern, otelhttp.NewHandler(handler, "HTTP "+pattern))
}

// ServeHTTP implements http.ServeMux#ServeHTTP.
func (tm *TracedServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm.mux.ServeHTTP(w, r)
}
