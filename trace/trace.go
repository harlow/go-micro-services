package trace

import (
  "log"
  "time"
)

func Request(traceID string, from string, to string) {
  log.Printf("%s %v->%v:\n", traceID, from, to)
}

func Reply(traceID string, from string, to string, start time.Time) {
  elapsed := time.Since(start)
  log.Printf("%s %v-->%v: %v\n", traceID, from, to, elapsed)
}
