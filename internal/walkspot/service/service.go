package walkspot_service

import (
	"context"
	"fmt"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type WalkSpotRepository interface {
	ListNearby(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.WalkSpot, error)
	ListByDistance(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.WalkSpotWithDistance, error)
	GetWalkSpot(ctx context.Context, id string) (core_domain.WalkSpot, error)
}

type PresenceService interface {
	CheckIn(ctx context.Context, spotId, petId string) (core_domain.PresenceEntry, error)
	CheckOut(ctx context.Context, spotId, petId string) error
	CountPresent(ctx context.Context, spotIds []string) (map[string]int, error)
	ListPresent(ctx context.Context, spotId string) ([]core_domain.PresenceEntry, error)
}

type PetService interface {
	EnsurePetOwner(ctx context.Context, ownerId string, petId string) error
	GetPetCards(ctx context.Context, petIds []string) (map[string]core_domain.PetCard, error)
}

type WalkSpotService struct {
	repository WalkSpotRepository
	presence   PresenceService
	pets       PetService
}

func NewWalkSpotService(repository WalkSpotRepository, presence PresenceService, pets PetService) *WalkSpotService {
	return &WalkSpotService{
		repository: repository,
		presence:   presence,
		pets:       pets,
	}
}

func (s *WalkSpotService) ListNearby(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.WalkSpotSummary, error) {
	spots, err := s.repository.ListNearby(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list nearby spots: %w", err)
	}

	spotIds := make([]string, len(spots))
	for i, spot := range spots {
		spotIds[i] = spot.Id
	}
	counts, err := s.presence.CountPresent(ctx, spotIds)
	if err != nil {
		return nil, fmt.Errorf("count present: %w", err)
	}

	summaries := make([]core_domain.WalkSpotSummary, len(spots))
	for i, spot := range spots {
		summaries[i] = core_domain.WalkSpotSummary{WalkSpot: spot, PresentCount: counts[spot.Id]}
	}
	return summaries, nil
}

// ListForPicker returns the spots for the "Иду гулять" spot picker: at most MaxSpotPickerRadiusM away,
// closest first (the repository sorts by distance).
func (s *WalkSpotService) ListForPicker(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.NearbyWalkSpotSummary, error) {
	if query.RadiusM <= 0 || query.RadiusM > core_domain.MaxSpotPickerRadiusM {
		return nil, fmt.Errorf("radius_m must be in (0, %d]: %w", core_domain.MaxSpotPickerRadiusM, core_errors.ErrInvalidArgument)
	}

	spots, err := s.repository.ListByDistance(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list spots by distance: %w", err)
	}

	spotIds := make([]string, len(spots))
	for i, spot := range spots {
		spotIds[i] = spot.Id
	}
	counts, err := s.presence.CountPresent(ctx, spotIds)
	if err != nil {
		return nil, fmt.Errorf("count present: %w", err)
	}

	summaries := make([]core_domain.NearbyWalkSpotSummary, len(spots))
	for i, spot := range spots {
		summaries[i] = core_domain.NearbyWalkSpotSummary{
			WalkSpotSummary: core_domain.WalkSpotSummary{WalkSpot: spot.WalkSpot, PresentCount: counts[spot.Id]},
			DistanceM:       spot.DistanceM,
		}
	}
	return summaries, nil
}

func (s *WalkSpotService) GetWalkSpot(ctx context.Context, id string) (core_domain.WalkSpotDetails, error) {
	spot, err := s.repository.GetWalkSpot(ctx, id)
	if err != nil {
		return core_domain.WalkSpotDetails{}, fmt.Errorf("get walk spot: %w", err)
	}

	entries, err := s.presence.ListPresent(ctx, id)
	if err != nil {
		return core_domain.WalkSpotDetails{}, fmt.Errorf("list present: %w", err)
	}

	petIds := make([]string, len(entries))
	for i, entry := range entries {
		petIds[i] = entry.PetId
	}
	cards, err := s.pets.GetPetCards(ctx, petIds)
	if err != nil {
		return core_domain.WalkSpotDetails{}, fmt.Errorf("load present pets: %w", err)
	}

	present := make([]core_domain.PresentPet, 0, len(entries))
	for _, entry := range entries {
		card, ok := cards[entry.PetId]
		if !ok {
			continue // the pet was deleted while checked in; its presence expires with the TTL
		}
		present = append(present, core_domain.PresentPet{PetCard: card, CheckedInAt: entry.CheckedInAt})
	}

	return core_domain.WalkSpotDetails{WalkSpot: spot, Present: present}, nil
}

func (s *WalkSpotService) CheckIn(ctx context.Context, ownerId, spotId, petId string) (core_domain.PresenceEntry, error) {
	if err := s.ensureCanAct(ctx, ownerId, spotId, petId); err != nil {
		return core_domain.PresenceEntry{}, err
	}
	entry, err := s.presence.CheckIn(ctx, spotId, petId)
	if err != nil {
		return core_domain.PresenceEntry{}, fmt.Errorf("check in: %w", err)
	}
	return entry, nil
}

func (s *WalkSpotService) CheckOut(ctx context.Context, ownerId, spotId, petId string) error {
	if err := s.ensureCanAct(ctx, ownerId, spotId, petId); err != nil {
		return err
	}
	if err := s.presence.CheckOut(ctx, spotId, petId); err != nil {
		return fmt.Errorf("check out: %w", err)
	}
	return nil
}

// ensureCanAct checks that the spot exists and the pet belongs to the owner.
func (s *WalkSpotService) ensureCanAct(ctx context.Context, ownerId, spotId, petId string) error {
	if _, err := s.repository.GetWalkSpot(ctx, spotId); err != nil {
		return fmt.Errorf("get walk spot: %w", err)
	}
	if err := s.pets.EnsurePetOwner(ctx, ownerId, petId); err != nil {
		return fmt.Errorf("check pet owner: %w", err)
	}
	return nil
}
