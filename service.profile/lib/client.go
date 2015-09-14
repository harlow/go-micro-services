package lib

import (
	"time"

	pb "github.com/harlow/go-micro-services/service.profile/proto"
	trace "github.com/harlow/go-micro-services/api.trace/client"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Hotel pb.Hotel

type ProfileReply struct {
	Hotels []*pb.Hotel
	Err       error
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

type Client struct {
	conn   *grpc.ClientConn
	client pb.ProfileClient
}

func (c Client) GetHotels(ctx context.Context, hotelIDs []int32) ProfileReply {
	md, _ := metadata.FromContext(ctx)

	trace.Req(md["traceID"], md["from"], "service.profile", "GetHotels")
	defer trace.Rep(md["traceID"], "service.profile", md["from"], time.Now())

	args := &pb.Args{HotelIds: hotelIDs}
	reply, err := c.client.GetHotels(ctx, args)

	if err != nil {
		return ProfileReply{
			Hotels:  []*pb.Hotel{},
			Err: err,
		}
	}

	return ProfileReply{
		Hotels: reply.Hotels,
		Err:       nil,
	}
}

func (c Client) Close() error {
	return c.conn.Close()
}
