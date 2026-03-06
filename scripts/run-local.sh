#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$ROOT_DIR/.tmp/local"
mkdir -p "$LOG_DIR"

PIDS=()

cleanup() {
  for pid in "${PIDS[@]}"; do
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
    fi
  done
}
trap cleanup EXIT INT TERM

start_service() {
  local name="$1"
  shift
  local logfile="$LOG_DIR/$name.log"
  echo "starting $name (log: $logfile)"
  (cd "$ROOT_DIR" && go run ./cmd/go-micro-services "$@" >"$logfile" 2>&1) &
  PIDS+=("$!")
}

start_service geo -port=8081 -jaeger=localhost:4317 geo
start_service rate -port=8082 -jaeger=localhost:4317 rate
start_service profile -port=8083 -jaeger=localhost:4317 profile
start_service search -port=8084 -geoaddr=localhost:8081 -rateaddr=localhost:8082 -jaeger=localhost:4317 search
start_service frontend -port=5001 -searchaddr=localhost:8084 -profileaddr=localhost:8083 -jaeger=localhost:4317 frontend

echo
echo "local stack is starting:"
echo "- frontend: http://localhost:5001/"
echo "- frontend ready endpoint: http://localhost:5001/readyz"
echo
echo "press Ctrl+C to stop all services"

while true; do
  for pid in "${PIDS[@]}"; do
    if ! kill -0 "$pid" >/dev/null 2>&1; then
      wait "$pid" >/dev/null 2>&1 || true
      echo "one or more services exited unexpectedly; check logs in $LOG_DIR"
      exit 1
    fi
  done
  sleep 1
done
