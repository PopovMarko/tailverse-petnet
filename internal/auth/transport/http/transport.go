package auth_transport_http

import (
	"context"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type AuthService interface {
	Register(ctx context.Context, email, password, nickname string) (core_domain.Owner, core_auth.TokenPair, error)
	Login(ctx context.Context, email, password string) (core_auth.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (core_auth.TokenPair, error)
}

type AuthHttpHandler struct {
	authService AuthService
}

func NewAuthHttpHandler(authService AuthService) *AuthHttpHandler {
	return &AuthHttpHandler{
		authService: authService,
	}
}

func NewAuthRouter(h *AuthHttpHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)

	return r
}
