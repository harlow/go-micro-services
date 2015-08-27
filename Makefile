.PHONY: proto data build

proto:
	for f in **/*.proto; do \
		echo compiled: $$f; \
		protoc --go_out=plugins=grpc:. $$f; \
	done

build:
	./build.sh

data:
	go-bindata -o data/bindata.go -pkg data data/*.json
