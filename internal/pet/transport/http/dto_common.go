package pet_transport_http

import "time"

type PetRequestDto struct {
	Name          string    `json:"name" required:"true"`
	Breed         string    `json:"breed" required:"true"`
	Species       string    `json:"species"`
	BirthDate     time.Time `json:"birth_date"`
	Age           int       `json:"age"`
	ApproxAddress string    `json:"approx_address" required:"true"`
}

type PetResponseDto struct {
	Id             string    `json:"id"`
	OwnerId        string    `json:"owner_ic"`
	Name           string    `json:"name"`
	Breed          string    `json:"breed"`
	Species        string    `json:"species"`
	BirthDate      time.Time `json:"birth_date"`
	ApproxLocation string    `json:"approx_location"`
	ApproxAddress  string    `json:"approx_address"`
	CreatedAt      time.Time `json:"created_at"`
}
