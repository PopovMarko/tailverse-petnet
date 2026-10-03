package pet_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type fakePetRepository struct {
	pets map[string]core_domain.Pet
}

func newFakePetRepository(pets ...core_domain.Pet) *fakePetRepository {
	repository := &fakePetRepository{pets: map[string]core_domain.Pet{}}
	for _, pet := range pets {
		repository.pets[pet.Id] = pet
	}
	return repository
}

func (f *fakePetRepository) CreatePet(_ context.Context, pet core_domain.Pet) (core_domain.Pet, error) {
	pet.Id = "new-pet"
	f.pets[pet.Id] = pet
	return pet, nil
}

func (f *fakePetRepository) GetPet(_ context.Context, id string) (core_domain.Pet, error) {
	pet, ok := f.pets[id]
	if !ok {
		return core_domain.Pet{}, core_errors.ErrNotFound
	}
	return pet, nil
}

func (f *fakePetRepository) GetPetsByOwner(_ context.Context, ownerId string) ([]core_domain.Pet, error) {
	var pets []core_domain.Pet
	for _, pet := range f.pets {
		if pet.OwnerId == ownerId {
			pets = append(pets, pet)
		}
	}
	return pets, nil
}

func (f *fakePetRepository) UpdatePet(_ context.Context, pet core_domain.Pet) (core_domain.Pet, error) {
	f.pets[pet.Id] = pet
	return pet, nil
}

func (f *fakePetRepository) DeletePet(_ context.Context, id string) error {
	delete(f.pets, id)
	return nil
}

func (f *fakePetRepository) GetPetCards(context.Context, []string) (map[string]core_domain.PetCard, error) {
	return map[string]core_domain.PetCard{}, nil
}

func TestCreatePetDerivesSpeciesFromBreed(t *testing.T) {
	service := NewPetService(newFakePetRepository())

	pet, err := service.CreatePet(context.Background(), "owner-1", core_domain.Pet{Name: " Rex ", Breed: "Golden Retriever", ApproxAddress: "Центр"})
	if err != nil {
		t.Fatalf("CreatePet: %v", err)
	}
	if pet.Species != "dog" {
		t.Errorf("species = %q, want dog", pet.Species)
	}
	if pet.OwnerId != "owner-1" {
		t.Errorf("owner = %q, want owner-1", pet.OwnerId)
	}
	if pet.Name != "Rex" {
		t.Errorf("name = %q, want trimmed Rex", pet.Name)
	}
}

func TestCreatePetRequiresSpeciesForUnknownBreed(t *testing.T) {
	service := NewPetService(newFakePetRepository())

	_, err := service.CreatePet(context.Background(), "owner-1", core_domain.Pet{Name: "Kesha", Breed: "Budgerigar", ApproxAddress: "Центр"})
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("err = %v, want ErrInvalidArgument", err)
	}

	pet, err := service.CreatePet(context.Background(), "owner-1", core_domain.Pet{Name: "Kesha", Breed: "Budgerigar", Species: "Bird", ApproxAddress: "Центр"})
	if err != nil {
		t.Fatalf("CreatePet with explicit species: %v", err)
	}
	if pet.Species != "bird" {
		t.Errorf("species = %q, want lower-cased bird", pet.Species)
	}
}

func TestCreatePetRejectsFutureBirthDate(t *testing.T) {
	service := NewPetService(newFakePetRepository())

	_, err := service.CreatePet(context.Background(), "owner-1", core_domain.Pet{
		Name: "Rex", Species: "dog", Breed: "mixed", ApproxAddress: "Центр", BirthDate: time.Now().AddDate(1, 0, 0),
	})
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestUpdatePetOnlyByOwner(t *testing.T) {
	repository := newFakePetRepository(core_domain.Pet{Id: "pet-1", OwnerId: "owner-1", Name: "Rex", Species: "dog"})
	service := NewPetService(repository)

	_, err := service.UpdatePet(context.Background(), "owner-2", "pet-1", core_domain.PetPatch{Name: new("Hacked")})
	if !errors.Is(err, core_errors.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
	if repository.pets["pet-1"].Name != "Rex" {
		t.Error("pet was changed by another owner")
	}

	_, err = service.UpdatePet(context.Background(), "owner-1", "missing", core_domain.PetPatch{})
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdatePetBreedChangesSpeciesUnlessExplicit(t *testing.T) {
	repository := newFakePetRepository(core_domain.Pet{Id: "pet-1", OwnerId: "owner-1", Name: "Murka", Species: "dog", Breed: "mixed", ApproxAddress: "Центр"})
	service := NewPetService(repository)

	updated, err := service.UpdatePet(context.Background(), "owner-1", "pet-1", core_domain.PetPatch{Breed: new("Maine Coon")})
	if err != nil {
		t.Fatalf("UpdatePet: %v", err)
	}
	if updated.Species != "cat" {
		t.Errorf("species = %q, want cat derived from the new breed", updated.Species)
	}

	updated, err = service.UpdatePet(context.Background(), "owner-1", "pet-1", core_domain.PetPatch{Breed: new("Siamese"), Species: new("ferret")})
	if err != nil {
		t.Fatalf("UpdatePet: %v", err)
	}
	if updated.Species != "ferret" {
		t.Errorf("species = %q, want the explicit ferret", updated.Species)
	}
}

func TestCreatePetRequiresBreedAndArea(t *testing.T) {
	service := NewPetService(newFakePetRepository())

	for name, pet := range map[string]core_domain.Pet{
		"blank breed":   {Name: "Rex", Species: "dog", Breed: "   ", ApproxAddress: "Центр"},
		"blank address": {Name: "Rex", Species: "dog", Breed: "mixed", ApproxAddress: " "},
	} {
		if _, err := service.CreatePet(context.Background(), "owner-1", pet); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("%s: err = %v, want ErrInvalidArgument", name, err)
		}
	}
}

func TestUpdatePetCanNotClearBreedOrArea(t *testing.T) {
	original := core_domain.Pet{Id: "pet-1", OwnerId: "owner-1", Name: "Rex", Species: "dog", Breed: "mixed", ApproxAddress: "Центр"}
	repository := newFakePetRepository(original)
	service := NewPetService(repository)

	for name, patch := range map[string]core_domain.PetPatch{
		"empty breed":   {Breed: new("")},
		"blank breed":   {Breed: new("  ")},
		"empty address": {ApproxAddress: new("")},
		"blank name":    {Name: new(" ")},
	} {
		if _, err := service.UpdatePet(context.Background(), "owner-1", "pet-1", patch); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("%s: err = %v, want ErrInvalidArgument", name, err)
		}
	}
	if repository.pets["pet-1"] != original {
		t.Errorf("pet was changed by a rejected patch: %+v", repository.pets["pet-1"])
	}
}

func TestUpdatePetSetsAndClearsBirthDate(t *testing.T) {
	repository := newFakePetRepository(core_domain.Pet{Id: "pet-1", OwnerId: "owner-1", Name: "Rex", Species: "dog", Breed: "mixed", ApproxAddress: "Центр"})
	service := NewPetService(repository)
	birthDate := time.Date(2021, 4, 10, 0, 0, 0, 0, time.UTC)

	updated, err := service.UpdatePet(context.Background(), "owner-1", "pet-1", core_domain.PetPatch{BirthDate: &birthDate})
	if err != nil {
		t.Fatalf("set birth date: %v", err)
	}
	if !updated.BirthDate.Equal(birthDate) {
		t.Fatalf("birth date = %v, want %v", updated.BirthDate, birthDate)
	}

	updated, err = service.UpdatePet(context.Background(), "owner-1", "pet-1", core_domain.PetPatch{Name: new("Rex II")})
	if err != nil {
		t.Fatalf("patch without birth date: %v", err)
	}
	if !updated.BirthDate.Equal(birthDate) {
		t.Errorf("a patch without birth_date changed it to %v", updated.BirthDate)
	}

	updated, err = service.UpdatePet(context.Background(), "owner-1", "pet-1", core_domain.PetPatch{BirthDate: &time.Time{}})
	if err != nil {
		t.Fatalf("clear birth date: %v", err)
	}
	if !updated.BirthDate.IsZero() || updated.Age(time.Now()) != nil {
		t.Errorf("birth date = %v, want it removed", updated.BirthDate)
	}
}

func TestDeletePetForbiddenForOtherOwner(t *testing.T) {
	repository := newFakePetRepository(core_domain.Pet{Id: "pet-1", OwnerId: "owner-1", Name: "Rex", Species: "dog"})
	service := NewPetService(repository)

	if err := service.DeletePet(context.Background(), "owner-2", "pet-1"); !errors.Is(err, core_errors.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
	if err := service.DeletePet(context.Background(), "owner-1", "pet-1"); err != nil {
		t.Fatalf("DeletePet by owner: %v", err)
	}
	if _, ok := repository.pets["pet-1"]; ok {
		t.Error("pet was not deleted")
	}
}

func TestSpeciesByBreed(t *testing.T) {
	cases := map[string]string{
		"Golden Retriever":  "dog",
		"такса":             "dog",
		"Welsh Corgi":       "dog",
		"British Shorthair": "cat",
		"Мейн-кун":          "cat",
		"":                  "",
		"Axolotl":           "",
	}
	for breed, want := range cases {
		if got := SpeciesByBreed(breed); got != want {
			t.Errorf("SpeciesByBreed(%q) = %q, want %q", breed, got, want)
		}
	}
}
