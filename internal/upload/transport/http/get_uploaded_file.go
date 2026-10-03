package upload_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// GetUploadedFile handles GET /uploads/{name}. Files never change once stored, so they are cached for a year.
func (h *UploadHttpHandler) GetUploadedFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	file, err := h.uploadService.Open(ctx, r.PathValue("name"))
	if err != nil {
		httpResponseHandler.ErrorResponse("GetUploadedFile handler: upload service", err)
		return
	}
	defer file.Content.Close()

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, file.Name, file.ModTime, file.Content)
}
