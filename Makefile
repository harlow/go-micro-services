.PHONY: proto data run down

COMPOSE ?= docker-compose

proto:
	@command -v protoc >/dev/null 2>&1 || { echo "error: protoc not found"; exit 1; }
	@command -v protoc-gen-go >/dev/null 2>&1 || { echo "error: protoc-gen-go not found"; exit 1; }
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || { echo "error: protoc-gen-go-grpc not found"; exit 1; }
	@set -e; for f in internal/services/*/proto/*.proto; do \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative $$f; \
		echo "compiled: $$f"; \
	done

data:
	@command -v go-bindata >/dev/null 2>&1 || { echo "error: go-bindata not found"; exit 1; }
	go-bindata -o data/bindata.go -pkg data data/*.json

run:
	$(COMPOSE) build
	$(COMPOSE) up --remove-orphans

down:
	$(COMPOSE) down --remove-orphans
