.PHONY: pb data build

pb:
	for f in pb/**/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

lint:
	./bin/lint.sh

build:
	./bin/build.sh

data:
	go-bindata -o data/bindata.go -pkg data data/*.json

run:
	docker-compose build
	docker-compose up
