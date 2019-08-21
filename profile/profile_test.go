package profile

import (
	"testing"

	pb "github.com/harlow/go-micro-services/profile/proto"
)

func TestGetProfile(t *testing.T) {
	s := &Server{
		profiles: map[string]*pb.Hotel{
			"1": &pb.Hotel{Id: "1", Name: "Cliff Hotel"},
		},
	}

	got := s.getProfile("1")

	if got.Name != "Cliff Hotel" {
		t.Fatalf("get profile error: expected %v, got %v", "Cliff Hotel", got.Name)
	}
}
