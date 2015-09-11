package lib

import (
	"time"

	auth "github.com/harlow/go-micro-services/service.auth/proto"
	trace "github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn   *grpc.ClientConn
	client auth.AuthClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, err
	}

	client := auth.NewAuthClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c Client) Close() error {
	return c.conn.Close()
}

func (c Client) VerifyToken(ctx context.Context, authToken string) error {
	md, _ := metadata.FromContext(ctx)

	t := trace.Tracer{TraceID: md["traceID"]}
	t.Req(md["from"], "service.auth", "VerifyToken")
	defer t.Rep("service.auth", md["from"], time.Now())

	args := &auth.Args{AuthToken: authToken}
	if _, err := c.client.VerifyToken(ctx, args); err != nil {
		return err
	}

	return nil
}
