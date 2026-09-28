package services_transport_http

import (
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
)

type ServiceRequestDto struct {
	Title       string   `json:"title" validate:"required,max=150"`
	Description string   `json:"description" validate:"max=5000"`
	Category    string   `json:"category" validate:"required,max=50"`
	Lat         *float64 `json:"lat" validate:"required,gte=-90,lte=90"`
	Lng         *float64 `json:"lng" validate:"required,gte=-180,lte=180"`
	Price       float64  `json:"price" validate:"gte=0"`
}

type ServiceResponseDto struct {
	Id              string    `json:"id"`
	ProviderOwnerId string    `json:"provider_owner_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	Price           float64   `json:"price"`
	CreatedAt       time.Time `json:"created_at"`
}

type ServicesListResponseDto struct {
	Services []ServiceResponseDto `json:"services"`
}

func DtoToDomain(dto ServiceRequestDto) core_domain.ServiceOffer {
	return core_domain.ServiceOffer{
		Title:       dto.Title,
		Description: dto.Description,
		Category:    dto.Category,
		Location:    core_domain.GeoPoint{Lat: *dto.Lat, Lng: *dto.Lng},
		Price:       dto.Price,
	}
}

func DomainToDto(offer core_domain.ServiceOffer) ServiceResponseDto {
	return ServiceResponseDto{
		Id:              offer.Id,
		ProviderOwnerId: offer.ProviderOwnerId,
		Title:           offer.Title,
		Description:     offer.Description,
		Category:        offer.Category,
		Lat:             offer.Location.Lat,
		Lng:             offer.Location.Lng,
		Price:           offer.Price,
		CreatedAt:       offer.CreatedAt,
	}
}
