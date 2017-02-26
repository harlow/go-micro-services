package geoindex

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

func pointsEqual(p1, p2 []Point) bool {
	return fmt.Sprintf("%v", p1) == fmt.Sprintf("%v", p2)
}

func toMap(points []Point) map[string]Point {
	result := make(map[string]Point)
	for _, p := range points {
		result[p.Id()] = p
	}
	return result
}

func pointsEqualIgnoreOrder(p1, p2 []Point) bool {
	return reflect.DeepEqual(toMap(p1), toMap(p2))
}

var (
	waterloo     = &GeoPoint{"Waterloo", 51.502973, -0.114723}
	kingsCross   = &GeoPoint{"Kings Cross", 51.529999, -0.124481}
	leicester    = &GeoPoint{"Leicester Square", 51.511291, -0.128242}
	coventGarden = &GeoPoint{"Covent Garden", 51.51276, -0.124507}
	totenham     = &GeoPoint{"Tottenham Court Road", 51.516206, -0.13087}
	picadilly    = &GeoPoint{"Piccadilly Circus", 51.50986, -0.1337}
	charring     = &GeoPoint{"Charing Cross", 51.508359, -0.124803}
	embankment   = &GeoPoint{"Embankment", 51.507312, -0.122367}
	oxford       = &GeoPoint{"Oxford Circus", 51.51511, -0.1417}
	westminster  = &GeoPoint{"Westminster", 51.501402, -0.125002}
	regentsPark  = &GeoPoint{"Regents Park", 51.52347, -0.1468}
	londonBridge = &GeoPoint{"London Bridge", 51.504674, -0.086006}
	brentCross   = &GeoPoint{"Brent Cross", 51.576599, -0.213336}
	lewisham     = &GeoPoint{"Lewisham", 51.46532, -0.0134}
	swanley      = &GeoPoint{"Swanley", 51.392994, 0.168716}
	watford      = &GeoPoint{"Watford", 51.65747, -0.41726}
	aylesbury    = &GeoPoint{"Aylesbury", 51.808615, -0.772219}
	aylesford    = &GeoPoint{"Aylesford", 51.28597, 0.507689}

	reykjavik = &GeoPoint{"Reykjavik", 64.15, -21.95}
	ankara    = &GeoPoint{"Ankara", 39.93, 32.86}

	points = [...]Point{leicester, coventGarden, totenham, picadilly, charring, embankment, oxford, westminster, regentsPark, londonBridge, brentCross, lewisham}
)

func tubeStations() []Point {
	file, _ := os.Open("test/tube.csv")
	defer file.Close()

	records, _ := csv.NewReader(file).ReadAll()

	points := make([]Point, 0)
	for _, record := range records {
		id := record[0]
		lat, _ := strconv.ParseFloat(record[1], 64)
		lon, _ := strconv.ParseFloat(record[2], 64)

		point := &GeoPoint{id, lat, lon}
		points = append(points, point)
	}

	return points
}

func worldCapitals() []Point {
	file, err := os.Open("test/capitals.csv")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'

	records, _ := reader.ReadAll()
	capitals := make([]Point, 0)

	for _, record := range records {
		id := record[0]
		lat, _ := strconv.ParseFloat(record[3], 64)
		lon, _ := strconv.ParseFloat(record[4], 64)

		capital := &GeoPoint{id, lat, lon}
		capitals = append(capitals, capital)
	}

	return capitals
}

var now time.Time

func getNow() time.Time {
	if now.IsZero() {
		return time.Now()
	} else {
		return now
	}
}

func toCountPoints(points []Point) []*CountPoint {
	result := make([]*CountPoint, len(points))

	for i, point := range points {
		result[i] = point.(*CountPoint)
	}

	return result
}
