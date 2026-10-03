package walkspot_transport_http

import (
	"context"
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type WalkSpotService interface {
	ListNearby(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.WalkSpotSummary, error)
	ListForPicker(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.NearbyWalkSpotSummary, error)
	GetWalkSpot(ctx context.Context, id string) (core_domain.WalkSpotDetails, error)
	CheckIn(ctx context.Context, ownerId, spotId, petId string) (core_domain.PresenceEntry, error)
	CheckOut(ctx context.Context, ownerId, spotId, petId string) error
}

type WalkSpotHttpHandler struct {
	walkSpotService WalkSpotService
}

func NewWalkSpotHttpHandler(walkSpotService WalkSpotService) *WalkSpotHttpHandler {
	return &WalkSpotHttpHandler{
		walkSpotService: walkSpotService,
	}
}

func NewWalkSpotsRouter(h *WalkSpotHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", h.ListWalkSpots)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		// Registered before "/{id}" for readability; chi always prefers the static "/nearby" over the "{id}" pattern.
		r.Get("/nearby", h.ListNearbyWalkSpots)
		r.Post("/{id}/checkin", h.CheckIn)
		r.Delete("/{id}/checkin", h.CheckOut)
	})

	r.Get("/{id}", h.GetWalkSpot)

	return r
}
