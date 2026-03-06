.PHONY: proto data run down

COMPOSE ?= docker-compose

proto:
	@command -v protoc >/dev/null 2>&1 || { echo "error: protoc not found"; exit 1; }
	@set -e; for f in internal/services/*/proto/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
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
