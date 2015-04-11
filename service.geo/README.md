# service.geo

A service allowing caller to filter Hotels based on a point location using a bounding box.

```
conn, err := grpc.Dial(*geoServerAddr)

if err != nil {
  log.Fatalf(err)
}

rect := &pb.BoundingBox{
  &pb.Point{400000000, -750000000},
  &pb.Point{420000000, -730000000},
}

client := pb.NewGeoClient(conn)
stream, err := client.NearbyLocations(context.Background(), rect)

if err != nil {
  log.Fatalf("%v.NearbyLocations(_) = _, %v", conn, err)
}

for {
  location, err := stream.Recv()
  if err == io.EOF {
    break
  }
  if err != nil {
    log.Fatalf("%v.NearbyLocations(_) = _, %v", conn, err)
  }
  log.Println(location)
}
```
