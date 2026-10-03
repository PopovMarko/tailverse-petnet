package core_domain

import (
	"errors"
	"testing"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

func ptr[T any](v T) *T { return &v }

func TestNewSpotPickerQueryRadius(t *testing.T) {
	lat, lng := ptr(47.905549), ptr(33.390509)

	cases := []struct {
		name    string
		radius  *int
		want    int
		wantErr bool
	}{
		{name: "defaults to 500", radius: nil, want: 500},
		{name: "smaller radius", radius: ptr(120), want: 120},
		{name: "exactly the cap", radius: ptr(500), want: 500},
		{name: "above the cap", radius: ptr(501), wantErr: true},
		{name: "map radius is not allowed", radius: ptr(2000), wantErr: true},
		{name: "zero", radius: ptr(0), wantErr: true},
		{name: "negative", radius: ptr(-10), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewSpotPickerQuery(lat, lng, tc.radius)
			if tc.wantErr {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("err = %v, want ErrInvalidArgument", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewSpotPickerQuery: %v", err)
			}
			if query.RadiusM != tc.want {
				t.Errorf("radius = %d, want %d", query.RadiusM, tc.want)
			}
		})
	}
}

func TestNewSpotPickerQueryRequiresValidCenter(t *testing.T) {
	if _, err := NewSpotPickerQuery(nil, ptr(33.39), nil); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("missing lat: err = %v, want ErrInvalidArgument", err)
	}
	if _, err := NewSpotPickerQuery(ptr(91.0), ptr(33.39), nil); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("lat out of range: err = %v, want ErrInvalidArgument", err)
	}
}

func TestNewNearbyQueryKeepsMapLimits(t *testing.T) {
	query, err := NewNearbyQuery(ptr(47.9), ptr(33.39), nil)
	if err != nil || query.RadiusM != DefaultRadiusM {
		t.Fatalf("NewNearbyQuery default = %d, %v; want %d", query.RadiusM, err, DefaultRadiusM)
	}
	if _, err := NewNearbyQuery(ptr(47.9), ptr(33.39), ptr(MaxRadiusM)); err != nil {
		t.Errorf("NewNearbyQuery(%d): %v", MaxRadiusM, err)
	}
}
