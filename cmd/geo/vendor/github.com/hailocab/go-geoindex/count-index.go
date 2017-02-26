package geoindex

import (
	"fmt"
)

type CountIndex struct {
	index           *geoIndex
	currentPosition map[string]Point
}

type CountPoint struct {
	*GeoPoint
	Count interface{}
}

func (p *CountPoint) String() string {
	return fmt.Sprintf("%f %f %d", p.Lat(), p.Lon(), p.Count)
}

// NewCountIndex creates an index which counts the points in each cell.
func NewCountIndex(resolution Meters) *CountIndex {
	newCounter := func() interface{} {
		return &singleValueAccumulatingCounter{}
	}

	return &CountIndex{newGeoIndex(resolution, newCounter), make(map[string]Point)}
}

// NewExpiringCountIndex creates an index, which maintains an expiring counter for each cell.
func NewExpiringCountIndex(resolution Meters, expiration Minutes) *CountIndex {
	newExpiringCounter := func() interface{} {
		return newExpiringCounter(expiration)
	}

	return &CountIndex{newGeoIndex(resolution, newExpiringCounter), make(map[string]Point)}
}

func (index *CountIndex) Clone() *CountIndex {
	clone := &CountIndex{}

	// Copy all entries from current positions
	clone.currentPosition = make(map[string]Point, len(index.currentPosition))
	for k, v := range index.currentPosition {
		clone.currentPosition[k] = v
	}

	// Copying underlying geoindex data
	clone.index = index.index.Clone()

	return clone
}

// Add adds a point.
func (countIndex *CountIndex) Add(point Point) {
	countIndex.Remove(point.Id())
	countIndex.currentPosition[point.Id()] = point
	countIndex.index.AddEntryAt(point).(counter).Add(point)
}

// Remove removes a point.
func (countIndex *CountIndex) Remove(id string) {
	if prev, ok := countIndex.currentPosition[id]; ok {
		countIndex.index.GetEntryAt(prev).(counter).Remove(prev)
		delete(countIndex.currentPosition, id)
	}
}

// Range returns the counters within some lat, lng range.
func (countIndex *CountIndex) Range(topLeft Point, bottomRight Point) []Point {
	counters := countIndex.index.Range(topLeft, bottomRight)

	points := make([]Point, 0)

	for _, c := range counters {
		if c.(counter).Point() != nil {
			points = append(points, c.(counter).Point())
		}
	}

	return points
}

// KNearest just to satisfy an interface. Doesn't make much sense for count index.
func (index *CountIndex) KNearest(point Point, k int, maxDistance Meters, accept func(p Point) bool) []Point {
	panic("Unsupported operation")
}
