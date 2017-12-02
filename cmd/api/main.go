package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/search"
	"google.golang.org/grpc"
)

func main() {
	var (
		port        = flag.String("port", "8080", "The server port")
		searchAddr  = flag.String("searchaddr", "search:8080", "Search service addr")
		profileAddr = flag.String("profileaddr", "profile:8080", "Profile service addr")
	)
	flag.Parse()

	srv := &server{
		searchClient:  search.NewSearchClient(mustDial(*searchAddr)),
		profileClient: profile.NewProfileClient(mustDial(*profileAddr)),
	}

	http.HandleFunc("/", srv.searchHandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// mustDial ensures a tcp connection to specified address.
func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
		panic(err)
	}
	return conn
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

// returns a slice of features from hotel records
func buildFeatures(hotels []*profile.Hotel) []interface{} {
	fs := []interface{}{}
	for _, h := range hotels {
		fs = append(fs, buildFeature(h))
	}
	return fs
}

// returns a feature node for plotting on map
func buildFeature(hotel *profile.Hotel) map[string]interface{} {
	return map[string]interface{}{
		"type": "Feature",
		"id":   hotel.Id,
		"properties": map[string]string{
			"name":         hotel.Name,
			"phone_number": hotel.PhoneNumber,
		},
		"geometry": map[string]interface{}{
			"type": "Point",
			"coordinates": []float32{
				hotel.Address.Lon,
				hotel.Address.Lat,
			},
		},
	}
}
