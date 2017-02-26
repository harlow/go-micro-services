package geoindex

var (
	minLon          = -180.0
	minLat          = -90.0
	latDegreeLength = Km(111.0)
	lonDegreeLength = Km(85.0)
)

type Meters float64

func Km(km float64) Meters {
	return Meters(km * 1000)
}

func Meter(meters float64) Meters {
	return Meters(meters)
}

type cell struct {
	x int
	y int
}

func cellOf(point Point, resolution Meters) cell {
	x := int((-minLat + point.Lat()) * float64(latDegreeLength) / float64(resolution))
	y := int((-minLon + point.Lon()) * float64(lonDegreeLength) / float64(resolution))

	return cell{x, y}
}

type geoIndex struct {
	resolution Meters
	index      map[cell]interface{}
	newEntry   func() interface{}
}

// Creates new geo index with resolution a function that returns a new entry that is stored in each cell.
func newGeoIndex(resolution Meters, newEntry func() interface{}) *geoIndex {
	return &geoIndex{resolution, make(map[cell]interface{}), newEntry}
}

func (i *geoIndex) Clone() *geoIndex {
	clone := &geoIndex{
		resolution: i.resolution,
		index:      make(map[cell]interface{}, len(i.index)),
		newEntry:   i.newEntry,
	}
	for k, v := range i.index {
		set, ok := v.(set)
		if !ok {
			panic("Cannot cast value to set")
		}
		clone.index[k] = set.Clone()
	}

	return clone
}

// AddEntryAt adds an entry if missing, returns the entry at specific position.
func (geoIndex *geoIndex) AddEntryAt(point Point) interface{} {
	square := cellOf(point, geoIndex.resolution)

	if _, ok := geoIndex.index[square]; !ok {
		geoIndex.index[square] = geoIndex.newEntry()
	}

	return geoIndex.index[square]
}

// GetEntryAt gets an entry from the geoindex, if missing returns an empty entry without changing the index.
func (geoIndex *geoIndex) GetEntryAt(point Point) interface{} {
	square := cellOf(point, geoIndex.resolution)

	entries, ok := geoIndex.index[square]
	if !ok {
		return geoIndex.newEntry()
	}

	return entries
}

// Range returns the index entries within lat, lng range.
func (geoIndex *geoIndex) Range(topLeft Point, bottomRight Point) []interface{} {
	topLeftIndex := cellOf(topLeft, geoIndex.resolution)
	bottomRightIndex := cellOf(bottomRight, geoIndex.resolution)

	return geoIndex.get(bottomRightIndex.x, topLeftIndex.x, topLeftIndex.y, bottomRightIndex.y)
}

func (geoIndex *geoIndex) get(minx int, maxx int, miny int, maxy int) []interface{} {
	entries := make([]interface{}, 0, 0)

	for x := minx; x <= maxx; x++ {
		for y := miny; y <= maxy; y++ {
			if indexEntry, ok := geoIndex.index[cell{x, y}]; ok {
				entries = append(entries, indexEntry)
			}
		}
	}

	return entries
}

func (g *geoIndex) getCells(minx int, maxx int, miny int, maxy int) []cell {
	indices := make([]cell, 0)

	for x := minx; x <= maxx; x++ {
		for y := miny; y <= maxy; y++ {
			indices = append(indices, cell{x, y})
		}
	}

	return indices
}
