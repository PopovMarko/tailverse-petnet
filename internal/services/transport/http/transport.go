package services_transport_http

import (
	"context"
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type ServiceOfferService interface {
	CreateServiceOffer(ctx context.Context, ownerId string, offer core_domain.ServiceOffer) (core_domain.ServiceOffer, error)
	GetServiceOffer(ctx context.Context, id string) (core_domain.ServiceOffer, error)
	ListNearby(ctx context.Context, filter core_domain.ServiceOfferFilter) ([]core_domain.ServiceOffer, error)
}

type ServicesHttpHandler struct {
	serviceOfferService ServiceOfferService
}

func NewServicesHttpHandler(serviceOfferService ServiceOfferService) *ServicesHttpHandler {
	return &ServicesHttpHandler{
		serviceOfferService: serviceOfferService,
	}
}

func NewServicesRouter(h *ServicesHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", h.ListServices)
	r.Get("/{id}", h.GetService)
	r.With(authMiddleware).Post("/", h.CreateService)

	return r
}
