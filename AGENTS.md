# AGENTS.md

## Purpose
This repository is a Go microservices demo (`github.com/harlow/go-micro-services`) with a gRPC backend and an HTTP frontend.

Use this file as the default operating guide for coding agents working in this repo.

## Repo map
- `cmd/go-micro-services/main.go`: service entrypoint (`frontend`, `search`, `profile`, `geo`, `rate`).
- `internal/services/*`: service implementations.
- `internal/services/*/proto/*.proto`: protobuf sources.
- `internal/services/*/proto/*.pb.go`: generated protobuf Go code.
- `internal/trace/*`: tracing setup (Jaeger/OpenTracing).
- `data/*.json`: source dataset files.
- `data/bindata.go`: generated bindata from `data/*.json`.
- `public/*`: frontend static assets.
- `docker-compose.yml`: multi-service local stack.
- `Makefile`: helper targets for run/proto/data.

## Environment and prerequisites
- Go `1.19` (per `go.mod`).
- Docker + Docker Compose (for full-stack run).
- `protoc` (for regenerating `*.pb.go`).
- `go-bindata` (for regenerating `data/bindata.go`).

Install helper binaries when needed:
- `go install github.com/golang/protobuf/protoc-gen-go@latest`
- `go install github.com/go-bindata/go-bindata/...@latest`

## Validated commands (run before opening a PR)
These were validated in this repo and currently pass:

1. `go test ./...`
2. `go test -race ./...`
3. `go vet ./...`
4. `go build ./...`

Optional coverage snapshot:
- `go test -cover ./...`

## Running the system
- Full stack with Docker:
  - `make run`
  - Frontend: `http://localhost:5001/`
  - Jaeger UI: `http://localhost:16686/search`

- Quick API smoke check after stack is up:
  - `curl "http://localhost:5001/hotels?inDate=2015-04-09&outDate=2015-04-10"`

## Code generation workflows
Regenerate generated artifacts only when relevant sources change.

- Protobuf stubs after editing `internal/services/*/proto/*.proto`:
  - `make proto`

- Bindata after editing `data/*.json`:
  - `make data`

After generation, rerun:
1. `go test ./...`
2. `go build ./...`

## Agent workflow expectations
1. Keep changes scoped to the user request.
2. Prefer small, reviewable commits.
3. Do not manually edit generated files unless intentionally regenerating.
4. If you change service behavior, add or update tests in the relevant package.
5. Before finishing, run the validated checks listed above and report results.

## Known gotchas
- `cmd/go-micro-services/main.go` expects the service name as the first arg (`os.Args[1]`), so running the binary without a subcommand will fail.
- Most packages currently have no tests; at present, `internal/services/profile/profile_test.go` is the primary test file.

## Suggested per-change verification matrix
- Logic-only changes: `go test ./... && go vet ./...`
- Concurrency/networking changes: add `go test -race ./...`
- Entrypoint/wiring/build changes: add `go build ./...`
- Contract/data changes: regenerate artifacts (`make proto` / `make data`) and rerun tests/build.
