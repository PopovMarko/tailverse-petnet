package ws

import (
	"encoding/json"
	"fmt"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type spotUpdateMessage struct {
	Type         string `json:"type"`
	SpotId       string `json:"spot_id"`
	PresentCount int    `json:"present_count"`
}

type pointDto struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type announcementDto struct {
	Id          string    `json:"id"`
	PetId       string    `json:"pet_id"`
	SpotId      *string   `json:"spot_id"`
	CustomPoint *pointDto `json:"custom_point"`
	StartsAt    time.Time `json:"starts_at"`
	DurationMin int       `json:"duration_min"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type announcementCreatedMessage struct {
	Type         string          `json:"type"`
	Announcement announcementDto `json:"announcement"`
}

// clientMessage is what a client may send; only "location_update" is defined so far.
type clientMessage struct {
	Type string  `json:"type"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// encodeEvent converts a domain event into the JSON message sent to clients.
func encodeEvent(event any) ([]byte, error) {
	switch e := event.(type) {
	case core_domain.SpotUpdatedEvent:
		return json.Marshal(spotUpdateMessage{Type: "spot_update", SpotId: e.SpotId, PresentCount: e.PresentCount})
	case core_domain.AnnouncementCreatedEvent:
		a := e.Announcement
		dto := announcementDto{
			Id:          a.Id,
			PetId:       a.PetId,
			SpotId:      a.SpotId,
			StartsAt:    a.StartsAt,
			DurationMin: a.DurationMin,
			Status:      a.Status,
			CreatedAt:   a.CreatedAt,
		}
		if a.CustomPoint != nil {
			dto.CustomPoint = &pointDto{Lat: a.CustomPoint.Lat, Lng: a.CustomPoint.Lng}
		}
		return json.Marshal(announcementCreatedMessage{Type: "announcement_created", Announcement: dto})
	default:
		return nil, fmt.Errorf("unknown event type %T", event)
	}
}
