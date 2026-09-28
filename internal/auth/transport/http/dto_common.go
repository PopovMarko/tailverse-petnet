package auth_transport_http

import (
	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type RegisterRequestDto struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Nickname string `json:"nickname" validate:"required,max=50"`
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
