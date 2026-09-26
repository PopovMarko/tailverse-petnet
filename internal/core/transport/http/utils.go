package core_http_utils

import (
	"fmt"
	"net/http"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

func GetStringPathParams(r *http.Request, key string) (string, error) {
	param := r.PathValue(key)
	if param == "" {
		return "", fmt.Errorf("empty parameter in request: %w", core_errors.ErrInvalidArgument)
	}
	return param, nil
}
