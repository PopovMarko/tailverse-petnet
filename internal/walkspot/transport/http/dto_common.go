package walkspot_transport_http

import (
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type WalkSpotSummaryDto struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Tags         []string `json:"tags"`
	PresentCount int      `json:"present_count"`
}

type WalkSpotsListResponseDto struct {
	Spots []WalkSpotSummaryDto `json:"spots"`
}

type NearbyWalkSpotDto struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Tags         []string `json:"tags"`
	PresentCount int      `json:"present_count"`
	DistanceM    int      `json:"distance_m"`
}

type NearbyWalkSpotsResponseDto struct {
	Spots []NearbyWalkSpotDto `json:"spots"`
}

type PresentPetDto struct {
	PetId         string    `json:"pet_id"`
	PetName       string    `json:"pet_name"`
	OwnerNickname string    `json:"owner_nickname"`
	CheckedInAt   time.Time `json:"checked_in_at"`
}

type WalkSpotDetailsDto struct {
	Id      string          `json:"id"`
	Name    string          `json:"name"`
	Lat     float64         `json:"lat"`
	Lng     float64         `json:"lng"`
	Tags    []string        `json:"tags"`
	Present []PresentPetDto `json:"present"`
}

type CheckInRequestDto struct {
	PetId string `json:"pet_id" validate:"required,uuid"`
}

type CheckInResponseDto struct {
	SpotId      string    `json:"spot_id"`
	PetId       string    `json:"pet_id"`
	CheckedInAt time.Time `json:"checked_in_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func SummaryToDto(spot core_domain.WalkSpotSummary) WalkSpotSummaryDto {
	return WalkSpotSummaryDto{
		Id:           spot.Id,
		Name:         spot.Name,
		Lat:          spot.Location.Lat,
		Lng:          spot.Location.Lng,
		Tags:         nonNilTags(spot.Tags),
		PresentCount: spot.PresentCount,
	}
}

func NearbySummaryToDto(spot core_domain.NearbyWalkSpotSummary) NearbyWalkSpotDto {
	return NearbyWalkSpotDto{
		Id:           spot.Id,
		Name:         spot.Name,
		Lat:          spot.Location.Lat,
		Lng:          spot.Location.Lng,
		Tags:         nonNilTags(spot.Tags),
		PresentCount: spot.PresentCount,
		DistanceM:    spot.DistanceM,
	}
}

func DetailsToDto(spot core_domain.WalkSpotDetails) WalkSpotDetailsDto {
	present := make([]PresentPetDto, len(spot.Present))
	for i, pet := range spot.Present {
		present[i] = PresentPetDto{
			PetId:         pet.PetId,
			PetName:       pet.PetName,
			OwnerNickname: pet.OwnerNickname,
			CheckedInAt:   pet.CheckedInAt,
		}
	}
	return WalkSpotDetailsDto{
		Id:      spot.Id,
		Name:    spot.Name,
		Lat:     spot.Location.Lat,
		Lng:     spot.Location.Lng,
		Tags:    nonNilTags(spot.Tags),
		Present: present,
	}
}

func PresenceToDto(entry core_domain.PresenceEntry) CheckInResponseDto {
	return CheckInResponseDto{
		SpotId:      entry.SpotId,
		PetId:       entry.PetId,
		CheckedInAt: entry.CheckedInAt,
		ExpiresAt:   entry.ExpiresAt,
	}
}

func nonNilTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
