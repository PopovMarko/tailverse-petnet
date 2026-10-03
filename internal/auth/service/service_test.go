package auth_service

import (
	"context"
	"errors"
	"testing"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

type fakeOwnerRepository struct {
	created []core_domain.Owner
}

func (f *fakeOwnerRepository) CreateOwner(_ context.Context, owner core_domain.Owner) (core_domain.Owner, error) {
	owner.Id = "owner-1"
	f.created = append(f.created, owner)
	return owner, nil
}

func (f *fakeOwnerRepository) GetOwnerByEmail(context.Context, string) (core_domain.Owner, error) {
	return core_domain.Owner{}, core_errors.ErrNotFound
}

func (f *fakeOwnerRepository) GetOwner(context.Context, string) (core_domain.Owner, error) {
	return core_domain.Owner{}, core_errors.ErrNotFound
}

type fakeTokens struct{}

func (fakeTokens) IssuePair(ownerId string) (core_auth.TokenPair, error) {
	return core_auth.TokenPair{AccessToken: "access-" + ownerId, RefreshToken: "refresh-" + ownerId}, nil
}

func (fakeTokens) ParseRefreshToken(string) (string, error) { return "", core_errors.ErrUnauthorized }

func ptr[T any](v T) *T { return &v }

func TestRegisterWithoutProfileFieldsUsesDefaults(t *testing.T) {
	repository := &fakeOwnerRepository{}
	service := NewAuthService(repository, fakeTokens{})

	owner, tokens, err := service.Register(context.Background(), " Me@Example.com ", "password123", core_domain.OwnerPatch{Nickname: ptr(" marko ")})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if owner.Email != "me@example.com" || owner.Nickname != "marko" {
		t.Errorf("email/nickname = %q/%q", owner.Email, owner.Nickname)
	}
	if owner.Gender != "" || owner.AvatarUrl != "" || owner.Visibility != core_domain.DefaultOwnerVisibility {
		t.Errorf("profile = %+v, want empty with default visibility", owner)
	}
	if bcrypt.CompareHashAndPassword([]byte(repository.created[0].PasswordHash), []byte("password123")) != nil {
		t.Errorf("password is not stored as a bcrypt hash")
	}
	if tokens.AccessToken == "" {
		t.Errorf("no access token")
	}
}

func TestRegisterWithProfileFields(t *testing.T) {
	service := NewAuthService(&fakeOwnerRepository{}, fakeTokens{})

	owner, _, err := service.Register(context.Background(), "me@example.com", "password123", core_domain.OwnerPatch{
		Nickname:   ptr("marko"),
		Gender:     ptr("Male"),
		AvatarUrl:  ptr("http://127.0.0.1:8080/uploads/0123456789abcdef0123456789abcdef.jpg"),
		Visibility: core_domain.OwnerVisibilityPatch{Gender: ptr(true)},
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if owner.Gender != "male" || owner.AvatarUrl == "" {
		t.Errorf("profile = %+v", owner)
	}
	if !owner.Visibility.Gender || !owner.Visibility.AvatarUrl {
		t.Errorf("visibility = %+v, want gender set to true and avatar_url default true", owner.Visibility)
	}
}

func TestRegisterRejectsInvalidProfile(t *testing.T) {
	repository := &fakeOwnerRepository{}
	service := NewAuthService(repository, fakeTokens{})

	for name, profile := range map[string]core_domain.OwnerPatch{
		"no nickname":    {},
		"bad gender":     {Nickname: ptr("m"), Gender: ptr("unknown")},
		"bad avatar url": {Nickname: ptr("m"), AvatarUrl: ptr("not a url")},
	} {
		if _, _, err := service.Register(context.Background(), "me@example.com", "password123", profile); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("%s: err = %v, want ErrInvalidArgument", name, err)
		}
	}
	if len(repository.created) != 0 {
		t.Errorf("invalid owners were created")
	}
}
