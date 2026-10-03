package announcement_service

import (
	"context"
	"fmt"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

const (
	MaxDurationMin = 12 * 60
	// A walk may be announced a little after it has started, but not long after.
	maxStartInPast   = time.Hour
	maxStartInFuture = 30 * 24 * time.Hour
)

type AnnouncementRepository interface {
	CreateAnnouncement(ctx context.Context, a core_domain.Announcement) (core_domain.Announcement, error)
	GetAnnouncement(ctx context.Context, id string) (core_domain.Announcement, error)
	ListAnnouncements(ctx context.Context, filter core_domain.AnnouncementFilter, from time.Time) ([]core_domain.Announcement, error)
	ListParticipants(ctx context.Context, announcementId string) ([]core_domain.Participant, error)
	AddParticipant(ctx context.Context, announcementId, petId string) (core_domain.Participation, error)
	RemoveParticipant(ctx context.Context, announcementId, petId string) error
}

type PetService interface {
	EnsurePetOwner(ctx context.Context, ownerId string, petId string) error
}

type EventPublisher interface {
	Publish(event any)
}

type AnnouncementService struct {
	repository AnnouncementRepository
	pets       PetService
	publisher  EventPublisher
	now        func() time.Time
}

func NewAnnouncementService(repository AnnouncementRepository, pets PetService, publisher EventPublisher) *AnnouncementService {
	return &AnnouncementService{
		repository: repository,
		pets:       pets,
		publisher:  publisher,
		now:        time.Now,
	}
}

func (s *AnnouncementService) CreateAnnouncement(ctx context.Context, ownerId string, a core_domain.Announcement) (core_domain.Announcement, error) {
	if err := s.validate(a); err != nil {
		return core_domain.Announcement{}, err
	}
	if err := s.pets.EnsurePetOwner(ctx, ownerId, a.PetId); err != nil {
		return core_domain.Announcement{}, fmt.Errorf("check pet owner: %w", err)
	}

	a.Status = core_domain.AnnouncementStatusActive
	created, err := s.repository.CreateAnnouncement(ctx, a)
	if err != nil {
		return core_domain.Announcement{}, fmt.Errorf("create announcement: %w", err)
	}

	s.publisher.Publish(core_domain.AnnouncementCreatedEvent{Announcement: created})
	return created, nil
}

func (s *AnnouncementService) ListAnnouncements(ctx context.Context, filter core_domain.AnnouncementFilter) ([]core_domain.Announcement, error) {
	from := s.now()
	if filter.From != nil {
		from = *filter.From
	}
	if filter.To != nil && filter.To.Before(from) {
		return nil, fmt.Errorf("to is before from: %w", core_errors.ErrInvalidArgument)
	}

	announcements, err := s.repository.ListAnnouncements(ctx, filter, from)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	return announcements, nil
}

func (s *AnnouncementService) GetAnnouncement(ctx context.Context, id string) (core_domain.AnnouncementDetails, error) {
	announcement, err := s.repository.GetAnnouncement(ctx, id)
	if err != nil {
		return core_domain.AnnouncementDetails{}, fmt.Errorf("get announcement: %w", err)
	}
	participants, err := s.repository.ListParticipants(ctx, id)
	if err != nil {
		return core_domain.AnnouncementDetails{}, fmt.Errorf("list participants: %w", err)
	}
	return core_domain.AnnouncementDetails{Announcement: announcement, Participants: participants}, nil
}

func (s *AnnouncementService) Join(ctx context.Context, ownerId, announcementId, petId string) (core_domain.Participation, error) {
	if err := s.pets.EnsurePetOwner(ctx, ownerId, petId); err != nil {
		return core_domain.Participation{}, fmt.Errorf("check pet owner: %w", err)
	}

	announcement, err := s.repository.GetAnnouncement(ctx, announcementId)
	if err != nil {
		return core_domain.Participation{}, fmt.Errorf("get announcement: %w", err)
	}
	if announcement.PetId == petId {
		return core_domain.Participation{}, fmt.Errorf("the pet already leads this walk: %w", core_errors.ErrInvalidArgument)
	}
	if announcement.Status != core_domain.AnnouncementStatusActive {
		return core_domain.Participation{}, fmt.Errorf("announcement is %s: %w", announcement.Status, core_errors.ErrInvalidArgument)
	}
	if announcement.StartsAt.Add(time.Duration(announcement.DurationMin) * time.Minute).Before(s.now()) {
		return core_domain.Participation{}, fmt.Errorf("the walk has already ended: %w", core_errors.ErrInvalidArgument)
	}

	participation, err := s.repository.AddParticipant(ctx, announcementId, petId)
	if err != nil {
		return core_domain.Participation{}, fmt.Errorf("join announcement: %w", err)
	}
	return participation, nil
}

func (s *AnnouncementService) Leave(ctx context.Context, ownerId, announcementId, petId string) error {
	if err := s.pets.EnsurePetOwner(ctx, ownerId, petId); err != nil {
		return fmt.Errorf("check pet owner: %w", err)
	}
	if err := s.repository.RemoveParticipant(ctx, announcementId, petId); err != nil {
		return fmt.Errorf("leave announcement: %w", err)
	}
	return nil
}

func (s *AnnouncementService) validate(a core_domain.Announcement) error {
	if (a.SpotId == nil) == (a.CustomPoint == nil) {
		return fmt.Errorf("exactly one of spot_id and custom_point is required: %w", core_errors.ErrInvalidArgument)
	}
	if a.CustomPoint != nil {
		if err := a.CustomPoint.Validate(); err != nil {
			return err
		}
	}
	if a.DurationMin <= 0 || a.DurationMin > MaxDurationMin {
		return fmt.Errorf("duration_min must be in [1, %d]: %w", MaxDurationMin, core_errors.ErrInvalidArgument)
	}

	now := s.now()
	if a.StartsAt.Before(now.Add(-maxStartInPast)) || a.StartsAt.After(now.Add(maxStartInFuture)) {
		return fmt.Errorf("starts_at must be between 1 hour ago and 30 days ahead: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
