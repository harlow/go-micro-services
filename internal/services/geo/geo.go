package geo

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"sort"

	"github.com/harlow/go-micro-services/data"
	runtime "github.com/harlow/go-micro-services/internal/runtime"
	geo "github.com/harlow/go-micro-services/internal/services/geo/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	maxSearchRadius  = 10
	maxSearchResults = 20
	earthRadiusKm    = 6371.0
)

// point represents a hotel's geo location on map.
type point struct {
	Pid  string  `json:"hotelId"`
	Plat float64 `json:"lat"`
	Plon float64 `json:"lon"`
}

// New returns a new server.
func New() *Geo {
	return &Geo{
		points: loadPoints("data/geo.json"),
	}
}

// Geo implements the geo service.
type Geo struct {
	points []*point
}

// Run starts the server.
func (s *Geo) Run(port int) error {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	geo.RegisterGeoServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return runtime.ServeGRPCGracefully(lis, srv)
}

// Nearby returns all hotels within a given distance.
func (s *Geo) Nearby(ctx context.Context, req *geo.Request) (*geo.Result, error) {
	_ = ctx

	res := &geo.Result{}
	for _, p := range s.getNearbyPoints(float64(req.Lat), float64(req.Lon)) {
		res.HotelIds = append(res.HotelIds, p.Pid)
	}

	return res, nil
}

func (s *Geo) getNearbyPoints(lat, lon float64) []*point {
	type candidate struct {
		point *point
		dist  float64
	}

	candidates := make([]candidate, 0, len(s.points))
	for _, p := range s.points {
		d := haversineKm(lat, lon, p.Plat, p.Plon)
		if d <= maxSearchRadius {
			candidates = append(candidates, candidate{point: p, dist: d})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist < candidates[j].dist
	})

	if len(candidates) > maxSearchResults {
		candidates = candidates[:maxSearchResults]
	}

	out := make([]*point, 0, len(candidates))
	for _, c := range candidates {
		out = append(out, c.point)
	}
	return out
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := math.Pi / 180.0
	dLat := (lat2 - lat1) * toRad
	dLon := (lon2 - lon1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func loadPoints(path string) []*point {
	file := data.MustAsset(path)

	var points []*point
	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("failed to load geo points: %v", err)
	}

	return points
}
