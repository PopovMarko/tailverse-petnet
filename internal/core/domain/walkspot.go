package core_domain

import "time"

type WalkSpot struct {
	Id        string
	Name      string
	Location  GeoPoint
	Tags      []string
	CreatedAt time.Time
}

// PresenceEntry says that a pet is at a walk spot right now (stored in Redis with a TTL).
type PresenceEntry struct {
	SpotId      string
	PetId       string
	CheckedInAt time.Time
	ExpiresAt   time.Time
}

// PresentPet is a PresenceEntry enriched with the pet and owner names from Postgres.
type PresentPet struct {
	PetCard
	CheckedInAt time.Time
}

type WalkSpotDetails struct {
	WalkSpot
	Present []PresentPet
}

// WalkSpotSummary is a map marker: the spot plus how many pets are there now.
type WalkSpotSummary struct {
	WalkSpot
	PresentCount int
}

// WalkSpotWithDistance is a walk spot with its distance from the search centre, rounded to whole metres.
type WalkSpotWithDistance struct {
	WalkSpot
	DistanceM int
}

// NearbyWalkSpotSummary is a spot picker entry: the spot, how many pets are there now and how far away it is.
type NearbyWalkSpotSummary struct {
	WalkSpotSummary
	DistanceM int
}
