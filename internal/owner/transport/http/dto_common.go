package owner_transport_http

import (
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type VisibilityDto struct {
	Gender    bool `json:"gender"`
	AvatarUrl bool `json:"avatar_url"`
}

// OwnerResponseDto is the owner's own full profile (GET/PATCH /owners/me). Unset gender/avatar_url are null.
type OwnerResponseDto struct {
	Id         string        `json:"id"`
	Email      string        `json:"email"`
	Nickname   string        `json:"nickname"`
	Gender     *string       `json:"gender"`
	AvatarUrl  *string       `json:"avatar_url"`
	Visibility VisibilityDto `json:"visibility"`
	CreatedAt  time.Time     `json:"created_at"`
}

// PublicOwnerResponseDto is GET /owners/{id}: hidden or unset fields are null, the email is never sent.
type PublicOwnerResponseDto struct {
	Id        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Gender    *string   `json:"gender"`
	AvatarUrl *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// OwnerUpdateDto is the PATCH body: every field is optional, absent fields are left unchanged.
// An empty string clears gender or avatar_url.
type OwnerUpdateDto struct {
	Nickname   *string              `json:"nickname" validate:"omitempty,min=1,max=50"`
	Gender     *string              `json:"gender" validate:"omitempty,max=20"`
	AvatarUrl  *string              `json:"avatar_url" validate:"omitempty,max=2048"`
	Visibility *VisibilityUpdateDto `json:"visibility"`
}

type VisibilityUpdateDto struct {
	Gender    *bool `json:"gender"`
	AvatarUrl *bool `json:"avatar_url"`
}

func UpdateDtoToDomain(dto OwnerUpdateDto) core_domain.OwnerPatch {
	patch := core_domain.OwnerPatch{
		Nickname:  dto.Nickname,
		Gender:    dto.Gender,
		AvatarUrl: dto.AvatarUrl,
	}
	if dto.Visibility != nil {
		patch.Visibility = core_domain.OwnerVisibilityPatch{
			Gender:    dto.Visibility.Gender,
			AvatarUrl: dto.Visibility.AvatarUrl,
		}
	}
	return patch
}

func OwnerToDto(owner core_domain.Owner) OwnerResponseDto {
	return OwnerResponseDto{
		Id:        owner.Id,
		Email:     owner.Email,
		Nickname:  owner.Nickname,
		Gender:    nullableString(owner.Gender),
		AvatarUrl: nullableString(owner.AvatarUrl),
		Visibility: VisibilityDto{
			Gender:    owner.Visibility.Gender,
			AvatarUrl: owner.Visibility.AvatarUrl,
		},
		CreatedAt: owner.CreatedAt,
	}
}

func PublicOwnerToDto(owner core_domain.PublicOwner) PublicOwnerResponseDto {
	return PublicOwnerResponseDto{
		Id:        owner.Id,
		Nickname:  owner.Nickname,
		Gender:    owner.Gender,
		AvatarUrl: owner.AvatarUrl,
		CreatedAt: owner.CreatedAt,
	}
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
