package core_http_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"go.uber.org/zap"
)

type HttpResponseHandler struct {
	logger *core_logger.Logger
	w      http.ResponseWriter
}

func NewHttpResponseHandler(logger *core_logger.Logger, w http.ResponseWriter) *HttpResponseHandler {
	return &HttpResponseHandler{
		logger: logger,
		w:      w,
	}
}

// ErrorResponse maps domain errors from core_errors to HTTP status codes.
// Internal errors are logged, but their details are not sent to the client.
func (h *HttpResponseHandler) ErrorResponse(msg string, err error) {
	status := statusFromError(err)
	if status == http.StatusInternalServerError {
		h.logger.Error(msg, zap.Error(err))
		h.errorResponse(msg, status, errors.New("internal server error"))
		return
	}

	h.logger.Debug(msg, zap.Int("status", status), zap.Error(err))
	h.errorResponse(msg, status, err)
}

func (h *HttpResponseHandler) PanicResponse(message string, p any) {
	status := http.StatusInternalServerError
	err := fmt.Errorf("During handle request id: %s, get unexpected panic %v", h.w.Header().Get("X-Request-ID"), p)

	h.logger.Error(message, zap.Error(err))
	h.errorResponse(message, status, err)
}

func (h *HttpResponseHandler) errorResponse(message string, status int, err error) {
	response := map[string]string{
		"msg":   message,
		"error": err.Error(),
	}
	h.JsonResponse(response, status)
}

func (h *HttpResponseHandler) JsonResponse(body any, status int) {
	h.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	h.w.WriteHeader(status)

	if err := json.NewEncoder(h.w).Encode(body); err != nil {
		h.logger.Error("encode json response", zap.Error(err))
	}
}

func (h *HttpResponseHandler) NoContentResponse() {
	h.w.WriteHeader(http.StatusNoContent)
}

func statusFromError(err error) int {
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		return http.StatusBadRequest
	case errors.Is(err, core_errors.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, core_errors.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, core_errors.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, core_errors.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
