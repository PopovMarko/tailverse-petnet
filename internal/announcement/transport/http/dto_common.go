package announcement_transport_http

import (
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type PointDto struct {
	Lat float64 `json:"lat" validate:"gte=-90,lte=90"`
	Lng float64 `json:"lng" validate:"gte=-180,lte=180"`
}

type AnnouncementRequestDto struct {
	PetId       string    `json:"pet_id" validate:"required,uuid"`
	SpotId      *string   `json:"spot_id" validate:"omitempty,uuid"`
	CustomPoint *PointDto `json:"custom_point"`
	StartsAt    time.Time `json:"starts_at" validate:"required"`
	DurationMin int       `json:"duration_min" validate:"required,gt=0"`
}

type AnnouncementResponseDto struct {
	Id          string    `json:"id"`
	PetId       string    `json:"pet_id"`
	SpotId      *string   `json:"spot_id"`
	CustomPoint *PointDto `json:"custom_point"`
	StartsAt    time.Time `json:"starts_at"`
	DurationMin int       `json:"duration_min"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type AnnouncementsListResponseDto struct {
	Announcements []AnnouncementResponseDto `json:"announcements"`
}

type ParticipantDto struct {
	PetId         string    `json:"pet_id"`
	PetName       string    `json:"pet_name"`
	OwnerNickname string    `json:"owner_nickname"`
	JoinedAt      time.Time `json:"joined_at"`
}

type AnnouncementDetailsDto struct {
	AnnouncementResponseDto
	Participants []ParticipantDto `json:"participants"`
}

type JoinRequestDto struct {
	PetId string `json:"pet_id" validate:"required,uuid"`
}

type JoinResponseDto struct {
	AnnouncementId string    `json:"announcement_id"`
	PetId          string    `json:"pet_id"`
	JoinedAt       time.Time `json:"joined_at"`
}

func DtoToDomain(dto AnnouncementRequestDto) core_domain.Announcement {
	announcement := core_domain.Announcement{
		PetId:       dto.PetId,
		SpotId:      dto.SpotId,
		StartsAt:    dto.StartsAt,
		DurationMin: dto.DurationMin,
	}
	if dto.CustomPoint != nil {
		announcement.CustomPoint = &core_domain.GeoPoint{Lat: dto.CustomPoint.Lat, Lng: dto.CustomPoint.Lng}
	}
	return announcement
}

func DomainToDto(a core_domain.Announcement) AnnouncementResponseDto {
	dto := AnnouncementResponseDto{
		Id:          a.Id,
		PetId:       a.PetId,
		SpotId:      a.SpotId,
		StartsAt:    a.StartsAt,
		DurationMin: a.DurationMin,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
	}
	if a.CustomPoint != nil {
		dto.CustomPoint = &PointDto{Lat: a.CustomPoint.Lat, Lng: a.CustomPoint.Lng}
	}
	return dto
}

func DetailsToDto(details core_domain.AnnouncementDetails) AnnouncementDetailsDto {
	participants := make([]ParticipantDto, len(details.Participants))
	for i, p := range details.Participants {
		participants[i] = ParticipantDto{
			PetId:         p.PetId,
			PetName:       p.PetName,
			OwnerNickname: p.OwnerNickname,
			JoinedAt:      p.JoinedAt,
		}
	}
	return AnnouncementDetailsDto{
		AnnouncementResponseDto: DomainToDto(details.Announcement),
		Participants:            participants,
	}
}

func ParticipationToDto(p core_domain.Participation) JoinResponseDto {
	return JoinResponseDto{
		AnnouncementId: p.AnnouncementId,
		PetId:          p.PetId,
		JoinedAt:       p.JoinedAt,
	}
}
