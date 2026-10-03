package core_domain

import (
	"errors"
	"strings"
	"testing"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

func TestOwnerPublicHonoursVisibility(t *testing.T) {
	owner := Owner{
		Id:              "owner-1",
		Email:           "me@example.com",
		Nickname:        "marko",
		Gender:          "male",
		AvatarUrl:       "http://127.0.0.1:8080/uploads/a.jpg",
		IsProfilePublic: true,
	}

	cases := []struct {
		name          string
		visibility    OwnerVisibility
		profilePublic bool
		wantGender    bool
		wantAvatar    bool
	}{
		{name: "all visible", visibility: OwnerVisibility{Gender: true, AvatarUrl: true}, profilePublic: true, wantGender: true, wantAvatar: true},
		{name: "gender hidden", visibility: OwnerVisibility{Gender: false, AvatarUrl: true}, profilePublic: true, wantAvatar: true},
		{name: "avatar hidden", visibility: OwnerVisibility{Gender: true, AvatarUrl: false}, profilePublic: true, wantGender: true},
		{name: "all hidden", visibility: OwnerVisibility{}, profilePublic: true},
		{name: "private profile hides everything but the nickname", visibility: OwnerVisibility{Gender: true, AvatarUrl: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := owner
			o.Visibility = tc.visibility
			o.IsProfilePublic = tc.profilePublic

			public := o.Public()
			if public.Nickname != "marko" || public.Id != "owner-1" {
				t.Errorf("id/nickname = %q/%q, want always public", public.Id, public.Nickname)
			}
			if (public.Gender != nil) != tc.wantGender {
				t.Errorf("gender = %v, want visible=%v", public.Gender, tc.wantGender)
			}
			if (public.AvatarUrl != nil) != tc.wantAvatar {
				t.Errorf("avatar_url = %v, want visible=%v", public.AvatarUrl, tc.wantAvatar)
			}
		})
	}
}

func TestOwnerPublicOmitsUnsetFields(t *testing.T) {
	public := Owner{Nickname: "marko", IsProfilePublic: true, Visibility: OwnerVisibility{Gender: true, AvatarUrl: true}}.Public()
	if public.Gender != nil || public.AvatarUrl != nil {
		t.Errorf("unset fields should be nil, got gender=%v avatar=%v", public.Gender, public.AvatarUrl)
	}
}

func TestOwnerApplyPatch(t *testing.T) {
	owner := Owner{Nickname: "marko", Gender: "male", AvatarUrl: "https://x/a.jpg", Visibility: OwnerVisibility{Gender: true, AvatarUrl: true}}

	updated := owner.Apply(OwnerPatch{
		Gender:     ptr(""),
		Visibility: OwnerVisibilityPatch{AvatarUrl: ptr(false)},
	})
	if updated.Nickname != "marko" || updated.AvatarUrl != "https://x/a.jpg" {
		t.Errorf("absent fields changed: %+v", updated)
	}
	if updated.Gender != "" {
		t.Errorf("gender = %q, want cleared", updated.Gender)
	}
	if !updated.Visibility.Gender || updated.Visibility.AvatarUrl {
		t.Errorf("visibility = %+v, want gender kept true and avatar_url set to false", updated.Visibility)
	}
}

func TestOwnerValidateProfile(t *testing.T) {
	valid := Owner{Nickname: "marko", Gender: "female", AvatarUrl: "https://cdn.example.com/a.png"}
	if err := valid.ValidateProfile(); err != nil {
		t.Fatalf("valid profile: %v", err)
	}

	invalid := map[string]Owner{
		"empty nickname":    {Nickname: ""},
		"unknown gender":    {Nickname: "m", Gender: "robot"},
		"relative avatar":   {Nickname: "m", AvatarUrl: "/uploads/a.jpg"},
		"non-http avatar":   {Nickname: "m", AvatarUrl: "ftp://x/a.jpg"},
		"too long nickname": {Nickname: strings.Repeat("ы", MaxNicknameLength+1)},
	}
	for name, owner := range invalid {
		if err := owner.ValidateProfile(); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("%s: err = %v, want ErrInvalidArgument", name, err)
		}
	}
}

func TestOwnerNormalizeProfile(t *testing.T) {
	owner := Owner{Nickname: "  marko ", Gender: " Female ", AvatarUrl: " https://x/a.jpg "}.NormalizeProfile()
	if owner.Nickname != "marko" || owner.Gender != "female" || owner.AvatarUrl != "https://x/a.jpg" {
		t.Errorf("normalized = %+v", owner)
	}
}
