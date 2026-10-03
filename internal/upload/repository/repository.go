package upload_repository

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

// DiskRepository keeps uploaded files in one directory on the local disk.
type DiskRepository struct {
	dir string
}

// NewDiskRepository creates dir if it does not exist yet.
func NewDiskRepository(dir string) (*DiskRepository, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create uploads dir %s: %w", dir, err)
	}
	return &DiskRepository{dir: dir}, nil
}

// Save writes the file atomically: a temporary file in the same directory is renamed into place.
func (r *DiskRepository) Save(_ context.Context, name string, content []byte) error {
	tmp, err := os.CreateTemp(r.dir, ".upload-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(r.dir, name)); err != nil {
		return fmt.Errorf("rename upload %s: %w", name, err)
	}
	return nil
}

// Open expects a plain file name (validated by the service), never a path.
func (r *DiskRepository) Open(_ context.Context, name string) (core_domain.StoredFile, error) {
	file, err := os.Open(filepath.Join(r.dir, filepath.Base(name)))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return core_domain.StoredFile{}, fmt.Errorf("upload %s: %w", name, core_errors.ErrNotFound)
		}
		return core_domain.StoredFile{}, fmt.Errorf("open upload %s: %w", name, err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return core_domain.StoredFile{}, fmt.Errorf("stat upload %s: %w", name, err)
	}
	return core_domain.StoredFile{Name: name, Size: info.Size(), ModTime: info.ModTime(), Content: file}, nil
}
