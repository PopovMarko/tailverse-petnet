package upload_transport_http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"go.uber.org/zap"
)

// withLogger puts a no-op logger into the request context, as the Logger middleware does in the server.
func withLogger(r *http.Request) *http.Request {
	return r.WithContext(core_logger.ToContext(r.Context(), &core_logger.Logger{Logger: zap.NewNop()}))
}

type fakeUploadService struct {
	received []byte
}

func (f *fakeUploadService) Upload(_ context.Context, content io.Reader) (core_domain.Upload, error) {
	var err error
	if f.received, err = io.ReadAll(content); err != nil {
		return core_domain.Upload{}, err
	}
	return core_domain.Upload{Name: "0123456789abcdef0123456789abcdef.jpg", ContentType: "image/jpeg"}, nil
}

func (f *fakeUploadService) Open(context.Context, string) (core_domain.StoredFile, error) {
	return core_domain.StoredFile{}, nil
}

func multipartRequest(t *testing.T, field string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("comment", "ignored"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile(field, "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "http://192.168.1.20:8080/api/v1/uploads", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return withLogger(r)
}

func TestCreateUploadReturnsAbsoluteUrl(t *testing.T) {
	service := &fakeUploadService{}
	h := NewUploadHttpHandler(service, "")

	w := httptest.NewRecorder()
	h.CreateUpload(w, multipartRequest(t, "file", []byte("image-bytes")))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	var response UploadResponseDto
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if want := "http://192.168.1.20:8080/uploads/0123456789abcdef0123456789abcdef.jpg"; response.Url != want {
		t.Errorf("url = %q, want %q", response.Url, want)
	}
	if string(service.received) != "image-bytes" {
		t.Errorf("service received %q", service.received)
	}
}

func TestCreateUploadRequiresFileField(t *testing.T) {
	w := httptest.NewRecorder()
	NewUploadHttpHandler(&fakeUploadService{}, "").CreateUpload(w, multipartRequest(t, "photo", []byte("x")))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreateUploadRequiresMultipart(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", bytes.NewReader([]byte(`{"file":"x"}`)))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewUploadHttpHandler(&fakeUploadService{}, "").CreateUpload(w, withLogger(r))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreateUploadRejectsOversizedBody(t *testing.T) {
	w := httptest.NewRecorder()
	NewUploadHttpHandler(&fakeUploadService{}, "").CreateUpload(w, multipartRequest(t, "file", make([]byte, maxRequestSize+1)))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
}

func TestFileUrlPrefersConfiguredBase(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/uploads", nil)
	if got := fileUrl(r, "https://api.tailverse.app/", "a.jpg"); got != "https://api.tailverse.app/uploads/a.jpg" {
		t.Errorf("fileUrl = %q", got)
	}
	r.Header.Set("X-Forwarded-Proto", "https")
	if got := fileUrl(r, "", "a.jpg"); got != "https://127.0.0.1:8080/uploads/a.jpg" {
		t.Errorf("fileUrl behind proxy = %q", got)
	}
}
