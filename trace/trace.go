package trace

import (
	"log"
	"time"

	"github.com/nu7hatch/gouuid"
)

func NewTracer() Tracer {
	traceID, _ := uuid.NewV4()
	return Tracer{
		TraceID: traceID.String(),
	}
}

type Tracer struct {
	TraceID string
}

func (t Tracer) Req(from string, to string, call string) {
	log.Printf("[REQ] %s %v->%v: %v\n", t.TraceID, from, to, call)
}

func (t Tracer) Rep(from string, to string, startTime time.Time) {
	elapsed := time.Since(startTime)
	log.Printf("[REP] %s %v-->%v: %v\n", t.TraceID, from, to, elapsed)
}

func (t Tracer) In(from string, to string) {
	log.Printf("[IN]  %s %v->%v:\n", t.TraceID, from, to)
}

func (t Tracer) Out(from string, to string, startTime time.Time) {
	elapsed := time.Since(startTime)
	log.Printf("[OUT] %s %v-->%v: %v\n", t.TraceID, from, to, elapsed)
}
