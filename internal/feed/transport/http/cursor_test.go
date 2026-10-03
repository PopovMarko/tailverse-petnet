package feed_transport_http

import (
	"errors"
	"testing"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

func TestCursorRoundTrip(t *testing.T) {
	cursor := core_domain.PostCursor{
		CreatedAt: time.Date(2026, 9, 28, 10, 11, 12, 345678000, time.UTC),
		Id:        "0b6c3a5e-51c4-4d53-9f42-2a0f1a4f9c11",
	}

	decoded, err := DecodeCursor(EncodeCursor(cursor))
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if !decoded.CreatedAt.Equal(cursor.CreatedAt) || decoded.Id != cursor.Id {
		t.Errorf("decoded %+v, want %+v (microseconds must survive)", decoded, cursor)
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	for _, raw := range []string{"not-base64!", "bm8tc2VwYXJhdG9y", EncodeCursor(core_domain.PostCursor{Id: "not-a-uuid"})} {
		if _, err := DecodeCursor(raw); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("DecodeCursor(%q) err = %v, want ErrInvalidArgument", raw, err)
		}
	}
}
