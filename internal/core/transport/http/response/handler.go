package core_http_response

import (
	"fmt"
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
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

func (h *HttpResponseHandler) ErrorResponse(msg string, err error) {}

func (h *HttpResponseHandler) PanicResponse(message string, p any) {
	status := http.StatusInternalServerError
	err := fmt.Errorf("During handle request id: %s, get unexpected panic %v", h.w.Header().Get("X-Request-ID"), p)

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

}
