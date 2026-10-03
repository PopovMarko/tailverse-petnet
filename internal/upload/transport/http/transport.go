package upload_transport_http

import (
	"context"
	"io"
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type UploadService interface {
	Upload(ctx context.Context, content io.Reader) (core_domain.Upload, error)
	Open(ctx context.Context, name string) (core_domain.StoredFile, error)
}

type UploadHttpHandler struct {
	uploadService UploadService
	// publicBaseUrl is the origin of the returned file URLs; empty means "take it from the request".
	publicBaseUrl string
}

func NewUploadHttpHandler(uploadService UploadService, publicBaseUrl string) *UploadHttpHandler {
	return &UploadHttpHandler{
		uploadService: uploadService,
		publicBaseUrl: publicBaseUrl,
	}
}

// NewUploadsRouter registers POST /api/v1/uploads.
func NewUploadsRouter(h *UploadHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.With(authMiddleware).Post("/", h.CreateUpload)

	return r
}

// NewUploadedFilesRouter serves the stored files at GET /uploads/{name}, without authentication.
func NewUploadedFilesRouter(h *UploadHttpHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/{name}", h.GetUploadedFile)

	return r
}
