package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/search"
	"github.com/harlow/go-micro-services/tracing"
)

func main() {
	var (
		port        = flag.String("port", "8080", "The server port")
		searchAddr  = flag.String("searchaddr", "search:8080", "Search service addr")
		profileAddr = flag.String("profileaddr", "profile:8080", "Profile service addr")
		jaegerAddr  = flag.String("jaegeraddr", "jaeger:6831", "Jaeger server addr")
	)
	flag.Parse()

	var (
		tracer        = tracing.Init("api", *jaegerAddr)
		searchClient  = search.NewSearchClient(tracing.MustDial(*searchAddr, tracer))
		profileClient = profile.NewProfileClient(tracing.MustDial(*profileAddr, tracer))
	)
	srv := &server{
		searchClient:  searchClient,
		profileClient: profileClient,
	}

	mux := tracing.NewServeMux(tracer)
	mux.Handle("/", http.HandlerFunc(srv.searchHandler))
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}

// server holds open the grpc connections and serves the JSON http endpoint
type server struct {
	searchClient  search.SearchClient
	profileClient profile.ProfileClient
}

func (s *server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// in/out dates from query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

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

	// return a geoJSON response that allows google map to plot points directly on map
	// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":     "FeatureCollection",
		"features": buildFeatures(profileResp.Hotels),
	})
}

// returns a slice of features from hotel records, the feature nodes
// can be used for plotting on map
func buildFeatures(hs []*profile.Hotel) []interface{} {
	fs := []interface{}{}
	for _, h := range hs {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]string{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
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
	return fs
}
