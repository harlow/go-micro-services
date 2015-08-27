.PHONY: proto data build

proto:
	for f in **/*.proto; do \
		echo compiled: $$f; \
		protoc --go_out=plugins=grpc:. $$f; \
	done

build:
	p=`pwd`
	for d in cmd/*; do
		if [[ -d $d ]]; then
			cd $p/$d;
			env GOOS=linux GOARCH=386 go build;
		fi
	done
	cd $p

data:
	go-bindata -o data/bindata.go -pkg data data/*.json
