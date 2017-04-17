package geoindex

import (
	"fmt"
	"math"
)

var (
	earthRadius = Km(6371.0)
)

type Direction int

const (
	NorthEast Direction = iota
	East
	SouthEast
	South
	SouthWest
	West
	NorthWest
	North
)

type Point interface {
	Id() string
	Lat() float64
	Lon() float64
}

// Point implementation.
type GeoPoint struct {
	Pid  string  `json:"Id"`
	Plat float64 `json:"Lat"`
	Plon float64 `json:"Lon"`
}

func NewGeoPoint(id string, lat, lon float64) *GeoPoint {
	return &GeoPoint{id, lat, lon}
}

func (p *GeoPoint) String() string {
	return fmt.Sprintf("%s %f %f", p.Id(), p.Lat(), p.Lon())
}

func (p *GeoPoint) Id() string {
	return p.Pid
}

func (p *GeoPoint) Lat() float64 {
	return p.Plat
}

func (p *GeoPoint) Lon() float64 {
	return p.Plon
}

// DirectionTo returns the direction from p1 to p2
func DirectionTo(p1, p2 Point) Direction {
	bearing := BearingTo(p1, p2)

	index := bearing - 22.5

	if index < 0 {
		index += 360
	}
	indexInt := int(index / 45.0)

	return Direction(indexInt)
}

// BearingTo returns the bearing from p1 to p2
func BearingTo(p1, p2 Point) float64 {
	dLon := toRadians(p2.Lon() - p1.Lon())

	lat1 := toRadians(p1.Lat())
	lat2 := toRadians(p2.Lat())

	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	brng := toDegrees(math.Atan2(y, x))

	return brng
}

func Distance(p1, p2 Point) Meters {
	return distance(p1, p2)
}

func toDegrees(x float64) float64 {
	return x * 180.0 / math.Pi
}

func toRadians(x float64) float64 {
	return x * math.Pi / 180.0
}

func distance(p1, p2 Point) Meters {

	dLat := toRadians(p2.Lat() - p1.Lat())
	dLng := toRadians(p2.Lon() - p1.Lon())
	sindLat := math.Sin(dLat / 2)
	sindLng := math.Sin(dLng / 2)
	a := math.Pow(sindLat, 2) + math.Pow(sindLng, 2)*math.Cos(toRadians(p1.Lat()))*math.Cos(toRadians(p2.Lat()))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	dist := float64(earthRadius) * c

	return Meters(dist)
}

type lonDegreeDistance map[int]Meters

func (lonDist lonDegreeDistance) get(lat float64) Meters {
	latIndex := int(lat * 10)
	latRounded := float64(latIndex) / 10

	if value, ok := lonDist[latIndex]; ok {
		return value
	} else {
		dist := distance(&GeoPoint{"", latRounded, 0.0}, &GeoPoint{"", latRounded, 1.0})
		lonDist[latIndex] = dist
		return dist
	}
}

var (
	lonLength = lonDegreeDistance{}
)

// Calculates approximate distance between two points using euclidian distance. The assumption here
// is that the points are relatively close to each other.
func approximateSquareDistance(p1, p2 Point) Meters {
	avgLat := (p1.Lat() + p2.Lat()) / 2.0

	latLen := math.Abs(p1.Lat()-p2.Lat()) * float64(latDegreeLength)
	lonLen := math.Abs(p1.Lon()-p2.Lon()) * float64(lonLength.get(avgLat))

	return Meters(latLen*latLen + lonLen*lonLen)
}
