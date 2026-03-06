package rate

import (
	"testing"

	ratepb "github.com/harlow/go-micro-services/internal/services/rate/proto"
	"golang.org/x/net/context"
)

func TestGetRatesReturnsOnlyMatchingStays(t *testing.T) {
	s := &Rate{rateTable: map[stay]*ratepb.RatePlan{
		{HotelID: "1", InDate: "2015-04-09", OutDate: "2015-04-10"}: {HotelId: "1"},
		{HotelID: "2", InDate: "2015-04-09", OutDate: "2015-04-10"}: {HotelId: "2"},
	}}

	res, err := s.GetRates(context.Background(), &ratepb.Request{
		HotelIds: []string{"1", "3", "2"},
		InDate:   "2015-04-09",
		OutDate:  "2015-04-10",
	})
	if err != nil {
		t.Fatalf("GetRates returned error: %v", err)
	}
	if len(res.RatePlans) != 2 {
		t.Fatalf("expected 2 rate plans, got %d", len(res.RatePlans))
	}
	if res.RatePlans[0].HotelId != "1" || res.RatePlans[1].HotelId != "2" {
		t.Fatalf("unexpected hotel id order: %q, %q", res.RatePlans[0].HotelId, res.RatePlans[1].HotelId)
	}
}
