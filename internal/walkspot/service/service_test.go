package walkspot_service

import (
	"context"
	"errors"
	"testing"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type fakeWalkSpotRepository struct {
	byDistance []core_domain.WalkSpotWithDistance
	calls      int
}

func (f *fakeWalkSpotRepository) ListNearby(context.Context, core_domain.NearbyQuery) ([]core_domain.WalkSpot, error) {
	return nil, nil
}

func (f *fakeWalkSpotRepository) ListByDistance(context.Context, core_domain.NearbyQuery) ([]core_domain.WalkSpotWithDistance, error) {
	f.calls++
	return f.byDistance, nil
}

func (f *fakeWalkSpotRepository) GetWalkSpot(context.Context, string) (core_domain.WalkSpot, error) {
	return core_domain.WalkSpot{}, core_errors.ErrNotFound
}

type fakePresence struct {
	counts map[string]int
}

func (f *fakePresence) CheckIn(context.Context, string, string) (core_domain.PresenceEntry, error) {
	return core_domain.PresenceEntry{}, nil
}
func (f *fakePresence) CheckOut(context.Context, string, string) error { return nil }
func (f *fakePresence) CountPresent(context.Context, []string) (map[string]int, error) {
	return f.counts, nil
}
func (f *fakePresence) ListPresent(context.Context, string) ([]core_domain.PresenceEntry, error) {
	return nil, nil
}

func TestListForPickerAddsPresentCountAndKeepsDistanceOrder(t *testing.T) {
	repository := &fakeWalkSpotRepository{byDistance: []core_domain.WalkSpotWithDistance{
		{WalkSpot: core_domain.WalkSpot{Id: "near", Name: "Near"}, DistanceM: 12},
		{WalkSpot: core_domain.WalkSpot{Id: "far", Name: "Far"}, DistanceM: 480},
	}}
	service := NewWalkSpotService(repository, &fakePresence{counts: map[string]int{"far": 3}}, nil)

	spots, err := service.ListForPicker(context.Background(), core_domain.NearbyQuery{RadiusM: 500})
	if err != nil {
		t.Fatalf("ListForPicker: %v", err)
	}
	if len(spots) != 2 || spots[0].Id != "near" || spots[1].Id != "far" {
		t.Fatalf("spots = %+v, want near then far", spots)
	}
	if spots[0].DistanceM != 12 || spots[1].DistanceM != 480 {
		t.Errorf("distances = %d, %d", spots[0].DistanceM, spots[1].DistanceM)
	}
	if spots[0].PresentCount != 0 || spots[1].PresentCount != 3 {
		t.Errorf("present counts = %d, %d; want 0, 3", spots[0].PresentCount, spots[1].PresentCount)
	}
}

func TestListForPickerRejectsRadiusAboveCap(t *testing.T) {
	repository := &fakeWalkSpotRepository{}
	service := NewWalkSpotService(repository, &fakePresence{}, nil)

	for _, radius := range []int{0, -1, core_domain.MaxSpotPickerRadiusM + 1, core_domain.DefaultRadiusM} {
		if _, err := service.ListForPicker(context.Background(), core_domain.NearbyQuery{RadiusM: radius}); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("radius %d: err = %v, want ErrInvalidArgument", radius, err)
		}
	}
	if repository.calls != 0 {
		t.Errorf("repository called %d times for invalid radii", repository.calls)
	}
}
