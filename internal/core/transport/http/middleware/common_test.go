package core_http_middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedactedURLHidesToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/ws/presence?token=eyJhbGciOiJIUzI1NiJ9.secret&x=1", nil)

	got := redactedURL(r)
	if strings.Contains(got, "eyJ") || !strings.Contains(got, "token=REDACTED") || !strings.Contains(got, "x=1") {
		t.Errorf("redactedURL = %q, want the token replaced and other params kept", got)
	}
	if r.URL.Query().Get("token") == "REDACTED" {
		t.Error("redactedURL modified the original request URL")
	}
}

func TestRedactedURLWithoutToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/walkspots?lat=55.7&lng=37.6", nil)
	if got := redactedURL(r); got != "/api/v1/walkspots?lat=55.7&lng=37.6" {
		t.Errorf("redactedURL = %q, want the URL unchanged", got)
	}
}
