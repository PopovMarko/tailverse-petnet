package owner_transport_http

import (
	"context"
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type OwnerService interface {
	GetOwner(ctx context.Context, id string) (core_domain.Owner, error)
	GetPublicOwner(ctx context.Context, id string) (core_domain.PublicOwner, error)
	UpdateOwner(ctx context.Context, id string, patch core_domain.OwnerPatch) (core_domain.Owner, error)
}

type OwnerHttpHandler struct {
	ownerService OwnerService
}

func NewOwnerHttpHandler(ownerService OwnerService) *OwnerHttpHandler {
	return &OwnerHttpHandler{
		ownerService: ownerService,
	}
}

// NewOwnersRouter registers the /owners routes. "/me" is a static route, so chi matches it before "/{id}".
func NewOwnersRouter(h *OwnerHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/me", h.GetMe)
		r.Patch("/me", h.UpdateMe)
	})

	r.Get("/{id}", h.GetOwner)

	return r
}
