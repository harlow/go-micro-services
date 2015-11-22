.PHONY: proto data build

proto:
	for f in proto/**/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

lint:
	./lint.sh

build:
	./build.sh

data:
	go-bindata -o data/bindata.go -pkg data data/*.json

up:
	docker-compose build
	docker-compose up
