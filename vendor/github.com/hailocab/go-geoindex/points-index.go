// Package geoindex provides in memory geoindex implementation. It works by splitting the earth surface
// into grid with fixed size cells and storing data in each cell. The data can be points, count of points,
// and expiring points/counts. Has Range and K-Nearest queries.
package geoindex

import (
	"math"
	"sort"
)

// A geoindex that stores points.
type PointsIndex struct {
	index           *geoIndex
	currentPosition map[string]Point
}

// NewPointsIndex creates new PointsIndex that maintains the points in each cell.
func NewPointsIndex(resolution Meters) *PointsIndex {
	newSet := func() interface{} {
		return newSet()
	}

	return &PointsIndex{newGeoIndex(resolution, newSet), make(map[string]Point)}
}

// NewExpiringPointsIndex creates new PointIndex that expires the points in each cell after expiration minutes.
func NewExpiringPointsIndex(resolution Meters, expiration Minutes) *PointsIndex {
	currentPosition := make(map[string]Point)

	newExpiringSet := func() interface{} {
		set := newExpiringSet(expiration)

		set.OnExpire(func(id string, value interface{}) {
			point := value.(Point)
			delete(currentPosition, point.Id())
		})

		return set
	}

	return &PointsIndex{newGeoIndex(resolution, newExpiringSet), currentPosition}
}

func (pi *PointsIndex) Clone() *PointsIndex {
	clone := &PointsIndex{}

	// Copy all entries from current positions
	clone.currentPosition = make(map[string]Point, len(pi.currentPosition))
	for k, v := range pi.currentPosition {
		clone.currentPosition[k] = v
	}

	// Copying underlying geoindex data
	clone.index = pi.index.Clone()

	return clone
}

// Get gets a point from the index given an id.
func (points *PointsIndex) Get(id string) Point {
	if point, ok := points.currentPosition[id]; ok {
		// first it gets the set of the currentPosition and then gets the point from the set
		// this is done so it triggers expiration on expiringSet, and returns nil if a point has expired
		if result, resultOk := points.index.GetEntryAt(point).(set).Get(id); resultOk {
			return result.(Point)
		}
	}
	return nil
}

// GetAll get all Points from the index as a map from id to point
func (points *PointsIndex) GetAll() map[string]Point {
	newpoints := make(map[string]Point, 0)
	for i, p := range points.currentPosition {
		newpoints[i] = p
	}
	return newpoints
}

// Add adds a point to the index. If a point with the same Id already exists it gets replaced.
func (points *PointsIndex) Add(point Point) {
	points.Remove(point.Id())
	newSet := points.index.AddEntryAt(point).(set)
	newSet.Add(point.Id(), point)
	points.currentPosition[point.Id()] = point
}

// Remove removes a point from the index.
func (points *PointsIndex) Remove(id string) {
	if prevPoint, ok := points.currentPosition[id]; ok {
		set := points.index.GetEntryAt(prevPoint).(set)
		set.Remove(prevPoint.Id())
		delete(points.currentPosition, prevPoint.Id())
	}
}

func between(value float64, min float64, max float64) bool {
	return value >= min && value <= max
}

func getPoints(entries []interface{}, accept func(point Point) bool) []Point {
	result := make([]Point, 0)
	result = getPointsAppend(result, entries, accept)
	return result
}

func getPointsAppend(s []Point, entries []interface{}, accept func(point Point) bool) []Point {
	for _, entry := range entries {
		pointsSetEntry := (entry).(set)

		for _, value := range pointsSetEntry.Values() {
			point := value.(Point)
			if accept(point) {
				s = append(s, point)
			}
		}
	}
	return s
}

// Range returns the points within the range defined by top left and bottom right.
func (points *PointsIndex) Range(topLeft Point, bottomRight Point) []Point {
	entries := points.index.Range(topLeft, bottomRight)
	accept := func(point Point) bool {
		return between(point.Lat(), bottomRight.Lat(), topLeft.Lat()) &&
			between(point.Lon(), topLeft.Lon(), bottomRight.Lon())
	}

	return getPoints(entries, accept)
}

type sortedPoints struct {
	points []Point
	point  Point
}

func (p *sortedPoints) Len() int {
	return len(p.points)
}

func (p *sortedPoints) Swap(i, j int) {
	p.points[i], p.points[j] = p.points[j], p.points[i]
}

func (p *sortedPoints) Less(i, j int) bool {
	return approximateSquareDistance(p.points[i], p.point) < approximateSquareDistance(p.points[j], p.point)
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// KNearest returns the k nearest points near point within maxDistance that match the accept criteria.
func (points *PointsIndex) KNearest(point Point, k int, maxDistance Meters, accept func(p Point) bool) []Point {
	nearbyPoints := make([]Point, 0)
	pointEntry := points.index.GetEntryAt(point).(set)
	nearbyPoints = append(nearbyPoints, getPoints([]interface{}{pointEntry}, accept)...)

	totalCount := 0
	idx := cellOf(point, points.index.resolution)
	// Explicitely assign a greater max distance so that we definitely return enough points
	// and make sure it searches at least one square away.
	coarseMaxDistance := math.Max(float64(maxDistance)*2.0, float64(points.index.resolution)*2.0+0.01)

	for d := 1; float64(d)*float64(points.index.resolution) <= coarseMaxDistance; d++ {
		oldCount := len(nearbyPoints)

		nearbyPoints = getPointsAppend(nearbyPoints, points.index.get(idx.x-d, idx.x+d, idx.y+d, idx.y+d), accept)
		nearbyPoints = getPointsAppend(nearbyPoints, points.index.get(idx.x-d, idx.x+d, idx.y-d, idx.y-d), accept)
		nearbyPoints = getPointsAppend(nearbyPoints, points.index.get(idx.x-d, idx.x-d, idx.y-d+1, idx.y+d-1), accept)
		nearbyPoints = getPointsAppend(nearbyPoints, points.index.get(idx.x+d, idx.x+d, idx.y-d+1, idx.y+d-1), accept)

		totalCount += len(nearbyPoints) - oldCount

		if totalCount > k {
			break
		}
	}

	sortedPoints := &sortedPoints{nearbyPoints, point}
	sort.Sort(sortedPoints)

	k = min(k, len(sortedPoints.points))

	// filter points which longer than maxDistance away from point.
	for i, nearbyPoint := range sortedPoints.points {
		if Distance(point, nearbyPoint) > maxDistance || i == k {
			k = i
			break
		}
	}

	return sortedPoints.points[0:k]
}

// PointsWithin returns all points with distance of point that match the accept criteria.
func (points *PointsIndex) PointsWithin(point Point, distance Meters, accept func(p Point) bool) []Point {

	d := int(distance / points.index.resolution)

	if d == 0 {
		d = 1
	}

	idx := cellOf(point, points.index.resolution)

	nearbyPoints := make([]Point, 0)
	nearbyPoints = getPointsAppend(nearbyPoints, points.index.get(idx.x-d, idx.x+d, idx.y-d, idx.y+d), accept)

	// filter points which longer than maxDistance away from point.
	withinPoints := make([]Point, 0)
	for _, nearbyPoint := range nearbyPoints {
		if Distance(point, nearbyPoint) < distance {
			withinPoints = append(withinPoints, nearbyPoint)
		}
	}

	return withinPoints
}
