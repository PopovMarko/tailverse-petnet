package auth_service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength = 8

// dummyHash is compared against when the email is unknown, so that login takes the same time either way.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("tailverse-dummy-password"), bcrypt.DefaultCost)

type OwnerRepository interface {
	CreateOwner(ctx context.Context, owner core_domain.Owner) (core_domain.Owner, error)
	GetOwnerByEmail(ctx context.Context, email string) (core_domain.Owner, error)
	GetOwner(ctx context.Context, id string) (core_domain.Owner, error)
}

type TokenManager interface {
	IssuePair(ownerId string) (core_auth.TokenPair, error)
	ParseRefreshToken(token string) (string, error)
}

type AuthService struct {
	repository OwnerRepository
	tokens     TokenManager
}

func NewAuthService(repository OwnerRepository, tokens TokenManager) *AuthService {
	return &AuthService{
		repository: repository,
		tokens:     tokens,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, nickname string) (core_domain.Owner, core_auth.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	nickname = strings.TrimSpace(nickname)

	if _, err := mail.ParseAddress(email); err != nil {
		return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("invalid email: %w", core_errors.ErrInvalidArgument)
	}
	if len(password) < minPasswordLength {
		return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("password must be at least %d characters: %w", minPasswordLength, core_errors.ErrInvalidArgument)
	}
	if nickname == "" {
		return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("nickname is required: %w", core_errors.ErrInvalidArgument)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	owner, err := s.repository.CreateOwner(ctx, core_domain.Owner{
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     nickname,
	})
	if err != nil {
		if errors.Is(err, core_errors.ErrConflict) {
			return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("email is already registered: %w", core_errors.ErrConflict)
		}
		return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("create owner: %w", err)
	}

	tokens, err := s.tokens.IssuePair(owner.Id)
	if err != nil {
		return core_domain.Owner{}, core_auth.TokenPair{}, fmt.Errorf("issue tokens: %w", err)
	}
	return owner, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (core_auth.TokenPair, error) {
	owner, err := s.repository.GetOwnerByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return core_auth.TokenPair{}, fmt.Errorf("invalid email or password: %w", core_errors.ErrUnauthorized)
		}
		return core_auth.TokenPair{}, fmt.Errorf("get owner: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(owner.PasswordHash), []byte(password)); err != nil {
		return core_auth.TokenPair{}, fmt.Errorf("invalid email or password: %w", core_errors.ErrUnauthorized)
	}

	tokens, err := s.tokens.IssuePair(owner.Id)
	if err != nil {
		return core_auth.TokenPair{}, fmt.Errorf("issue tokens: %w", err)
	}
	return tokens, nil
}

// Refresh exchanges a valid refresh token for a new token pair. The owner must still exist.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (core_auth.TokenPair, error) {
	ownerId, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return core_auth.TokenPair{}, err
	}

	if _, err := s.repository.GetOwner(ctx, ownerId); err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return core_auth.TokenPair{}, fmt.Errorf("owner no longer exists: %w", core_errors.ErrUnauthorized)
		}
		return core_auth.TokenPair{}, fmt.Errorf("get owner: %w", err)
	}

	tokens, err := s.tokens.IssuePair(ownerId)
	if err != nil {
		return core_auth.TokenPair{}, fmt.Errorf("issue tokens: %w", err)
	}
	return tokens, nil
}
