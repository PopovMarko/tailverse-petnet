package upload_service

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

type fakeFileRepository struct {
	files map[string][]byte
}

func newFakeFileRepository() *fakeFileRepository {
	return &fakeFileRepository{files: map[string][]byte{}}
}

func (f *fakeFileRepository) Save(_ context.Context, name string, content []byte) error {
	f.files[name] = content
	return nil
}

func (f *fakeFileRepository) Open(_ context.Context, name string) (core_domain.StoredFile, error) {
	content, ok := f.files[name]
	if !ok {
		return core_domain.StoredFile{}, core_errors.ErrNotFound
	}
	return core_domain.StoredFile{Name: name, Size: int64(len(content)), Content: nopCloser{bytes.NewReader(content)}}, nil
}

type nopCloser struct{ *bytes.Reader }

func (nopCloser) Close() error { return nil }

var (
	jpegHeader = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
	pngHeader  = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	webpHeader = []byte("RIFF\x24\x00\x00\x00WEBPVP8 ")
	heicHeader = []byte("\x00\x00\x00\x18ftypheic\x00\x00\x00\x00mif1heic")
	gifHeader  = []byte("GIF89a\x01\x00\x01\x00")
)

func TestUploadAcceptsSupportedImages(t *testing.T) {
	cases := []struct {
		header      []byte
		contentType string
		extension   string
	}{
		{jpegHeader, "image/jpeg", ".jpg"},
		{pngHeader, "image/png", ".png"},
		{webpHeader, "image/webp", ".webp"},
		{heicHeader, "image/heic", ".heic"},
	}
	for _, tc := range cases {
		t.Run(tc.contentType, func(t *testing.T) {
			repository := newFakeFileRepository()
			service := NewUploadService(repository)

			upload, err := service.Upload(context.Background(), bytes.NewReader(append(tc.header, "pixels"...)))
			if err != nil {
				t.Fatalf("Upload: %v", err)
			}
			if upload.ContentType != tc.contentType {
				t.Errorf("content type = %q, want %q", upload.ContentType, tc.contentType)
			}
			if !regexp.MustCompile(`^[0-9a-f]{32}\` + tc.extension + `$`).MatchString(upload.Name) {
				t.Errorf("name = %q, want 32 hex chars + %s", upload.Name, tc.extension)
			}
			if _, ok := repository.files[upload.Name]; !ok {
				t.Errorf("file %q was not saved", upload.Name)
			}
		})
	}
}

func TestUploadGeneratesUniqueNames(t *testing.T) {
	service := NewUploadService(newFakeFileRepository())
	first, _ := service.Upload(context.Background(), bytes.NewReader(jpegHeader))
	second, _ := service.Upload(context.Background(), bytes.NewReader(jpegHeader))
	if first.Name == second.Name {
		t.Errorf("two uploads got the same name %q", first.Name)
	}
}

func TestUploadRejectsUnsupportedTypes(t *testing.T) {
	for name, content := range map[string][]byte{
		"gif":         gifHeader,
		"text":        []byte("hello, this is not an image"),
		"svg":         []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`),
		"mp4":         []byte("\x00\x00\x00\x18ftypisom\x00\x00\x02\x00"),
		"tiny":        {0xFF, 0xD8},
		"html as png": []byte("<html>\x89PNG"),
	} {
		repository := newFakeFileRepository()
		_, err := NewUploadService(repository).Upload(context.Background(), bytes.NewReader(content))
		if !errors.Is(err, core_errors.ErrUnsupportedMediaType) {
			t.Errorf("%s: err = %v, want ErrUnsupportedMediaType", name, err)
		}
		if len(repository.files) != 0 {
			t.Errorf("%s: rejected file was saved", name)
		}
	}
}

func TestUploadSizeLimit(t *testing.T) {
	atLimit := append(append([]byte{}, jpegHeader...), bytes.Repeat([]byte{0}, MaxUploadSize-len(jpegHeader))...)
	if _, err := NewUploadService(newFakeFileRepository()).Upload(context.Background(), bytes.NewReader(atLimit)); err != nil {
		t.Errorf("exactly 10 MB: %v", err)
	}

	overLimit := append(atLimit, 0)
	repository := newFakeFileRepository()
	if _, err := NewUploadService(repository).Upload(context.Background(), bytes.NewReader(overLimit)); !errors.Is(err, core_errors.ErrTooLarge) {
		t.Errorf("10 MB + 1 byte: err = %v, want ErrTooLarge", err)
	}
	if len(repository.files) != 0 {
		t.Errorf("too large file was saved")
	}
}

func TestUploadRejectsEmptyFile(t *testing.T) {
	if _, err := NewUploadService(newFakeFileRepository()).Upload(context.Background(), strings.NewReader("")); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestOpenOnlyServesGeneratedNames(t *testing.T) {
	repository := newFakeFileRepository()
	service := NewUploadService(repository)
	upload, err := service.Upload(context.Background(), bytes.NewReader(pngHeader))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	file, err := service.Open(context.Background(), upload.Name)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if file.ContentType != "image/png" {
		t.Errorf("content type = %q, want image/png", file.ContentType)
	}

	for _, name := range []string{"../.env", "..%2F.env", ".upload-123", "0123456789abcdef0123456789abcdef.exe", "ABCDEF0123456789abcdef0123456789.jpg", ""} {
		if _, err := service.Open(context.Background(), name); !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("Open(%q): err = %v, want ErrNotFound", name, err)
		}
	}
}
