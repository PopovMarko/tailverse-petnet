package core_domain

import "time"

const (
	AnnouncementStatusActive    = "active"
	AnnouncementStatusCancelled = "cancelled"
	AnnouncementStatusFinished  = "finished"
)

// Announcement is a "Иду гулять" post: a pet goes for a walk at a spot or a free-form point.
// Exactly one of SpotId and CustomPoint is set.
type Announcement struct {
	Id          string
	PetId       string
	SpotId      *string
	CustomPoint *GeoPoint
	StartsAt    time.Time
	DurationMin int
	Status      string
	CreatedAt   time.Time
}

type Participant struct {
	PetCard
	JoinedAt time.Time
}

type AnnouncementDetails struct {
	Announcement
	Participants []Participant
}

type AnnouncementFilter struct {
	Nearby NearbyQuery
	From   *time.Time
	To     *time.Time
}

type Participation struct {
	AnnouncementId string
	PetId          string
	JoinedAt       time.Time
}
