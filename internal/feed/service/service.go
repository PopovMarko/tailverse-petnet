package feed_service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 50
	MaxPhotos       = 10
	MaxTextLength   = 5000
)

type PostRepository interface {
	CreatePost(ctx context.Context, post core_domain.Post) (core_domain.Post, error)
	GetPost(ctx context.Context, id string) (core_domain.Post, error)
	ListPosts(ctx context.Context, filter core_domain.PostFilter) ([]core_domain.Post, error)
	DeletePost(ctx context.Context, id string) error
}

type PetService interface {
	EnsurePetOwner(ctx context.Context, ownerId string, petId string) error
}

type FeedService struct {
	repository PostRepository
	pets       PetService
}

func NewFeedService(repository PostRepository, pets PetService) *FeedService {
	return &FeedService{
		repository: repository,
		pets:       pets,
	}
}

func (s *FeedService) CreatePost(ctx context.Context, ownerId string, post core_domain.Post) (core_domain.Post, error) {
	post.Text = strings.TrimSpace(post.Text)
	if post.PhotoUrls == nil {
		post.PhotoUrls = []string{}
	}
	if err := validatePost(post); err != nil {
		return core_domain.Post{}, err
	}
	if err := s.pets.EnsurePetOwner(ctx, ownerId, post.PetId); err != nil {
		return core_domain.Post{}, fmt.Errorf("check pet owner: %w", err)
	}

	created, err := s.repository.CreatePost(ctx, post)
	if err != nil {
		return core_domain.Post{}, fmt.Errorf("create post: %w", err)
	}
	return created, nil
}

func (s *FeedService) GetPost(ctx context.Context, id string) (core_domain.Post, error) {
	post, err := s.repository.GetPost(ctx, id)
	if err != nil {
		return core_domain.Post{}, fmt.Errorf("get post: %w", err)
	}
	return post, nil
}

// ListPosts returns a page of posts, newest first. NextCursor is nil on the last page.
func (s *FeedService) ListPosts(ctx context.Context, filter core_domain.PostFilter) (core_domain.PostPage, error) {
	if filter.Limit == 0 {
		filter.Limit = DefaultPageSize
	}
	if filter.Limit < 0 || filter.Limit > MaxPageSize {
		return core_domain.PostPage{}, fmt.Errorf("limit must be in [1, %d]: %w", MaxPageSize, core_errors.ErrInvalidArgument)
	}

	// Ask for one extra post to know whether there is a next page.
	pageSize := filter.Limit
	filter.Limit++
	posts, err := s.repository.ListPosts(ctx, filter)
	if err != nil {
		return core_domain.PostPage{}, fmt.Errorf("list posts: %w", err)
	}

	page := core_domain.PostPage{Posts: posts}
	if len(posts) > pageSize {
		page.Posts = posts[:pageSize]
		last := page.Posts[pageSize-1]
		page.NextCursor = &core_domain.PostCursor{CreatedAt: last.CreatedAt, Id: last.Id}
	}
	return page, nil
}

// DeletePost lets only the owner of the post's pet delete it.
func (s *FeedService) DeletePost(ctx context.Context, ownerId string, id string) error {
	post, err := s.repository.GetPost(ctx, id)
	if err != nil {
		return fmt.Errorf("get post: %w", err)
	}
	if err := s.pets.EnsurePetOwner(ctx, ownerId, post.PetId); err != nil {
		return fmt.Errorf("check post author: %w", err)
	}
	if err := s.repository.DeletePost(ctx, id); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func validatePost(post core_domain.Post) error {
	if post.Text == "" && len(post.PhotoUrls) == 0 {
		return fmt.Errorf("a post needs text or at least one photo: %w", core_errors.ErrInvalidArgument)
	}
	if len([]rune(post.Text)) > MaxTextLength {
		return fmt.Errorf("text is longer than %d characters: %w", MaxTextLength, core_errors.ErrInvalidArgument)
	}
	if len(post.PhotoUrls) > MaxPhotos {
		return fmt.Errorf("at most %d photos per post: %w", MaxPhotos, core_errors.ErrInvalidArgument)
	}
	for _, photoUrl := range post.PhotoUrls {
		parsed, err := url.Parse(photoUrl)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("photo url %q is not an http(s) url: %w", photoUrl, core_errors.ErrInvalidArgument)
		}
	}
	return nil
}
