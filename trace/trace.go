package trace

import (
  "log"
  "time"

  "github.com/nu7hatch/gouuid"
)

func NewTracer(name string) Tracer {
  traceID, _ := uuid.NewV4()
  return Tracer{
    Name: name,
    TraceID: traceID.String(),
  }
}

type Tracer struct {
  TraceID string
  Name string
}

func (t Tracer) Request(to string) {
  log.Printf("%s %v->%v:\n", t.TraceID, t.Name, to)
}

func (t Tracer) Reply(from string, startTime time.Time) {
  elapsed := time.Since(startTime)
  log.Printf("%s %v-->%v: %v\n", t.TraceID, from, t.Name, elapsed)
}

func (t Tracer) In() {
  log.Printf("%s %v\n", t.TraceID, t.Name)
}

func (t Tracer) Out(startTime time.Time) {
  elapsed := time.Since(startTime)
  log.Printf("%s %v: %v\n", t.TraceID, t.Name, elapsed)
}
