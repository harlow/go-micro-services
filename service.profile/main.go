package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "net"

  pb "github.com/harlow/go-micro-services/service.profile/proto"

  "golang.org/x/net/context"
  "google.golang.org/grpc"
)

var (
  port       = flag.Int("port", 10003, "The server port")
  jsonDBFile = flag.String("json_db_file", "data/profiles.json", "A json file containing a list of customers")
  serverName = "service.profile"
)

type profileServer struct {
  hotels []*pb.Hotel
}

// VerifyToken finds a customer by authentication token.
func (s *profileServer) GetProfiles(ctx context.Context, args *pb.Args) (*pb.Reply, error) {
  reply := new(pb.Reply)
  for _, hotel := range s.hotels {
    if inRange(hotel.Id, args.HotelIds) {
      reply.Hotels = append(reply.Hotels, hotel)
    }
  }
  return reply, nil
}

func inRange(hotelId int32, wantIds []int32) bool {
  for _, id := range wantIds {
    if id == hotelId {
      return true
    }
  }
  return false
}

// loadProfiles loads hotel profiles from a JSON file.
func (s *profileServer) loadProfiles(filePath string) {
  file, err := ioutil.ReadFile(filePath)
  if err != nil {
    log.Fatalf("Failed to load file: %v", err)
  }
  if err := json.Unmarshal(file, &s.hotels); err != nil {
    log.Fatalf("Failed to load json: %v", err)
  }
}

func newServer() *profileServer {
  s := new(profileServer)
  s.loadProfiles(*jsonDBFile)
  return s
}

func main() {
  flag.Parse()
  lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
  if err != nil {
    log.Fatalf("failed to listen: %v", err)
  }
  grpcServer := grpc.NewServer()
  pb.RegisterProfileServer(grpcServer, newServer())
  grpcServer.Serve(lis)
}
