package geoindex

type ClusteringIndex struct {
	streetLevel *PointsIndex
	cityLevel   *CountIndex
	worldLevel  *CountIndex
}

var (
	streetLevel = Km(45)
	cityLevel   = Km(1000)
)

// NewClusteringIndex creates index that clusters the points at three levels with cell size 0.5, 5 and 500km.
// Useful for creating maps.
func NewClusteringIndex() *ClusteringIndex {
	index := &ClusteringIndex{}
	index.streetLevel = NewPointsIndex(Km(0.5))
	index.cityLevel = NewCountIndex(Km(10))
	index.worldLevel = NewCountIndex(Km(500))

	return index
}

// NewExpiringClusteringIndex creates index that clusters the points at three levels with cell size 0.5, 5 and 500km and
// expires them after expiration minutes.
func NewExpiringClusteringIndex(expiration Minutes) *ClusteringIndex {
	index := &ClusteringIndex{}
	index.streetLevel = NewExpiringPointsIndex(Km(0.5), expiration)
	index.cityLevel = NewExpiringCountIndex(Km(10), expiration)
	index.worldLevel = NewExpiringCountIndex(Km(500), expiration)

	return index
}

func (index *ClusteringIndex) Clone() *ClusteringIndex {
	clone := &ClusteringIndex{}

	clone.streetLevel = index.streetLevel.Clone()
	clone.cityLevel = index.cityLevel.Clone()
	clone.worldLevel = index.worldLevel.Clone()

	return clone
}

// Add adds a point.
func (index *ClusteringIndex) Add(point Point) {
	index.streetLevel.Add(point)
	index.cityLevel.Add(point)
	index.worldLevel.Add(point)
}

// Remove removes a point.
func (index *ClusteringIndex) Remove(id string) {
	index.streetLevel.Remove(id)
	index.cityLevel.Remove(id)
	index.worldLevel.Remove(id)
}

// Range returns points or count points depending on the size of the topLeft and bottomRight range.
func (index *ClusteringIndex) Range(topLeft Point, bottomRight Point) []Point {
	dist := distance(topLeft, bottomRight)

	if dist < streetLevel {
		return index.streetLevel.Range(topLeft, bottomRight)
	} else if dist < cityLevel {
		return index.cityLevel.Range(topLeft, bottomRight)
	} else {
		return index.worldLevel.Range(topLeft, bottomRight)
	}
}

// KNearest returns the K-Nearest points near point within maxDistance, that match the accept function.
func (index *ClusteringIndex) KNearest(point Point, k int, maxDistance Meters, accept func(p Point) bool) []Point {
	return index.streetLevel.KNearest(point, k, maxDistance, accept)
}
