.PHONY: pb data build

pb:
	for f in pb/**/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

lint:
	./bin/lint.sh

build:
	go-bindata -o data/bindata.go -pkg data data/*.json
	./bin/build.sh

run:
	docker-compose build
	docker-compose up
