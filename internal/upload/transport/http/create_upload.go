package upload_transport_http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"go.uber.org/zap"
)

const (
	// maxRequestSize leaves room for the multipart headers around a file of the maximum size (10 MB).
	maxRequestSize = 11 << 20
	// uploadTimeout replaces the server-wide read/write timeouts: a 10 MB photo over a mobile network takes a while.
	uploadTimeout = 2 * time.Minute
)

// CreateUpload handles POST /uploads: multipart/form-data with the image in the "file" field.
func (h *UploadHttpHandler) CreateUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CreateUpload handler called")

	controller := http.NewResponseController(w)
	if err := controller.SetReadDeadline(time.Now().Add(uploadTimeout)); err != nil {
		logger.Debug("CreateUpload handler: can not extend read deadline", zap.Error(err))
	}
	if err := controller.SetWriteDeadline(time.Now().Add(uploadTimeout)); err != nil {
		logger.Debug("CreateUpload handler: can not extend write deadline", zap.Error(err))
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	reader, err := r.MultipartReader()
	if err != nil {
		httpResponseHandler.ErrorResponse("CreateUpload handler: expected multipart/form-data", fmt.Errorf("%w: %w", err, core_errors.ErrInvalidArgument))
		return
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			httpResponseHandler.ErrorResponse("CreateUpload handler: no file", fmt.Errorf("multipart field %q is required: %w", uploadFormField, core_errors.ErrInvalidArgument))
			return
		}
		if err != nil {
			httpResponseHandler.ErrorResponse("CreateUpload handler: read multipart body", bodyError(err))
			return
		}
		if part.FormName() != uploadFormField {
			part.Close()
			continue
		}

		upload, err := h.uploadService.Upload(ctx, part)
		part.Close()
		if err != nil {
			httpResponseHandler.ErrorResponse("CreateUpload handler: upload service", bodyError(err))
			return
		}

		logger.Debug("file uploaded", zap.String("name", upload.Name), zap.String("content_type", upload.ContentType), zap.Int64("size", upload.Size))
		httpResponseHandler.JsonResponse(UploadResponseDto{Url: fileUrl(r, h.publicBaseUrl, upload.Name)}, http.StatusCreated)
		return
	}
}

// bodyError classifies errors that come from reading the request body.
func bodyError(err error) error {
	var maxBytesErr *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytesErr):
		return fmt.Errorf("request body is larger than %d bytes: %w", maxBytesErr.Limit, core_errors.ErrTooLarge)
	case errors.Is(err, core_errors.ErrTooLarge), errors.Is(err, core_errors.ErrUnsupportedMediaType),
		errors.Is(err, core_errors.ErrInvalidArgument):
		return err
	case errors.Is(err, io.ErrUnexpectedEOF):
		return fmt.Errorf("malformed multipart body: %w", core_errors.ErrInvalidArgument)
	default:
		return err
	}
}
