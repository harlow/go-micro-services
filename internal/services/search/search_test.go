package search

import (
	"testing"

	geo "github.com/harlow/go-micro-services/internal/services/geo/proto"
	rate "github.com/harlow/go-micro-services/internal/services/rate/proto"
	searchpb "github.com/harlow/go-micro-services/internal/services/search/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type geoClientStub struct {
	res *geo.Result
	err error
}

func (g *geoClientStub) Nearby(ctx context.Context, in *geo.Request, opts ...grpc.CallOption) (*geo.Result, error) {
	return g.res, g.err
}

type rateClientStub struct {
	res *rate.Result
	err error
}

func (r *rateClientStub) GetRates(ctx context.Context, in *rate.Request, opts ...grpc.CallOption) (*rate.Result, error) {
	return r.res, r.err
}

func TestNearbyReturnsHotelIDsFromRatePlans(t *testing.T) {
	s := &Search{
		geoClient: &geoClientStub{res: &geo.Result{HotelIds: []string{"1", "2", "3"}}},
		rateClient: &rateClientStub{res: &rate.Result{RatePlans: []*rate.RatePlan{
			{HotelId: "2"},
			{HotelId: "1"},
		}}},
	}

	res, err := s.Nearby(context.Background(), &searchpb.NearbyRequest{
		Lat:     37.7749,
		Lon:     -122.4194,
		InDate:  "2015-04-09",
		OutDate: "2015-04-10",
	})
	if err != nil {
		t.Fatalf("Nearby returned error: %v", err)
	}
	if len(res.HotelIds) != 2 {
		t.Fatalf("expected 2 hotel ids, got %d", len(res.HotelIds))
	}
	if res.HotelIds[0] != "2" || res.HotelIds[1] != "1" {
		t.Fatalf("unexpected hotel ids: %v", res.HotelIds)
	}
}
