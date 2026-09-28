package core_http_response

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

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
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriter) GetStatusCode() int {
	if w.status == UninitializedStatus {
		return http.StatusOK
	}
	return w.status
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (w *ResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Hijack is required for the WebSocket upgrade to pass through the Trace middleware.
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying response writer does not support hijacking")
	}
	w.status = http.StatusSwitchingProtocols
	return hijacker.Hijack()
}
