package core_auth

import (
	"context"
	"errors"
	"fmt"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type ownerContextKey struct{}

var ownerKey = ownerContextKey{}

var errNoOwner = errors.New("no authenticated owner in context")

func ToContext(ctx context.Context, ownerId string) context.Context {
	return context.WithValue(ctx, ownerKey, ownerId)
}

// OwnerIdFromContext returns the id of the authenticated owner set by the Auth middleware.
func OwnerIdFromContext(ctx context.Context) (string, error) {
	ownerId, ok := ctx.Value(ownerKey).(string)
	if !ok || ownerId == "" {
		return "", fmt.Errorf("%w: %w", errNoOwner, core_errors.ErrUnauthorized)
	}
	return ownerId, nil
}
