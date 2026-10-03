package core_domain

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

const (
	MaxNicknameLength  = 50
	MaxAvatarUrlLength = 2048
)

// OwnerGenders are the accepted values of Owner.Gender; an empty string means "not set".
var OwnerGenders = []string{"male", "female", "other"}

type Owner struct {
	Id              string
	Email           string
	PasswordHash    string
	Nickname        string
	Gender          string // empty when not set
	AvatarUrl       string // empty when not set
	IsProfilePublic bool
	Visibility      OwnerVisibility
	CreatedAt       time.Time
}

// OwnerVisibility says which optional profile fields other users may see. The nickname is always public.
type OwnerVisibility struct {
	Gender    bool
	AvatarUrl bool
}

// DefaultOwnerVisibility matches the column defaults of migration 0002.
var DefaultOwnerVisibility = OwnerVisibility{Gender: false, AvatarUrl: true}

// OwnerPatch holds the editable profile fields; nil means "leave unchanged", an empty string clears gender/avatar_url.
type OwnerPatch struct {
	Nickname   *string
	Gender     *string
	AvatarUrl  *string
	Visibility OwnerVisibilityPatch
}

type OwnerVisibilityPatch struct {
	Gender    *bool
	AvatarUrl *bool
}

// PublicOwner is the profile as other users see it: hidden or unset fields are nil, the email is never included.
type PublicOwner struct {
	Id        string
	Nickname  string
	Gender    *string
	AvatarUrl *string
	CreatedAt time.Time
}

func (o Owner) Apply(patch OwnerPatch) Owner {
	if patch.Nickname != nil {
		o.Nickname = *patch.Nickname
	}
	if patch.Gender != nil {
		o.Gender = *patch.Gender
	}
	if patch.AvatarUrl != nil {
		o.AvatarUrl = *patch.AvatarUrl
	}
	if patch.Visibility.Gender != nil {
		o.Visibility.Gender = *patch.Visibility.Gender
	}
	if patch.Visibility.AvatarUrl != nil {
		o.Visibility.AvatarUrl = *patch.Visibility.AvatarUrl
	}
	return o
}

// NormalizeProfile trims the profile fields and lowercases the gender.
func (o Owner) NormalizeProfile() Owner {
	o.Nickname = strings.TrimSpace(o.Nickname)
	o.Gender = strings.ToLower(strings.TrimSpace(o.Gender))
	o.AvatarUrl = strings.TrimSpace(o.AvatarUrl)
	return o
}

// ValidateProfile checks the fields an owner can edit (call NormalizeProfile first).
func (o Owner) ValidateProfile() error {
	if o.Nickname == "" {
		return fmt.Errorf("nickname is required: %w", core_errors.ErrInvalidArgument)
	}
	if len([]rune(o.Nickname)) > MaxNicknameLength {
		return fmt.Errorf("nickname is longer than %d characters: %w", MaxNicknameLength, core_errors.ErrInvalidArgument)
	}
	if o.Gender != "" && !slices.Contains(OwnerGenders, o.Gender) {
		return fmt.Errorf("gender must be one of %v or empty: %w", OwnerGenders, core_errors.ErrInvalidArgument)
	}
	if o.AvatarUrl != "" {
		if len(o.AvatarUrl) > MaxAvatarUrlLength {
			return fmt.Errorf("avatar_url is longer than %d characters: %w", MaxAvatarUrlLength, core_errors.ErrInvalidArgument)
		}
		parsed, err := url.Parse(o.AvatarUrl)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("avatar_url %q is not an http(s) url: %w", o.AvatarUrl, core_errors.ErrInvalidArgument)
		}
	}
	return nil
}

// Public returns the profile as other users see it. A profile that is not public shows only the nickname.
func (o Owner) Public() PublicOwner {
	public := PublicOwner{
		Id:        o.Id,
		Nickname:  o.Nickname,
		CreatedAt: o.CreatedAt,
	}
	if o.IsProfilePublic && o.Visibility.Gender && o.Gender != "" {
		gender := o.Gender
		public.Gender = &gender
	}
	if o.IsProfilePublic && o.Visibility.AvatarUrl && o.AvatarUrl != "" {
		avatarUrl := o.AvatarUrl
		public.AvatarUrl = &avatarUrl
	}
	return public
}
