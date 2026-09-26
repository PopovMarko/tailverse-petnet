package pet_transport_http

import (
	"time"

	"github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type PetRequestDto struct {
	Name          string    `json:"name" validate:"required"`
	Breed         string    `json:"breed" validate:"required"`
	Species       string    `json:"species"`
	BirthDate     time.Time `json:"birth_date"`
	Age           int       `json:"age"`
	ApproxAddress string    `json:"approx_address" validate:"required"`
}

type PetResponseDto struct {
	Id             string    `json:"id"`
	OwnerId        string    `json:"owner_id"`
	Name           string    `json:"name"`
	Breed          string    `json:"breed"`
	Species        string    `json:"species"`
	BirthDate      time.Time `json:"birth_date"`
	ApproxLocation string    `json:"approx_location"`
	ApproxAddress  string    `json:"approx_address"`
	CreatedAt      time.Time `json:"created_at"`
}

func DtoToDomain(pet PetRequestDto) core_domain.Pet {
	return core_domain.NewUninitializedPet(
		pet.Name,
		pet.Breed,
		pet.Species,
		pet.BirthDate,
		pet.ApproxAddress,
	)
}

func DomainToDto(pet core_domain.Pet) PetResponseDto {
	return PetResponseDto{
		Id:             pet.Id,
		OwnerId:        pet.OwnerId,
		Name:           pet.Name,
		Breed:          pet.Breed,
		Species:        pet.Species,
		BirthDate:      pet.BirthDate,
		ApproxLocation: pet.ApproxLocation,
		ApproxAddress:  pet.ApproxAddress,
		CreatedAt:      pet.CreatedAt,
	}
}
