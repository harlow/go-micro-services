package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/harlow/go-micro-services/data"
	runtime "github.com/harlow/go-micro-services/internal/runtime"
	profile "github.com/harlow/go-micro-services/internal/services/profile/proto"
	search "github.com/harlow/go-micro-services/internal/services/search/proto"
	"github.com/harlow/go-micro-services/internal/trace"
	"google.golang.org/grpc"
)

// New returns a new server
func New(searchconn, profileconn *grpc.ClientConn) *Frontend {
	return &Frontend{
		searchClient:  search.NewSearchClient(searchconn),
		profileClient: profile.NewProfileClient(profileconn),
		ratings:       loadRatings("data/hotel_ratings.json"),
	}
}

// Frontend implements frontend service
type Frontend struct {
	searchClient  search.SearchClient
	profileClient profile.ProfileClient
	ratings       map[string]float64
}

// Run the server
func (s *Frontend) Run(port int) error {
	mux := trace.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.Handle("/hotels", http.HandlerFunc(s.searchHandler))
	mux.Handle("/healthz", http.HandlerFunc(s.healthHandler))
	mux.Handle("/readyz", http.HandlerFunc(s.readyHandler))

	return runtime.ServeHTTPGracefully(fmt.Sprintf(":%d", port), mux)
}

func (s *Frontend) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	inDate, outDate, err := parseDateRange(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
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
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR", "search service unavailable")
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
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR", "profile service unavailable")
		return
	}

	writeJSON(w, http.StatusOK, geoJSONResponse(profileResp.Hotels, s.ratings))
}

func (s *Frontend) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Frontend) readyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	_, err := s.searchClient.Nearby(ctx, &search.NearbyRequest{
		Lat:     37.7749,
		Lon:     -122.4194,
		InDate:  "2015-04-09",
		OutDate: "2015-04-10",
	})
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "NOT_READY", "search dependency is not ready")
		return
	}

	_, err = s.profileClient.GetProfiles(ctx, &profile.Request{
		HotelIds: []string{},
		Locale:   "en",
	})
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "NOT_READY", "profile dependency is not ready")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func parseDateRange(r *http.Request) (string, string, error) {
	inDate := r.URL.Query().Get("inDate")
	outDate := r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		return "", "", fmt.Errorf("missing required query params: inDate and outDate")
	}

	const dateFmt = "2006-01-02"
	inParsed, err := time.Parse(dateFmt, inDate)
	if err != nil {
		return "", "", fmt.Errorf("invalid inDate format, expected YYYY-MM-DD")
	}
	outParsed, err := time.Parse(dateFmt, outDate)
	if err != nil {
		return "", "", fmt.Errorf("invalid outDate format, expected YYYY-MM-DD")
	}
	if !outParsed.After(inParsed) {
		return "", "", fmt.Errorf("outDate must be after inDate")
	}
	if outParsed.Sub(inParsed) > 30*24*time.Hour {
		return "", "", fmt.Errorf("date range cannot exceed 30 days")
	}

	return inDate, outDate, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
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
