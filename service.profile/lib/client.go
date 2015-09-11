package lib

import (
	"time"

	pb "github.com/harlow/go-micro-services/service.profile/proto"
	trace "github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Hotel pb.Hotel

type Client struct {
	conn   *grpc.ClientConn
	client pb.ProfileClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, err
	}

	client := pb.NewProfileClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c Client) GetHotels(ctx context.Context, hotelIDs []int32) ([]*pb.Hotel, error) {
	md, _ := metadata.FromContext(ctx)
	t := trace.Tracer{TraceID: md["traceID"]}
	t.Req(md["from"], "service.profile", "GetHotels")
	defer t.Rep("service.profile", md["from"], time.Now())

	args := &pb.Args{HotelIds: hotelIDs}
	reply, err := c.client.GetHotels(ctx, args)

	if err != nil {
		return []*pb.Hotel{}, err
	}

	return reply.Hotels, nil
}

func (c Client) Close() error {
	return c.conn.Close()
}
