package announcement_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

var now = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

type fakeRepository struct {
	announcements map[string]core_domain.Announcement
	participants  map[string]bool
}

func newFakeRepository(announcements ...core_domain.Announcement) *fakeRepository {
	repository := &fakeRepository{announcements: map[string]core_domain.Announcement{}, participants: map[string]bool{}}
	for _, a := range announcements {
		repository.announcements[a.Id] = a
	}
	return repository
}

func (f *fakeRepository) CreateAnnouncement(_ context.Context, a core_domain.Announcement) (core_domain.Announcement, error) {
	a.Id = "ann-new"
	f.announcements[a.Id] = a
	return a, nil
}

func (f *fakeRepository) GetAnnouncement(_ context.Context, id string) (core_domain.Announcement, error) {
	a, ok := f.announcements[id]
	if !ok {
		return core_domain.Announcement{}, core_errors.ErrNotFound
	}
	return a, nil
}

func (f *fakeRepository) ListAnnouncements(context.Context, core_domain.AnnouncementFilter, time.Time) ([]core_domain.Announcement, error) {
	return nil, nil
}

func (f *fakeRepository) ListParticipants(context.Context, string) ([]core_domain.Participant, error) {
	return nil, nil
}

func (f *fakeRepository) AddParticipant(_ context.Context, announcementId, petId string) (core_domain.Participation, error) {
	key := announcementId + "/" + petId
	if f.participants[key] {
		return core_domain.Participation{}, core_errors.ErrConflict
	}
	f.participants[key] = true
	return core_domain.Participation{AnnouncementId: announcementId, PetId: petId, JoinedAt: now}, nil
}

func (f *fakeRepository) RemoveParticipant(context.Context, string, string) error { return nil }

// fakePets: every pet is owned by the owner with the same suffix, e.g. pet-1 → owner-1.
type fakePets struct{}

func (fakePets) EnsurePetOwner(_ context.Context, ownerId string, petId string) error {
	if petId == "pet-missing" {
		return core_errors.ErrNotFound
	}
	if "owner"+petId[len("pet"):] != ownerId {
		return core_errors.ErrForbidden
	}
	return nil
}

type recordingPublisher struct{ events []any }

func (p *recordingPublisher) Publish(event any) { p.events = append(p.events, event) }

func newService(repository *fakeRepository, publisher *recordingPublisher) *AnnouncementService {
	service := NewAnnouncementService(repository, fakePets{}, publisher)
	service.now = func() time.Time { return now }
	return service
}

func TestCreateAnnouncementValidation(t *testing.T) {
	spot := new("spot-1")
	point := &core_domain.GeoPoint{Lat: 55.7, Lng: 37.6}

	cases := []struct {
		name string
		a    core_domain.Announcement
		want error
	}{
		{"spot", core_domain.Announcement{PetId: "pet-1", SpotId: spot, StartsAt: now.Add(time.Hour), DurationMin: 60}, nil},
		{"custom point", core_domain.Announcement{PetId: "pet-1", CustomPoint: point, StartsAt: now, DurationMin: 30}, nil},
		{"no place", core_domain.Announcement{PetId: "pet-1", StartsAt: now, DurationMin: 30}, core_errors.ErrInvalidArgument},
		{"both places", core_domain.Announcement{PetId: "pet-1", SpotId: spot, CustomPoint: point, StartsAt: now, DurationMin: 30}, core_errors.ErrInvalidArgument},
		{"bad point", core_domain.Announcement{PetId: "pet-1", CustomPoint: &core_domain.GeoPoint{Lat: 91}, StartsAt: now, DurationMin: 30}, core_errors.ErrInvalidArgument},
		{"zero duration", core_domain.Announcement{PetId: "pet-1", SpotId: spot, StartsAt: now, DurationMin: 0}, core_errors.ErrInvalidArgument},
		{"too long", core_domain.Announcement{PetId: "pet-1", SpotId: spot, StartsAt: now, DurationMin: MaxDurationMin + 1}, core_errors.ErrInvalidArgument},
		{"long ago", core_domain.Announcement{PetId: "pet-1", SpotId: spot, StartsAt: now.Add(-2 * time.Hour), DurationMin: 30}, core_errors.ErrInvalidArgument},
		{"far future", core_domain.Announcement{PetId: "pet-1", SpotId: spot, StartsAt: now.AddDate(0, 2, 0), DurationMin: 30}, core_errors.ErrInvalidArgument},
		{"someone else's pet", core_domain.Announcement{PetId: "pet-2", SpotId: spot, StartsAt: now, DurationMin: 30}, core_errors.ErrForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			publisher := &recordingPublisher{}
			created, err := newService(newFakeRepository(), publisher).CreateAnnouncement(context.Background(), "owner-1", tc.a)

			if tc.want == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if created.Status != core_domain.AnnouncementStatusActive {
					t.Errorf("status = %q, want active", created.Status)
				}
				if len(publisher.events) != 1 {
					t.Errorf("published %d events, want 1 announcement_created", len(publisher.events))
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if len(publisher.events) != 0 {
				t.Error("an event was published for a rejected announcement")
			}
		})
	}
}

func TestJoinAnnouncement(t *testing.T) {
	active := core_domain.Announcement{Id: "ann-1", PetId: "pet-9", SpotId: new("spot-1"), StartsAt: now, DurationMin: 60, Status: core_domain.AnnouncementStatusActive}
	ended := core_domain.Announcement{Id: "ann-2", PetId: "pet-9", SpotId: new("spot-1"), StartsAt: now.Add(-3 * time.Hour), DurationMin: 60, Status: core_domain.AnnouncementStatusActive}
	cancelled := core_domain.Announcement{Id: "ann-3", PetId: "pet-9", SpotId: new("spot-1"), StartsAt: now, DurationMin: 60, Status: core_domain.AnnouncementStatusCancelled}
	service := newService(newFakeRepository(active, ended, cancelled), &recordingPublisher{})
	ctx := context.Background()

	if _, err := service.Join(ctx, "owner-1", "ann-1", "pet-1"); err != nil {
		t.Fatalf("first join: %v", err)
	}
	if _, err := service.Join(ctx, "owner-1", "ann-1", "pet-1"); !errors.Is(err, core_errors.ErrConflict) {
		t.Errorf("second join err = %v, want ErrConflict", err)
	}
	if _, err := service.Join(ctx, "owner-9", "ann-1", "pet-9"); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("author joining own walk err = %v, want ErrInvalidArgument", err)
	}
	if _, err := service.Join(ctx, "owner-1", "ann-2", "pet-1"); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("joining ended walk err = %v, want ErrInvalidArgument", err)
	}
	if _, err := service.Join(ctx, "owner-1", "ann-3", "pet-1"); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("joining cancelled walk err = %v, want ErrInvalidArgument", err)
	}
	if _, err := service.Join(ctx, "owner-1", "ann-404", "pet-1"); !errors.Is(err, core_errors.ErrNotFound) {
		t.Errorf("joining missing walk err = %v, want ErrNotFound", err)
	}
	if _, err := service.Join(ctx, "owner-1", "ann-1", "pet-2"); !errors.Is(err, core_errors.ErrForbidden) {
		t.Errorf("joining with someone else's pet err = %v, want ErrForbidden", err)
	}
}

func TestListAnnouncementsRejectsInvertedRange(t *testing.T) {
	service := newService(newFakeRepository(), &recordingPublisher{})
	from, to := now.Add(time.Hour), now

	_, err := service.ListAnnouncements(context.Background(), core_domain.AnnouncementFilter{From: &from, To: &to})
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("err = %v, want ErrInvalidArgument", err)
	}
}
