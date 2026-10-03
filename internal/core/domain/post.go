package core_domain

import "time"

type Post struct {
	Id        string
	PetId     string
	SpotId    *string
	Text      string
	PhotoUrls []string
	CreatedAt time.Time
}

// PostCursor points at the last post of the previous page (keyset pagination by created_at, id).
type PostCursor struct {
	CreatedAt time.Time
	Id        string
}

type PostFilter struct {
	SpotId *string
	After  *PostCursor
	Limit  int
}

type PostPage struct {
	Posts      []Post
	NextCursor *PostCursor
}
