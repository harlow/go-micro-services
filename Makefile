all:
	protoc --go_out=. user_service/proto/user.proto
