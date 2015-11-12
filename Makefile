.PHONY: proto data build

proto:
	for f in protos/**/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

build:
	./build.sh

data:
	go-bindata -o data/bindata.go -pkg data data/*.json
