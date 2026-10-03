package upload_transport_http

import (
	"net/http"
	"strings"
)

// uploadFormField is the multipart field that carries the file.
const uploadFormField = "file"

type UploadResponseDto struct {
	Url string `json:"url"`
}

// fileUrl builds the absolute URL of an uploaded file. Without a configured base URL the request's own
// scheme and Host are used, so the URL works for whichever address the client reached the API at.
func fileUrl(r *http.Request, publicBaseUrl, name string) string {
	base := strings.TrimRight(publicBaseUrl, "/")
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		} else if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" || proto == "http" {
			scheme = proto
		}
		base = scheme + "://" + r.Host
	}
	return base + "/uploads/" + name
}
