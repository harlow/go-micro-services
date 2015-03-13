all:
	protoc --go_out=. user_service/proto/user.proto
	protoc --go_out=. like_service/proto/like.proto
