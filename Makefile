all:
		protoc --go_out=. protobufs/user/user.proto
