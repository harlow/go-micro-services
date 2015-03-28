package req

import (
  "log"
  "time"
)

func LogReq(from string, to string) {
  log.Printf("[REQ] %v->%v:\n", from, to)
}

func LogRep(from string, to string, start time.Time) {
  elapsed := time.Since(start)
  log.Printf("[REP] %v-->%v: %v\n", from, to, elapsed)
}
