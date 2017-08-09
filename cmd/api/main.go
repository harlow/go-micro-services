package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/harlow/go-micro-services/pb/profile"
	"github.com/harlow/go-micro-services/pb/search"
)

func main() {
	e := newEnv()
	http.Handle("/", handler(e))
	log.Fatal(http.ListenAndServe(e.serviceAddr(), nil))
}

func handler(e *env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		ctx := r.Context()

		// in/out dates from query params
		inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
		if inDate == "" || outDate == "" {
			http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
			return
		}

		// nearby hotel ids
		// TODO(hw): allow lat/lon from input params
		nearby, err := e.SearchClient.Nearby(ctx, &search.NearbyRequest{
			Lat:     37.7749,
			Lon:     -122.4194,
			InDate:  inDate,
			OutDate: outDate,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// hotel profiles
		// TODO(hw): allow custom locale from input params
		profile, err := e.ProfileClient.GetProfiles(ctx, &profile.Request{
			HotelIds: nearby.HotelIds,
			Locale:   "en",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// geo json response body
		body := geoJSON(profile.Hotels)
		json.NewEncoder(w).Encode(body)
	})
}

// build a geoJSON response that allows google map to plot points directly on map
// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
func geoJSON(hotels []*profile.Hotel) response {
	r := response{
		Type: "FeatureCollection",
	}

	for _, hotel := range hotels {
		f := feature{
			Type: "Feature",
			ID:   hotel.Id,
		}
		f.Properties.Name = hotel.Name
		f.Properties.PhoneNumber = hotel.PhoneNumber
		f.Geometry.Type = "Point"
		f.Geometry.Coordinates = []float32{
			hotel.Address.Lon,
			hotel.Address.Lat,
		}
		r.Features = append(r.Features, f)
	}

	return r
}

type response struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

type feature struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Properties struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
	} `json:"properties"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float32 `json:"coordinates"`
	} `json:"geometry"`
}
