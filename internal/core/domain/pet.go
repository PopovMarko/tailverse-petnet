package core_domain

import "time"

var (
	UninitializedString = ""
)

type Pet struct {
	Id             string
	OwnerId        string
	Name           string
	Breed          string
	Species        string
	BirthDate      time.Time
	ApproxLocation string
	ApproxAddress  string
	CreatedAt      time.Time
}

func NewPet(
	id string,
	ownerId string,
	name string,
	breed string,
	species string,
	birthDate time.Time,
	approxLocation string,
	approxAddress string,
	createdAt time.Time,
) Pet {
	return Pet{
		Id:             id,
		OwnerId:        ownerId,
		Name:           name,
		Breed:          breed,
		Species:        species,
		BirthDate:      birthDate,
		ApproxLocation: approxLocation,
		ApproxAddress:  approxAddress,
		CreatedAt:      createdAt,
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
		UninitializedString,
		approxAddress,
		time.Now(),
	)
}
