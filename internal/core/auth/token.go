package core_auth

import (
	"fmt"
	"time"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type claims struct {
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies HS256 JWTs. Subject of every token is the owner id.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(config AuthConfig) *TokenManager {
	return &TokenManager{
		secret:     []byte(config.JWTSecret),
		accessTTL:  config.AccessTTL,
		refreshTTL: config.RefreshTTL,
	}
}

func (m *TokenManager) IssuePair(ownerId string) (TokenPair, error) {
	now := time.Now()

	access, err := m.sign(ownerId, tokenTypeAccess, now, m.accessTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}
	refresh, err := m.sign(ownerId, tokenTypeRefresh, now, m.refreshTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

// ParseAccessToken returns the owner id stored in a valid access token.
func (m *TokenManager) ParseAccessToken(token string) (string, error) {
	return m.parse(token, tokenTypeAccess)
}

// ParseRefreshToken returns the owner id stored in a valid refresh token.
func (m *TokenManager) ParseRefreshToken(token string) (string, error) {
	return m.parse(token, tokenTypeRefresh)
}

func (m *TokenManager) sign(ownerId string, tokenType string, now time.Time, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   ownerId,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
	return token.SignedString(m.secret)
}

func (m *TokenManager) parse(token string, expectedType string) (string, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&claims{},
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("parse token: %w: %w", err, core_errors.ErrUnauthorized)
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || c.Subject == "" {
		return "", fmt.Errorf("token has no subject: %w", core_errors.ErrUnauthorized)
	}
	if c.TokenType != expectedType {
		return "", fmt.Errorf("expected %s token, got %q: %w", expectedType, c.TokenType, core_errors.ErrUnauthorized)
	}
	return c.Subject, nil
}
