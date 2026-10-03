package feed_transport_http

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"github.com/google/uuid"
)

type PostRequestDto struct {
	PetId     string   `json:"pet_id" validate:"required,uuid"`
	SpotId    *string  `json:"spot_id" validate:"omitempty,uuid"`
	Text      string   `json:"text"`
	PhotoUrls []string `json:"photo_urls"`
}

type PostResponseDto struct {
	Id        string    `json:"id"`
	PetId     string    `json:"pet_id"`
	SpotId    *string   `json:"spot_id"`
	Text      string    `json:"text"`
	PhotoUrls []string  `json:"photo_urls"`
	CreatedAt time.Time `json:"created_at"`
}

type PostsPageResponseDto struct {
	Posts      []PostResponseDto `json:"posts"`
	NextCursor *string           `json:"next_cursor"`
}

func DtoToDomain(dto PostRequestDto) core_domain.Post {
	return core_domain.Post{
		PetId:     dto.PetId,
		SpotId:    dto.SpotId,
		Text:      dto.Text,
		PhotoUrls: dto.PhotoUrls,
	}
}

func DomainToDto(post core_domain.Post) PostResponseDto {
	photoUrls := post.PhotoUrls
	if photoUrls == nil {
		photoUrls = []string{}
	}
	return PostResponseDto{
		Id:        post.Id,
		PetId:     post.PetId,
		SpotId:    post.SpotId,
		Text:      post.Text,
		PhotoUrls: photoUrls,
		CreatedAt: post.CreatedAt,
	}
}

func PageToDto(page core_domain.PostPage) PostsPageResponseDto {
	posts := make([]PostResponseDto, len(page.Posts))
	for i, post := range page.Posts {
		posts[i] = DomainToDto(post)
	}

	dto := PostsPageResponseDto{Posts: posts}
	if page.NextCursor != nil {
		cursor := EncodeCursor(*page.NextCursor)
		dto.NextCursor = &cursor
	}
	return dto
}

// EncodeCursor makes the opaque next_cursor string: base64url("<created_at RFC3339Nano>|<post id>").
func EncodeCursor(cursor core_domain.PostCursor) string {
	raw := cursor.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + cursor.Id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (core_domain.PostCursor, error) {
	invalid := fmt.Errorf("invalid cursor: %w", core_errors.ErrInvalidArgument)

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return core_domain.PostCursor{}, invalid
	}
	createdAtRaw, id, ok := strings.Cut(string(raw), "|")
	if !ok || uuid.Validate(id) != nil {
		return core_domain.PostCursor{}, invalid
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
	if err != nil {
		return core_domain.PostCursor{}, invalid
	}
	return core_domain.PostCursor{CreatedAt: createdAt, Id: id}, nil
}
