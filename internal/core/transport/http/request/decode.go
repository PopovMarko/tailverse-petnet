package core_http_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"github.com/go-playground/validator/v10"
)

var dtoValidator = validator.New()

type Validator interface {
	Validate() error
}

// DecodeAndValidate check for destDto implement Validator interface
func DecodeAndValidate(r *http.Request, destDto any) error {
	if err := json.NewDecoder(r.Body).Decode(destDto); err != nil {
		return fmt.Errorf("json decode: %w, %w", err, core_errors.ErrInvalidArgument)
	}
	var err error
	if d, ok := destDto.(Validator); ok {
		err = d.Validate()
	} else {
		err = dtoValidator.Struct(destDto)
	}
	if err != nil {
		return fmt.Errorf("dto validator: %w: %w", err, core_errors.ErrInvalidArgument)
	}
	return nil
}
