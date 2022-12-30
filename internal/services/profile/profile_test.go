package profile

import (
	"testing"

	profile "github.com/harlow/go-micro-services/internal/services/profile/proto"
)

func TestGetProfile(t *testing.T) {
	s := &Profile{
		profiles: map[string]*profile.Hotel{
			"1": {Id: "1", Name: "Cliff Hotel"},
		},
	}

	got := s.getProfile("1")

	if got.Name != "Cliff Hotel" {
		t.Fatalf("get profile error: expected %v, got %v", "Cliff Hotel", got.Name)
	}
}
