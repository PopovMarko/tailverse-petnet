package core_domain

import (
	"io"
	"time"
)

// Upload is a file the client has uploaded (an avatar or a post photo).
type Upload struct {
	Name        string // random file name, e.g. "3f9c…e1.jpg"; the public URL is /uploads/{Name}
	ContentType string
	Size        int64
}

// StoredFile is an uploaded file opened for reading. The caller must close Content.
type StoredFile struct {
	Name        string
	ContentType string
	Size        int64
	ModTime     time.Time
	Content     io.ReadSeekCloser
}
