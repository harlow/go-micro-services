.PHONY: default data docker

default:
	for f in **/*.proto; do \
		echo compiled: $$f; \
		protoc --go_out=plugins=grpc:. $$f; \
	done

docker:
	pwd=`pwd`
	for d in cmd; do
		cd $pwd/cmd/$d;
		env GOOS=linux GOARCH=386 go build; docker build -t $d .;
	done

data:
	for d in cmd; do
		cp -r data cmd/$d;
	done

