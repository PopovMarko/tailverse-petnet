package pet_transport_http

import (
	"context"
	"net/http"

	"github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type PetService interface {
	CreatePet(ctx context.Context, ownerId string, pet core_domain.Pet) (core_domain.Pet, error)
	GetPets(ctx context.Context, ownerId string) ([]core_domain.Pet, error)
	GetPet(ctx context.Context, id string) (core_domain.Pet, error)
	UpdatePet(ctx context.Context, ownerId string, id string, patch core_domain.PetPatch) (core_domain.Pet, error)
	DeletePet(ctx context.Context, ownerId string, id string) error
}

type PetHttpHandler struct {
	petService PetService
}

func NewPetHttpHandler(petService PetService) *PetHttpHandler {
	return &PetHttpHandler{
		petService: petService,
	}
}

// NewPetsRouter registers the /pets routes. authMiddleware guards the routes that need the current owner.
func NewPetsRouter(h *PetHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/{id}", h.GetPet)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", h.CreatePet)
		r.Get("/", h.GetPets)
		r.Patch("/{id}", h.UpdatePet)
		r.Delete("/{id}", h.DeletePet)
	})

	return r
}
