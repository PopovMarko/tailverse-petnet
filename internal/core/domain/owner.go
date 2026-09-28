package core_domain

import "time"

type Owner struct {
	Id              string
	Email           string
	PasswordHash    string
	Nickname        string
	Gender          string
	AvatarUrl       string
	IsProfilePublic bool
	CreatedAt       time.Time
}
