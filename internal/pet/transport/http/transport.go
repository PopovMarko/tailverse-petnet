package pet_transport_http

import (
	"context"

	"github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type PetService interface {
	CreatePet(ctx context.Context, pet core_domain.Pet) (core_domain.Pet, error)
	GetPets(ctx context.Context) ([]core_domain.Pet, error)
	GetPet(ctx context.Context, key string) (core_domain.Pet, error)
}

type PetHttpHandler struct {
	petService PetService
}

func NewPetHttpHandler(petService PetService) *PetHttpHandler {
	return &PetHttpHandler{
		petService: petService,
	}
}

func NewPetsRouter(h *PetHttpHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.CreatePet)
	r.Get("/", h.GetPets)
	r.Get("/id", h.GetPet)

	return r
}
