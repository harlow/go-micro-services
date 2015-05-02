package lib

import (
	"time"

	pb "github.com/harlow/go-micro-services/service.geo/proto"
	trace "github.com/harlow/go-micro-services/trace"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.GeoClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, err
	}

	client := pb.NewGeoClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c Client) Close() error {
	return c.conn.Close()
}

func (c Client) HotelsWithinBoundedBox(ctx context.Context, latitude int32, longitude int32) ([]int32, error) {
	md, _ := metadata.FromContext(ctx)
	t := trace.Tracer{TraceID: md["traceID"]}
	t.Req(md["from"], "service.geo", "BoundedBox")
	defer t.Rep("service.geo", md["from"], time.Now())

	rect := &pb.Rectangle{
		Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	}

	reply, err := c.client.BoundedBox(ctx, rect)

	if err != nil {
		return []int32{}, err
	}

	return reply.HotelIds, nil
}
