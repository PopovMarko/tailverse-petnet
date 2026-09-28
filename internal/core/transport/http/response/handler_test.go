package core_http_response

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

func TestStatusFromError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{fmt.Errorf("decode: %w", core_errors.ErrInvalidArgument), http.StatusBadRequest},
		{fmt.Errorf("token: %w", core_errors.ErrUnauthorized), http.StatusUnauthorized},
		{fmt.Errorf("owner: %w", core_errors.ErrForbidden), http.StatusForbidden},
		{fmt.Errorf("pet: %w", core_errors.ErrNotFound), http.StatusNotFound},
		{fmt.Errorf("join: %w", core_errors.ErrConflict), http.StatusConflict},
		{errors.New("connection refused"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		if got := statusFromError(tc.err); got != tc.want {
			t.Errorf("statusFromError(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}
