package core_domain

import "time"

// ServiceOffer is an entry in the "Услуги" section (grooming, dog walking, ...).
// Named ServiceOffer so it is not confused with the service layer.
type ServiceOffer struct {
	Id              string
	ProviderOwnerId string
	Title           string
	Description     string
	Category        string
	Location        GeoPoint
	Price           float64
	CreatedAt       time.Time
}

type ServiceOfferFilter struct {
	Nearby   NearbyQuery
	Category *string
}
