package auth_transport_http

import (
	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

// RegisterRequestDto: gender, avatar_url and visibility are optional profile fields (see GET /owners/me).
type RegisterRequestDto struct {
	Email      string                        `json:"email" validate:"required,email,max=254"`
	Password   string                        `json:"password" validate:"required,min=8,max=72"`
	Nickname   string                        `json:"nickname" validate:"required,max=50"`
	Gender     *string                       `json:"gender" validate:"omitempty,max=20"`
	AvatarUrl  *string                       `json:"avatar_url" validate:"omitempty,max=2048"`
	Visibility *RegisterVisibilityRequestDto `json:"visibility"`
}

type RegisterVisibilityRequestDto struct {
	Gender    *bool `json:"gender"`
	AvatarUrl *bool `json:"avatar_url"`
}

type RegisterResponseDto struct {
	OwnerId      string `json:"owner_id"`
	Email        string `json:"email"`
	Nickname     string `json:"nickname"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequestDto struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequestDto struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokensResponseDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func RegisterDtoToProfile(dto RegisterRequestDto) core_domain.OwnerPatch {
	nickname := dto.Nickname
	profile := core_domain.OwnerPatch{
		Nickname:  &nickname,
		Gender:    dto.Gender,
		AvatarUrl: dto.AvatarUrl,
	}
	if dto.Visibility != nil {
		profile.Visibility = core_domain.OwnerVisibilityPatch{
			Gender:    dto.Visibility.Gender,
			AvatarUrl: dto.Visibility.AvatarUrl,
		}
	}
	return profile
}

func RegisterDomainToDto(owner core_domain.Owner, tokens core_auth.TokenPair) RegisterResponseDto {
	return RegisterResponseDto{
		OwnerId:      owner.Id,
		Email:        owner.Email,
		Nickname:     owner.Nickname,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
}

func TokensToDto(tokens core_auth.TokenPair) TokensResponseDto {
	return TokensResponseDto{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
}
