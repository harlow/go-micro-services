package lib

import (
	auth "github.com/harlow/go-micro-services/service.auth/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

func (c Client) VerifyToken(traceID string, serverName string, authToken string) error {
	// set up args and client
	args := &auth.Args{
		TraceId:   traceID,
		From:      serverName,
		AuthToken: authToken,
	}

	// verify auth token
	if _, err := c.client.VerifyToken(context.Background(), args); err != nil {
		return err
	}

	return nil
}
