package lib

import (
	"time"

	pb "github.com/harlow/go-micro-services/service.auth/proto"
	trace "github.com/harlow/go-micro-services/api.trace/client"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.AuthClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c Client) VerifyToken(ctx context.Context, authToken string) error {
	md, _ := metadata.FromContext(ctx)

	trace.Req(md["traceID"], md["from"], "service.auth", "VerifyToken")
	defer trace.Rep(md["traceID"], "service.auth", md["from"], time.Now())

	args := &pb.Args{AuthToken: authToken}
	if _, err := c.client.VerifyToken(ctx, args); err != nil {
		return err
	}

	return nil
}

func (c Client) Close() error {
	return c.conn.Close()
}
