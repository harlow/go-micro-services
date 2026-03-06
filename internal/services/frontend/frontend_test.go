package frontend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	profile "github.com/harlow/go-micro-services/internal/services/profile/proto"
	search "github.com/harlow/go-micro-services/internal/services/search/proto"
	"google.golang.org/grpc"
)

type fakeSearchClient struct {
	resp *search.SearchResult
	err  error
}

func (f *fakeSearchClient) Nearby(ctx context.Context, in *search.NearbyRequest, opts ...grpc.CallOption) (*search.SearchResult, error) {
	return f.resp, f.err
}

type fakeProfileClient struct {
	resp *profile.Result
	err  error
}

func (f *fakeProfileClient) GetProfiles(ctx context.Context, in *profile.Request, opts ...grpc.CallOption) (*profile.Result, error) {
	return f.resp, f.err
}

func TestSearchHandler_ValidatesDateRange(t *testing.T) {
	svc := &Frontend{
		searchClient:  &fakeSearchClient{},
		profileClient: &fakeProfileClient{},
		ratings:       map[string]float64{},
	}

	req := httptest.NewRequest(http.MethodGet, "/hotels?inDate=2015/04/09&outDate=2015-04-10", nil)
	rr := httptest.NewRecorder()
	svc.searchHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if body["error"]["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("error.code = %q, want INVALID_ARGUMENT", body["error"]["code"])
	}
}

func TestSearchHandler_ReturnsGeoJSON(t *testing.T) {
	svc := &Frontend{
		searchClient: &fakeSearchClient{
			resp: &search.SearchResult{
				HotelIds: []string{"hotel-1"},
			},
		},
		profileClient: &fakeProfileClient{
			resp: &profile.Result{
				Hotels: []*profile.Hotel{
					{
						Id:   "hotel-1",
						Name: "Hotel One",
						Address: &profile.Address{
							StreetNumber: "1",
							StreetName:   "Market St",
							City:         "San Francisco",
							State:        "CA",
							PostalCode:   "94105",
							Lat:          37.79,
							Lon:          -122.40,
						},
					},
				},
			},
		},
		ratings: map[string]float64{"hotel-1": 4.4},
	}

	req := httptest.NewRequest(http.MethodGet, "/hotels?inDate=2015-04-09&outDate=2015-04-10", nil)
	rr := httptest.NewRecorder()
	svc.searchHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if body["type"] != "FeatureCollection" {
		t.Fatalf("type = %v, want FeatureCollection", body["type"])
	}
}

func TestReadyHandler_DownstreamFailure(t *testing.T) {
	svc := &Frontend{
		searchClient:  &fakeSearchClient{err: context.DeadlineExceeded},
		profileClient: &fakeProfileClient{},
		ratings:       map[string]float64{},
	}

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	svc.readyHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
}
