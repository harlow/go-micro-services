package frontend

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

// Server implements frontend service
type Server struct {
	Port   string
	Tracer opentracing.Tracer
}

// Run the server
func (s *Server) Run() error {
	fs := http.FileServer(http.Dir("services/frontend/static"))
	http.Handle("/", fs)
	return http.ListenAndServe(":"+s.Port, nil)
}
