package services_service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

// Categories are free-form slugs such as "grooming" or "dog_walking" until the service model (partners vs P2P) is decided.
var categoryPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,49}$`)

type ServiceOfferRepository interface {
	CreateServiceOffer(ctx context.Context, offer core_domain.ServiceOffer) (core_domain.ServiceOffer, error)
	GetServiceOffer(ctx context.Context, id string) (core_domain.ServiceOffer, error)
	ListNearby(ctx context.Context, filter core_domain.ServiceOfferFilter) ([]core_domain.ServiceOffer, error)
}

type ServiceOfferService struct {
	repository ServiceOfferRepository
}

func NewServiceOfferService(repository ServiceOfferRepository) *ServiceOfferService {
	return &ServiceOfferService{repository: repository}
}

func (s *ServiceOfferService) CreateServiceOffer(ctx context.Context, ownerId string, offer core_domain.ServiceOffer) (core_domain.ServiceOffer, error) {
	offer.ProviderOwnerId = ownerId
	offer.Title = strings.TrimSpace(offer.Title)
	offer.Description = strings.TrimSpace(offer.Description)
	offer.Category = strings.ToLower(strings.TrimSpace(offer.Category))

	if offer.Title == "" {
		return core_domain.ServiceOffer{}, fmt.Errorf("title is required: %w", core_errors.ErrInvalidArgument)
	}
	if !categoryPattern.MatchString(offer.Category) {
		return core_domain.ServiceOffer{}, fmt.Errorf("category must be a slug like grooming or dog_walking: %w", core_errors.ErrInvalidArgument)
	}
	if offer.Price < 0 {
		return core_domain.ServiceOffer{}, fmt.Errorf("price can not be negative: %w", core_errors.ErrInvalidArgument)
	}
	if err := offer.Location.Validate(); err != nil {
		return core_domain.ServiceOffer{}, err
	}

	created, err := s.repository.CreateServiceOffer(ctx, offer)
	if err != nil {
		return core_domain.ServiceOffer{}, fmt.Errorf("create service: %w", err)
	}
	return created, nil
}

func (s *ServiceOfferService) GetServiceOffer(ctx context.Context, id string) (core_domain.ServiceOffer, error) {
	offer, err := s.repository.GetServiceOffer(ctx, id)
	if err != nil {
		return core_domain.ServiceOffer{}, fmt.Errorf("get service: %w", err)
	}
	return offer, nil
}

func (s *ServiceOfferService) ListNearby(ctx context.Context, filter core_domain.ServiceOfferFilter) ([]core_domain.ServiceOffer, error) {
	if filter.Category != nil {
		category := strings.ToLower(strings.TrimSpace(*filter.Category))
		filter.Category = &category
	}
	offers, err := s.repository.ListNearby(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	return offers, nil
}
