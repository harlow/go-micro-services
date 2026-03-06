package frontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/harlow/go-micro-services/data"
	runtime "github.com/harlow/go-micro-services/internal/runtime"
	profile "github.com/harlow/go-micro-services/internal/services/profile/proto"
	search "github.com/harlow/go-micro-services/internal/services/search/proto"
	"github.com/harlow/go-micro-services/internal/trace"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// New returns a new server
func New(t opentracing.Tracer, searchconn, profileconn *grpc.ClientConn) *Frontend {
	return &Frontend{
		searchClient:  search.NewSearchClient(searchconn),
		profileClient: profile.NewProfileClient(profileconn),
		ratings:       loadRatings("data/hotel_ratings.json"),
		tracer:        t,
	}
}

// Frontend implements frontend service
type Frontend struct {
	searchClient  search.SearchClient
	profileClient profile.ProfileClient
	ratings       map[string]float64
	tracer        opentracing.Tracer
}

// Run the server
func (s *Frontend) Run(port int) error {
	mux := trace.NewServeMux(s.tracer)
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/hotels", http.HandlerFunc(s.searchHandler))

	return runtime.ServeHTTPGracefully(fmt.Sprintf(":%d", port), mux)
}

func (s *Frontend) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	// in/out dates from query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	// search for best hotels
	// TODO(hw): allow lat/lon from input params
	searchResp, err := s.searchClient.Nearby(ctx, &search.NearbyRequest{
		Lat:     37.7749,
		Lon:     -122.4194,
		InDate:  inDate,
		OutDate: outDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	// hotel profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &profile.Request{
		HotelIds: searchResp.HotelIds,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Hotels, s.ratings))
}

func loadRatings(path string) map[string]float64 {
	file := data.MustAsset(path)

	var rows []struct {
		ID     string  `json:"id"`
		Rating float64 `json:"rating"`
	}

	if err := json.Unmarshal(file, &rows); err != nil {
		log.Printf("failed to load hotel ratings from %s: %v", path, err)
		return map[string]float64{}
	}

	ratings := make(map[string]float64, len(rows))
	for _, row := range rows {
		ratings[row.ID] = row.Rating
	}

	return ratings
}

func logoURL(images []*profile.Image) string {
	for _, img := range images {
		if img.Default {
			return img.Url
		}
	}
	if len(images) > 0 {
		return images[0].Url
	}
	return ""
}

func formatAddress(addr *profile.Address) string {
	if addr == nil {
		return ""
	}

	parts := []string{
		strings.TrimSpace(strings.TrimSpace(addr.StreetNumber) + " " + strings.TrimSpace(addr.StreetName)),
		strings.TrimSpace(addr.City),
		strings.TrimSpace(addr.State),
		strings.TrimSpace(addr.PostalCode),
	}

	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			clean = append(clean, part)
		}
	}

	return strings.Join(clean, ", ")
}

// return a geoJSON response that allows google map to plot points directly on map
// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
func geoJSONResponse(hs []*profile.Hotel, ratings map[string]float64) map[string]interface{} {
	fs := []interface{}{}

	for _, h := range hs {
		if h == nil || h.Address == nil {
			continue
		}

		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]interface{}{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
				"description":  h.Description,
				"address_line": formatAddress(h.Address),
				"city":         h.Address.GetCity(),
				"state":        h.Address.GetState(),
				"country":      h.Address.GetCountry(),
				"postal_code":  h.Address.GetPostalCode(),
				"rating":       ratings[h.Id],
				"logo_url":     logoURL(h.Images),
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float32{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}
