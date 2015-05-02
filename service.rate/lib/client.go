package lib

import (
	"time"

	pb "github.com/harlow/go-micro-services/service.rate/proto"
	trace "github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RatePlanReply struct {
	RatePlans []*pb.RatePlan
	Err       error
}

type Client struct {
	conn   *grpc.ClientConn
	client pb.RateClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, err
	}

	client := pb.NewRateClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c Client) GetRatePlans(ctx context.Context, hotelIDs []int32, inDate string, outDate string) RatePlanReply {
	md, _ := metadata.FromContext(ctx)
	t := trace.Tracer{TraceID: md["traceID"]}
	t.Req(md["from"], "service.rate", "GetRatePlans")
	defer t.Rep("service.rate", md["from"], time.Now())

	args := &pb.Args{
		HotelIds: hotelIDs,
		InDate:   inDate,
		OutDate:  outDate,
	}

	reply, err := c.client.GetRates(ctx, args)

	if err != nil {
		return RatePlanReply{
			RatePlans: []*pb.RatePlan{},
			Err:       err,
		}
	}

	return RatePlanReply{
		RatePlans: reply.RatePlans,
		Err:       nil,
	}
}

func (c Client) Close() error {
	return c.conn.Close()
}
