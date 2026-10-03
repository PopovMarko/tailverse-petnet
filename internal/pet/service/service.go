package pet_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type PetRepository interface {
	CreatePet(ctx context.Context, pet core_domain.Pet) (core_domain.Pet, error)
	GetPet(ctx context.Context, id string) (core_domain.Pet, error)
	GetPetsByOwner(ctx context.Context, ownerId string) ([]core_domain.Pet, error)
	UpdatePet(ctx context.Context, pet core_domain.Pet) (core_domain.Pet, error)
	DeletePet(ctx context.Context, id string) error
	GetPetCards(ctx context.Context, petIds []string) (map[string]core_domain.PetCard, error)
}

type PetService struct {
	repository PetRepository
	now        func() time.Time
}

func NewPetService(repository PetRepository) *PetService {
	return &PetService{
		repository: repository,
		now:        time.Now,
	}
}

func (s *PetService) CreatePet(ctx context.Context, ownerId string, pet core_domain.Pet) (core_domain.Pet, error) {
	pet.OwnerId = ownerId
	pet = normalize(pet)
	if pet.Species == "" {
		pet.Species = SpeciesByBreed(pet.Breed)
	}
	if err := s.validate(pet); err != nil {
		return core_domain.Pet{}, err
	}

	created, err := s.repository.CreatePet(ctx, pet)
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("create pet: %w", err)
	}
	return created, nil
}

func (s *PetService) GetPet(ctx context.Context, id string) (core_domain.Pet, error) {
	pet, err := s.repository.GetPet(ctx, id)
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("get pet: %w", err)
	}
	return pet, nil
}

func (s *PetService) GetPets(ctx context.Context, ownerId string) ([]core_domain.Pet, error) {
	pets, err := s.repository.GetPetsByOwner(ctx, ownerId)
	if err != nil {
		return nil, fmt.Errorf("get pets: %w", err)
	}
	return pets, nil
}

func (s *PetService) UpdatePet(ctx context.Context, ownerId string, id string, patch core_domain.PetPatch) (core_domain.Pet, error) {
	pet, err := s.ownedPet(ctx, ownerId, id)
	if err != nil {
		return core_domain.Pet{}, err
	}

	updated := normalize(pet.Apply(patch))
	// A new breed may imply a new species, unless the owner set the species explicitly.
	if patch.Breed != nil && patch.Species == nil {
		if species := SpeciesByBreed(updated.Breed); species != "" {
			updated.Species = species
		}
	}
	if err := s.validate(updated); err != nil {
		return core_domain.Pet{}, err
	}

	saved, err := s.repository.UpdatePet(ctx, updated)
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("update pet: %w", err)
	}
	return saved, nil
}

func (s *PetService) DeletePet(ctx context.Context, ownerId string, id string) error {
	if _, err := s.ownedPet(ctx, ownerId, id); err != nil {
		return err
	}
	if err := s.repository.DeletePet(ctx, id); err != nil {
		return fmt.Errorf("delete pet: %w", err)
	}
	return nil
}

// EnsurePetOwner returns ErrNotFound when the pet does not exist and ErrForbidden when it belongs to someone else.
// Other domains (walk spots, announcements, feed) use it to check that an owner acts on behalf of their own pet.
func (s *PetService) EnsurePetOwner(ctx context.Context, ownerId string, petId string) error {
	_, err := s.ownedPet(ctx, ownerId, petId)
	return err
}

// GetPetCards batch-loads pet name + owner nickname, e.g. for pets present at a walk spot.
func (s *PetService) GetPetCards(ctx context.Context, petIds []string) (map[string]core_domain.PetCard, error) {
	cards, err := s.repository.GetPetCards(ctx, petIds)
	if err != nil {
		return nil, fmt.Errorf("get pet cards: %w", err)
	}
	return cards, nil
}

func (s *PetService) ownedPet(ctx context.Context, ownerId string, id string) (core_domain.Pet, error) {
	pet, err := s.repository.GetPet(ctx, id)
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("get pet: %w", err)
	}
	if pet.OwnerId != ownerId {
		return core_domain.Pet{}, fmt.Errorf("pet %s belongs to another owner: %w", id, core_errors.ErrForbidden)
	}
	return pet, nil
}

func (s *PetService) validate(pet core_domain.Pet) error {
	if pet.Name == "" {
		return fmt.Errorf("name is required: %w", core_errors.ErrInvalidArgument)
	}
	// Required on create and kept non-empty by PATCH: a blank value would clear them.
	if pet.Breed == "" {
		return fmt.Errorf("breed is required: %w", core_errors.ErrInvalidArgument)
	}
	if pet.ApproxAddress == "" {
		return fmt.Errorf("approx_address is required: %w", core_errors.ErrInvalidArgument)
	}
	if pet.Species == "" {
		return fmt.Errorf("species can not be derived from breed %q, send it explicitly: %w", pet.Breed, core_errors.ErrInvalidArgument)
	}
	if !pet.BirthDate.IsZero() && pet.BirthDate.After(s.now()) {
		return fmt.Errorf("birth_date is in the future: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}

func normalize(pet core_domain.Pet) core_domain.Pet {
	pet.Name = strings.TrimSpace(pet.Name)
	pet.Breed = strings.TrimSpace(pet.Breed)
	pet.Species = strings.ToLower(strings.TrimSpace(pet.Species))
	pet.ApproxAddress = strings.TrimSpace(pet.ApproxAddress)
	return pet
}
