package req

import (
  "log"
  "time"
)

func LogReq(from string, to string) {
  log.Printf("[REQ] %v → %v\n", from, to)
}

func LogRep(from string, to string, start time.Time) {
  elapsed := time.Since(start)
  log.Printf("[REP] %v ← %v - %v\n", from, to, elapsed)
}

func LogIn(from string, to string) {
  log.Printf("[IN]  %v → %v\n", from, to)
}

func LogOut(from string, to string, start time.Time) {
  elapsed := time.Since(start)
  log.Printf("[OUT] %v ← %v - %v\n", from, to, elapsed)
}
