package core_errors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	// ErrTooLarge and ErrUnsupportedMediaType are returned by file uploads (413 and 415).
	ErrTooLarge             = errors.New("payload too large")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
)
