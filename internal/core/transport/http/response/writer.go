package core_http_response

import "net/http"

var UninitializedStatus = -1

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         UninitializedStatus,
	}
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.WriteHeader(status)
	w.status = status
}

func (w *ResponseWriter) GetStatusCode() int {
	if w.status == UninitializedStatus {
		return http.StatusOK
	}
	return w.status
}
