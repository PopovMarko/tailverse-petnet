package pet_transport_http

import (
	"time"

	"github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_http_types "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/types"
)

type PetRequestDto struct {
	Name          string                `json:"name" validate:"required,max=100"`
	Breed         string                `json:"breed" validate:"required,max=100"`
	Species       string                `json:"species" validate:"max=50"`
	BirthDate     *core_http_types.Date `json:"birth_date"`
	ApproxAddress string                `json:"approx_address" validate:"required,max=200"`
}

type PetResponseDto struct {
	Id            string                `json:"id"`
	OwnerId       string                `json:"owner_id"`
	Name          string                `json:"name"`
	Breed         string                `json:"breed"`
	Species       string                `json:"species"`
	BirthDate     *core_http_types.Date `json:"birth_date"`
	Age           *int                  `json:"age"`
	ApproxAddress string                `json:"approx_address"`
	CreatedAt     time.Time             `json:"created_at"`
}

type PetsListResponseDto struct {
	Pets []PetResponseDto `json:"pets"`
}

// PetUpdateDto is the PATCH body: every field is optional, absent fields are left unchanged.
type PetUpdateDto struct {
	Name          *string               `json:"name" validate:"omitempty,min=1,max=100"`
	Breed         *string               `json:"breed" validate:"omitempty,max=100"`
	Species       *string               `json:"species" validate:"omitempty,min=1,max=50"`
	BirthDate     *core_http_types.Date `json:"birth_date"`
	ApproxAddress *string               `json:"approx_address" validate:"omitempty,max=200"`
}

func DtoToDomain(pet PetRequestDto) core_domain.Pet {
	return core_domain.NewUninitializedPet(
		pet.Name,
		pet.Breed,
		pet.Species,
		pet.BirthDate.TimeOrZero(),
		pet.ApproxAddress,
	)
}

func UpdateDtoToDomain(patch PetUpdateDto) core_domain.PetPatch {
	domainPatch := core_domain.PetPatch{
		Name:          patch.Name,
		Breed:         patch.Breed,
		Species:       patch.Species,
		ApproxAddress: patch.ApproxAddress,
	}
	if patch.BirthDate != nil {
		domainPatch.BirthDate = &patch.BirthDate.Time
	}
	return domainPatch
}

func DomainToDto(pet core_domain.Pet) PetResponseDto {
	return PetResponseDto{
		Id:            pet.Id,
		OwnerId:       pet.OwnerId,
		Name:          pet.Name,
		Breed:         pet.Breed,
		Species:       pet.Species,
		BirthDate:     core_http_types.NewDate(pet.BirthDate),
		Age:           pet.Age(time.Now()),
		ApproxAddress: pet.ApproxAddress,
		CreatedAt:     pet.CreatedAt,
	}
}
