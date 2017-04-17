# Geo Index

Geo Index library

## Overview

Splits the earth surface in a grid. At each cell we can store data, such as list of points, count of points, etc. It can do KNearest and Range queries. For more detailed description check https://sudo.hailoapp.com/services/2015/02/18/geoindex/ .

### Demo

http://go-geoindex.appspot.com/static/nearest.html - Click to select the nearest points.

http://go-geoindex.appspot.com/static/cluster.html - A map with 100K points around the world. Zoom in and out to cluster. 

### API

```go
    type Driver struct {
        lat float64
        lon float64
        id string
        canAcceptJobs bool
    }

    // Implement Point interface
    func (d *Driver) Lat() float64 { return d.lat }
    func (d *Driver) Lon() float64 { return d.lon }
    func (d *Driver) Id() string { return d.id }

    // create points index with resolution (cell size) 0.5 km
    index := NewPointsIndex(Km(0.5))

    // Adds a point in the index, if a point with the same id exists it's removed and the new one is added
    index.Add(&Driver{"id1", lat, lng, true})
    index.Add(&Driver{"id2", lat, lng, false})

    // Removes a point from the index by id
    index.Remove("id1")

    // get the k-nearest points to a point, within some distance
    points := index.KNearest(&GeoPoint{id, lat, lng}, 5, Km(5), func(p Point) bool {
        return p.(* Driver).canAcceptJobs
    })

    // get the points within a range on the map
    points := index.Range(topLeftPoint, bottomRightPoint)
```

### Index types

There are several index types

```go
    NewPointsIndex(Km(0.5)) // Creates index that maintains points
    NewExpiringPointsIndex(Km(0.5), Minutes(5)) // Creates index that expires the points after some interval
    NewCountIndex(Km(0.5)) // Creates index that maintains counts of the points in each cell
    NewExpiringCountIndex(Km(0.5), Minutes(15)) // Creates index that maintains expiring count
    NewClusteringIndex() // index that clusters the points at different zoom levels, so we can create maps
    NewExpiringClusteringIndex(Minutes(15)) // index that clusters and expires the points at different zoom levels
                                            // so we can create real time maps of customer request, etc in the driver app
```

### Performance Benchmarks

    BenchmarkClusterIndexAdd                    500000          5395 ns/op
    BenchmarkClusterIndexStreetRange            100000         22207 ns/op
    BenchmarkClusterIndexCityRange              100000         16389 ns/op
    BenchmarkClusterIndexEuropeRange            50000          36559 ns/op

    BenchmarkExpiringClusterIndexAdd            300000          7124 ns/op
    BenchmarkExpiringClusterIndexStreetRange    50000          27030 ns/op
    BenchmarkExpiringClusterIndexCityRange      100000         22185 ns/op
    BenchmarkExpiringClusterIndexEuropeRange    30000          52080 ns/op

    BenchmarkCountIndexAdd                      1000000         1670 ns/op
    BenchmarkCountIndexCityRange                100000         20325 ns/op

    BenchmarkExpiringCountIndexAdd              500000          2808 ns/op
    BenchmarkExpiringCountIndexRange            50000          35791 ns/op

    BenchmarkPointIndexRange                    100000         15945 ns/op
    BenchmarkPointIndexAdd                      1000000         2416 ns/op
    BenchmarkPointIndexKNearest                 100000         13788 ns/op

    BenchmarkExpiringPointIndexAdd              500000          4324 ns/op
    BenchmarkExpiringPointIndexKNearest         100000         15638 ns/op
    BenchmarkExpiringPointIndexRange            100000         20386 ns/op
