.PHONY: pb data lint run

pb:
	for f in pb/**/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

vet:
	./bin/lint.sh

run:
	go-bindata -o data/bindata.go -pkg data data/*.json
	docker-compose build
	docker-compose up
