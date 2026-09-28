package core_domain

import "time"

var (
	UninitializedString = ""
)

type Pet struct {
	Id            string
	OwnerId       string
	Name          string
	Breed         string
	Species       string
	BirthDate     time.Time // zero value means the birth date is unknown
	ApproxAddress string
	CreatedAt     time.Time
}

func NewPet(
	id string,
	ownerId string,
	name string,
	breed string,
	species string,
	birthDate time.Time,
	approxAddress string,
	createdAt time.Time,
) Pet {
	return Pet{
		Id:            id,
		OwnerId:       ownerId,
		Name:          name,
		Breed:         breed,
		Species:       species,
		BirthDate:     birthDate,
		ApproxAddress: approxAddress,
		CreatedAt:     createdAt,
	}
}

func NewUninitializedPet(
	name string,
	breed string,
	species string,
	birthDate time.Time,
	approxAddress string,
) Pet {
	return NewPet(
		UninitializedString,
		UninitializedString,
		name,
		breed,
		species,
		birthDate,
		approxAddress,
		time.Now(),
	)
}

// Age returns the pet's age in full years at the moment now, or nil when the birth date is unknown.
func (p Pet) Age(now time.Time) *int {
	if p.BirthDate.IsZero() {
		return nil
	}
	years := now.Year() - p.BirthDate.Year()
	if now.Month() < p.BirthDate.Month() || (now.Month() == p.BirthDate.Month() && now.Day() < p.BirthDate.Day()) {
		years--
	}
	if years < 0 {
		years = 0
	}
	return &years
}

// PetPatch holds the fields of PATCH /pets/{id}; nil means "leave unchanged".
type PetPatch struct {
	Name          *string
	Breed         *string
	Species       *string
	BirthDate     *time.Time
	ApproxAddress *string
}

func (p Pet) Apply(patch PetPatch) Pet {
	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Breed != nil {
		p.Breed = *patch.Breed
	}
	if patch.Species != nil {
		p.Species = *patch.Species
	}
	if patch.BirthDate != nil {
		p.BirthDate = *patch.BirthDate
	}
	if patch.ApproxAddress != nil {
		p.ApproxAddress = *patch.ApproxAddress
	}
	return p
}

// PetCard is the short pet+owner view shown in "who is at the spot" and announcement participant lists.
type PetCard struct {
	PetId         string
	PetName       string
	OwnerNickname string
}
