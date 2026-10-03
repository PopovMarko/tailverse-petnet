package core_auth

import (
	"errors"
	"testing"
	"time"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

func newTestManager(accessTTL time.Duration) *TokenManager {
	return NewTokenManager(AuthConfig{
		JWTSecret:  "test-secret-that-is-at-least-32-characters",
		AccessTTL:  accessTTL,
		RefreshTTL: time.Hour,
	})
}

func TestTokenPairRoundTrip(t *testing.T) {
	manager := newTestManager(time.Minute)

	pair, err := manager.IssuePair("owner-1")
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}

	if ownerId, err := manager.ParseAccessToken(pair.AccessToken); err != nil || ownerId != "owner-1" {
		t.Errorf("ParseAccessToken = %q, %v", ownerId, err)
	}
	if ownerId, err := manager.ParseRefreshToken(pair.RefreshToken); err != nil || ownerId != "owner-1" {
		t.Errorf("ParseRefreshToken = %q, %v", ownerId, err)
	}
}

func TestTokenTypesAreNotInterchangeable(t *testing.T) {
	manager := newTestManager(time.Minute)
	pair, _ := manager.IssuePair("owner-1")

	if _, err := manager.ParseAccessToken(pair.RefreshToken); !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Errorf("refresh token accepted as access token: %v", err)
	}
	if _, err := manager.ParseRefreshToken(pair.AccessToken); !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Errorf("access token accepted as refresh token: %v", err)
	}
}

func TestRejectsExpiredAndForeignTokens(t *testing.T) {
	expired, _ := newTestManager(-time.Minute).IssuePair("owner-1")
	if _, err := newTestManager(time.Minute).ParseAccessToken(expired.AccessToken); !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Errorf("expired token: err = %v, want ErrUnauthorized", err)
	}

	other := NewTokenManager(AuthConfig{JWTSecret: "another-secret-that-is-at-least-32-chars", AccessTTL: time.Minute, RefreshTTL: time.Hour})
	foreign, _ := other.IssuePair("owner-1")
	if _, err := newTestManager(time.Minute).ParseAccessToken(foreign.AccessToken); !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Errorf("token signed with another secret: err = %v, want ErrUnauthorized", err)
	}

	if _, err := newTestManager(time.Minute).ParseAccessToken("garbage"); !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Errorf("garbage token: err = %v, want ErrUnauthorized", err)
	}
}
