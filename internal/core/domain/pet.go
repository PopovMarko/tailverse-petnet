package core_domain

import "time"

type Pet struct {
	Id             string
	OwnerId        string
	Name           string
	Breed          string
	Species        string
	BirthDate      time.Time
	ApproxLocation string
	ApproxAddress  string
	CreatedAt      time.Time
}
