package core_domain

import (
	"fmt"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

const (
	DefaultRadiusM = 2000
	MaxRadiusM     = 50000
)

type GeoPoint struct {
	Lat float64
	Lng float64
}

func (p GeoPoint) Validate() error {
	if p.Lat < -90 || p.Lat > 90 || p.Lng < -180 || p.Lng > 180 {
		return fmt.Errorf("coordinates out of range (lat %v, lng %v): %w", p.Lat, p.Lng, core_errors.ErrInvalidArgument)
	}
	return nil
}

// NearbyQuery is a "points within RadiusM metres of Center" search.
type NearbyQuery struct {
	Center  GeoPoint
	RadiusM int
}

func NewNearbyQuery(lat, lng *float64, radiusM *int) (NearbyQuery, error) {
	if lat == nil || lng == nil {
		return NearbyQuery{}, fmt.Errorf("lat and lng are required: %w", core_errors.ErrInvalidArgument)
	}
	query := NearbyQuery{Center: GeoPoint{Lat: *lat, Lng: *lng}, RadiusM: DefaultRadiusM}
	if radiusM != nil {
		query.RadiusM = *radiusM
	}
	if query.RadiusM <= 0 || query.RadiusM > MaxRadiusM {
		return NearbyQuery{}, fmt.Errorf("radius_m must be in (0, %d]: %w", MaxRadiusM, core_errors.ErrInvalidArgument)
	}
	if err := query.Center.Validate(); err != nil {
		return NearbyQuery{}, err
	}
	return query, nil
}
