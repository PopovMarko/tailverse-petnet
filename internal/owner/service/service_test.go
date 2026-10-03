package owner_service

import (
	"context"
	"errors"
	"testing"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type fakeOwnerRepository struct {
	owners map[string]core_domain.Owner
	saves  int
}

func (f *fakeOwnerRepository) GetOwner(_ context.Context, id string) (core_domain.Owner, error) {
	owner, ok := f.owners[id]
	if !ok {
		return core_domain.Owner{}, core_errors.ErrNotFound
	}
	return owner, nil
}

func (f *fakeOwnerRepository) UpdateProfile(_ context.Context, owner core_domain.Owner) (core_domain.Owner, error) {
	f.saves++
	f.owners[owner.Id] = owner
	return owner, nil
}

func ptr[T any](v T) *T { return &v }

func newService() (*OwnerService, *fakeOwnerRepository) {
	repository := &fakeOwnerRepository{owners: map[string]core_domain.Owner{
		"owner-1": {
			Id:              "owner-1",
			Email:           "me@example.com",
			Nickname:        "marko",
			Gender:          "male",
			AvatarUrl:       "http://127.0.0.1:8080/uploads/0123456789abcdef0123456789abcdef.jpg",
			IsProfilePublic: true,
			Visibility:      core_domain.DefaultOwnerVisibility,
		},
	}}
	return NewOwnerService(repository), repository
}

func TestGetPublicOwnerFiltersHiddenFields(t *testing.T) {
	service, _ := newService()

	public, err := service.GetPublicOwner(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("GetPublicOwner: %v", err)
	}
	// Default visibility: gender hidden, avatar shown.
	if public.Gender != nil {
		t.Errorf("gender = %q, want hidden by default", *public.Gender)
	}
	if public.AvatarUrl == nil {
		t.Errorf("avatar_url hidden, want visible by default")
	}
	if public.Nickname != "marko" {
		t.Errorf("nickname = %q, want marko", public.Nickname)
	}
}

func TestUpdateOwnerChangesVisibilityOfPublicView(t *testing.T) {
	service, _ := newService()
	ctx := context.Background()

	if _, err := service.UpdateOwner(ctx, "owner-1", core_domain.OwnerPatch{
		Visibility: core_domain.OwnerVisibilityPatch{Gender: ptr(true), AvatarUrl: ptr(false)},
	}); err != nil {
		t.Fatalf("UpdateOwner: %v", err)
	}

	public, err := service.GetPublicOwner(ctx, "owner-1")
	if err != nil {
		t.Fatalf("GetPublicOwner: %v", err)
	}
	if public.Gender == nil || *public.Gender != "male" {
		t.Errorf("gender = %v, want male", public.Gender)
	}
	if public.AvatarUrl != nil {
		t.Errorf("avatar_url = %q, want hidden", *public.AvatarUrl)
	}
}

func TestUpdateOwnerIsPartial(t *testing.T) {
	service, _ := newService()

	owner, err := service.UpdateOwner(context.Background(), "owner-1", core_domain.OwnerPatch{Nickname: ptr("  Marko P ")})
	if err != nil {
		t.Fatalf("UpdateOwner: %v", err)
	}
	if owner.Nickname != "Marko P" {
		t.Errorf("nickname = %q, want trimmed", owner.Nickname)
	}
	if owner.Gender != "male" || owner.AvatarUrl == "" || owner.Visibility != core_domain.DefaultOwnerVisibility {
		t.Errorf("absent fields changed: %+v", owner)
	}
}

func TestUpdateOwnerRejectsInvalidProfile(t *testing.T) {
	service, repository := newService()

	cases := map[string]core_domain.OwnerPatch{
		"blank nickname":    {Nickname: ptr("   ")},
		"unknown gender":    {Gender: ptr("robot")},
		"not an http url":   {AvatarUrl: ptr("file:///etc/passwd")},
		"relative url path": {AvatarUrl: ptr("/uploads/a.jpg")},
	}
	for name, patch := range cases {
		if _, err := service.UpdateOwner(context.Background(), "owner-1", patch); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("%s: err = %v, want ErrInvalidArgument", name, err)
		}
	}
	if repository.saves != 0 {
		t.Errorf("invalid patches were saved %d times", repository.saves)
	}
}

func TestUpdateOwnerUnknownOwner(t *testing.T) {
	service, _ := newService()
	if _, err := service.UpdateOwner(context.Background(), "missing", core_domain.OwnerPatch{}); !errors.Is(err, core_errors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
