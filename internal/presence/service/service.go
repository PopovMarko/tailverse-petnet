package presence_service

import (
	"context"
	"fmt"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"go.uber.org/zap"
)

const DefaultTTL = 2 * time.Hour

type PresenceRepository interface {
	CheckIn(ctx context.Context, spotId, petId string, at time.Time, ttl time.Duration) (string, error)
	CheckOut(ctx context.Context, spotId, petId string) (bool, error)
	CountPresent(ctx context.Context, spotIds []string, now time.Time, ttl time.Duration) (map[string]int, error)
	ListPresent(ctx context.Context, spotId string, now time.Time, ttl time.Duration) ([]core_domain.PresenceEntry, error)
}

type EventPublisher interface {
	Publish(event any)
}

// PresenceService keeps the live "who is at the spot right now" state and pushes spot_update events.
// It trusts its callers: the walk spot service checks that the spot exists and the pet belongs to the owner.
type PresenceService struct {
	repository PresenceRepository
	publisher  EventPublisher
	ttl        time.Duration
	now        func() time.Time
}

func NewPresenceService(repository PresenceRepository, publisher EventPublisher, ttl time.Duration) *PresenceService {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &PresenceService{
		repository: repository,
		publisher:  publisher,
		ttl:        ttl,
		now:        time.Now,
	}
}

func (s *PresenceService) CheckIn(ctx context.Context, spotId, petId string) (core_domain.PresenceEntry, error) {
	now := s.now().UTC().Truncate(time.Millisecond)

	previousSpotId, err := s.repository.CheckIn(ctx, spotId, petId, now, s.ttl)
	if err != nil {
		return core_domain.PresenceEntry{}, fmt.Errorf("presence check in: %w", err)
	}

	s.publishCounts(ctx, spotId, previousSpotId)
	return core_domain.PresenceEntry{
		SpotId:      spotId,
		PetId:       petId,
		CheckedInAt: now,
		ExpiresAt:   now.Add(s.ttl),
	}, nil
}

func (s *PresenceService) CheckOut(ctx context.Context, spotId, petId string) error {
	removed, err := s.repository.CheckOut(ctx, spotId, petId)
	if err != nil {
		return fmt.Errorf("presence check out: %w", err)
	}
	if !removed {
		return fmt.Errorf("pet %s is not checked in at spot %s: %w", petId, spotId, core_errors.ErrNotFound)
	}

	s.publishCounts(ctx, spotId)
	return nil
}

func (s *PresenceService) CountPresent(ctx context.Context, spotIds []string) (map[string]int, error) {
	counts, err := s.repository.CountPresent(ctx, spotIds, s.now(), s.ttl)
	if err != nil {
		return nil, fmt.Errorf("presence count: %w", err)
	}
	return counts, nil
}

func (s *PresenceService) ListPresent(ctx context.Context, spotId string) ([]core_domain.PresenceEntry, error) {
	entries, err := s.repository.ListPresent(ctx, spotId, s.now(), s.ttl)
	if err != nil {
		return nil, fmt.Errorf("presence list: %w", err)
	}
	return entries, nil
}

// publishCounts sends spot_update for every changed spot. Failures are only logged:
// the check-in itself already succeeded, and clients re-read counts over REST anyway.
func (s *PresenceService) publishCounts(ctx context.Context, spotIds ...string) {
	changed := make([]string, 0, len(spotIds))
	for _, spotId := range spotIds {
		if spotId != "" {
			changed = append(changed, spotId)
		}
	}

	counts, err := s.repository.CountPresent(ctx, changed, s.now(), s.ttl)
	if err != nil {
		core_logger.FromContext(ctx).Warn("count presence for spot_update", zap.Error(err))
		return
	}
	for _, spotId := range changed {
		s.publisher.Publish(core_domain.SpotUpdatedEvent{SpotId: spotId, PresentCount: counts[spotId]})
	}
}
