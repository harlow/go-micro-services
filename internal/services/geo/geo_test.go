package geo

import "testing"

func TestGetNearbyPointsSortedByDistance(t *testing.T) {
	s := &Geo{points: []*point{
		{Pid: "a", Plat: 37.7750, Plon: -122.4195},
		{Pid: "b", Plat: 37.7850, Plon: -122.4095},
		{Pid: "c", Plat: 37.8050, Plon: -122.3895},
	}}

	got := s.getNearbyPoints(37.7749, -122.4194)
	if len(got) < 2 {
		t.Fatalf("expected at least 2 nearby points, got %d", len(got))
	}
	if got[0].Pid != "a" {
		t.Fatalf("expected closest point 'a', got %q", got[0].Pid)
	}
	if got[1].Pid != "b" {
		t.Fatalf("expected second closest point 'b', got %q", got[1].Pid)
	}
}
