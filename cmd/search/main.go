package main

import (
	"log"

	"github.com/harlow/go-micro-services/pb/geo"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// closet hotel w/ lowest price (best discount?) and best reviews

// data needed:
// 1. user data (lat/lon/pref room type)
// 2. room data (pricing)
// 3. hotel profile data (lat/lon, reviews)

// www -> rank -> geoSrv (lat/lon of hotels)
//             -> rateSrv (we'll have rates)
//             -> reviewSrv (need agg of reviews, pos/neg etc)
//        rank calculate []
// www <- rank (hotelID, ratePlan)
// www -> profileSrv -> localeSrv
// www <- profileSrv

type room struct {
}

// either needs to return a number, or augment the argument
// multiplier (sin, cos, etc) (distance)
// figure out top/bottom and then rate compared to those?

type rule func(a, b int) int

// []hotels, rank each hotel
// user (geo point, pref to type)

type ranker struct {
	// hotels []hotel{id, score, lat/lon, rating, etc}
	// user user{lat/lon, pref type?}
	// rules []rule
}

func geoScore(a, b int) int {
	return 0
}

func reviewScore(a, b int) int {
	return 0
}

func priceScore(a, b int) int {
	return 0
}

// rules, features, scores, adjustments

var rules = []rule{
	geoScore,
	reviewScore,
	priceScore,
}

// compare against the others (1,2,3,4,etc)
// compare against scale

func main() {
	ctx := context.Background()
	geoClient := geo.NewGeoClient(mustDial(""))

	nearby, err := geoClient.Nearby(ctx, &geo.Request{
		Lat: 37.7749,
		Lon: -122.4194,
	})
	if err != nil {
		log.Fatalf("nearby error: %v", err)
	}

	log.Println(nearby)
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
