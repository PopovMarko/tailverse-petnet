package owner_service

import (
	"context"
	"fmt"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type OwnerRepository interface {
	GetOwner(ctx context.Context, id string) (core_domain.Owner, error)
	UpdateProfile(ctx context.Context, owner core_domain.Owner) (core_domain.Owner, error)
}

type OwnerService struct {
	repository OwnerRepository
}

func NewOwnerService(repository OwnerRepository) *OwnerService {
	return &OwnerService{repository: repository}
}

// GetOwner returns the full profile; only the owner themselves may see it (GET /owners/me).
func (s *OwnerService) GetOwner(ctx context.Context, id string) (core_domain.Owner, error) {
	owner, err := s.repository.GetOwner(ctx, id)
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("get owner: %w", err)
	}
	return owner, nil
}

// GetPublicOwner returns the profile as other users see it: hidden fields are left out, the email is never included.
func (s *OwnerService) GetPublicOwner(ctx context.Context, id string) (core_domain.PublicOwner, error) {
	owner, err := s.repository.GetOwner(ctx, id)
	if err != nil {
		return core_domain.PublicOwner{}, fmt.Errorf("get owner: %w", err)
	}
	return owner.Public(), nil
}

func (s *OwnerService) UpdateOwner(ctx context.Context, id string, patch core_domain.OwnerPatch) (core_domain.Owner, error) {
	owner, err := s.repository.GetOwner(ctx, id)
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("get owner: %w", err)
	}

	updated := owner.Apply(patch).NormalizeProfile()
	if err := updated.ValidateProfile(); err != nil {
		return core_domain.Owner{}, err
	}

	saved, err := s.repository.UpdateProfile(ctx, updated)
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("update owner: %w", err)
	}
	return saved, nil
}
