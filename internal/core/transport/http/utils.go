package core_http_utils

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"github.com/google/uuid"
)

func GetStringPathParams(r *http.Request, key string) (string, error) {
	param := r.PathValue(key)
	if param == "" {
		return "", fmt.Errorf("empty parameter in request: %w", core_errors.ErrInvalidArgument)
	}
	return param, nil
}

// GetUUIDPathParam returns a path parameter that must be a valid UUID.
func GetUUIDPathParam(r *http.Request, key string) (string, error) {
	param, err := GetStringPathParams(r, key)
	if err != nil {
		return "", err
	}
	if err := uuid.Validate(param); err != nil {
		return "", fmt.Errorf("path parameter %q is not a valid uuid: %w", key, core_errors.ErrInvalidArgument)
	}
	return param, nil
}

// GetFloatQueryParam returns the query parameter as float64, or nil when it is absent.
func GetFloatQueryParam(r *http.Request, key string) (*float64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("query parameter %q is not a number: %w", key, core_errors.ErrInvalidArgument)
	}
	return &value, nil
}

// GetIntQueryParam returns the query parameter as int, or nil when it is absent.
func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("query parameter %q is not an integer: %w", key, core_errors.ErrInvalidArgument)
	}
	return &value, nil
}

// GetTimeQueryParam returns the RFC 3339 query parameter, or nil when it is absent.
func GetTimeQueryParam(r *http.Request, key string) (*time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("query parameter %q is not an RFC 3339 timestamp: %w", key, core_errors.ErrInvalidArgument)
	}
	return &value, nil
}

// GetUUIDQueryParam returns the query parameter that must be a UUID, or nil when it is absent.
func GetUUIDQueryParam(r *http.Request, key string) (*string, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	if err := uuid.Validate(raw); err != nil {
		return nil, fmt.Errorf("query parameter %q is not a valid uuid: %w", key, core_errors.ErrInvalidArgument)
	}
	return &raw, nil
}
