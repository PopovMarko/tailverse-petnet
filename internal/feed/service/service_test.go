package feed_service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
)

// fakePostRepository keeps posts newest first, like the SQL ORDER BY created_at DESC, id DESC.
type fakePostRepository struct {
	posts []core_domain.Post
}

func (f *fakePostRepository) CreatePost(_ context.Context, post core_domain.Post) (core_domain.Post, error) {
	return post, nil
}

func (f *fakePostRepository) GetPost(_ context.Context, id string) (core_domain.Post, error) {
	for _, post := range f.posts {
		if post.Id == id {
			return post, nil
		}
	}
	return core_domain.Post{}, core_errors.ErrNotFound
}

func (f *fakePostRepository) ListPosts(_ context.Context, filter core_domain.PostFilter) ([]core_domain.Post, error) {
	var result []core_domain.Post
	for _, post := range f.posts {
		if filter.After != nil && !post.CreatedAt.Before(filter.After.CreatedAt) {
			continue
		}
		result = append(result, post)
		if len(result) == filter.Limit {
			break
		}
	}
	return result, nil
}

func (f *fakePostRepository) DeletePost(context.Context, string) error { return nil }

type fakePets struct{}

func (fakePets) EnsurePetOwner(_ context.Context, ownerId string, petId string) error {
	if ownerId != "owner-1" || petId != "pet-1" {
		return core_errors.ErrForbidden
	}
	return nil
}

func TestListPostsPaginates(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	repository := &fakePostRepository{}
	for i := 4; i >= 0; i-- { // newest first: post-4 … post-0
		repository.posts = append(repository.posts, core_domain.Post{Id: fmt.Sprintf("post-%d", i), CreatedAt: start.Add(time.Duration(i) * time.Hour)})
	}
	service := NewFeedService(repository, fakePets{})

	var seen []string
	filter := core_domain.PostFilter{Limit: 2}
	for pages := 0; ; pages++ {
		if pages > 5 {
			t.Fatal("pagination does not terminate")
		}
		page, err := service.ListPosts(context.Background(), filter)
		if err != nil {
			t.Fatalf("ListPosts: %v", err)
		}
		for _, post := range page.Posts {
			seen = append(seen, post.Id)
		}
		if page.NextCursor == nil {
			break
		}
		filter.After = page.NextCursor
	}

	want := []string{"post-4", "post-3", "post-2", "post-1", "post-0"}
	if fmt.Sprint(seen) != fmt.Sprint(want) {
		t.Errorf("pages = %v, want %v", seen, want)
	}
}

func TestListPostsLimit(t *testing.T) {
	service := NewFeedService(&fakePostRepository{}, fakePets{})

	for _, limit := range []int{-1, MaxPageSize + 1} {
		if _, err := service.ListPosts(context.Background(), core_domain.PostFilter{Limit: limit}); !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("limit %d: err = %v, want ErrInvalidArgument", limit, err)
		}
	}
	if _, err := service.ListPosts(context.Background(), core_domain.PostFilter{}); err != nil {
		t.Errorf("default limit: %v", err)
	}
}

func TestCreatePostValidation(t *testing.T) {
	service := NewFeedService(&fakePostRepository{}, fakePets{})
	ctx := context.Background()

	cases := []struct {
		name string
		post core_domain.Post
		want error
	}{
		{"text", core_domain.Post{PetId: "pet-1", Text: "Отличная прогулка"}, nil},
		{"photo only", core_domain.Post{PetId: "pet-1", PhotoUrls: []string{"https://cdn.example.com/a.jpg"}}, nil},
		{"empty", core_domain.Post{PetId: "pet-1", Text: "   "}, core_errors.ErrInvalidArgument},
		{"bad url", core_domain.Post{PetId: "pet-1", PhotoUrls: []string{"javascript:alert(1)"}}, core_errors.ErrInvalidArgument},
		{"too many photos", core_domain.Post{PetId: "pet-1", PhotoUrls: make([]string, MaxPhotos+1)}, core_errors.ErrInvalidArgument},
		{"someone else's pet", core_domain.Post{PetId: "pet-2", Text: "hi"}, core_errors.ErrForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CreatePost(ctx, "owner-1", tc.post)
			if tc.want == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDeletePostOnlyByAuthor(t *testing.T) {
	repository := &fakePostRepository{posts: []core_domain.Post{{Id: "post-1", PetId: "pet-1"}}}
	service := NewFeedService(repository, fakePets{})

	if err := service.DeletePost(context.Background(), "owner-2", "post-1"); !errors.Is(err, core_errors.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
	if err := service.DeletePost(context.Background(), "owner-1", "post-1"); err != nil {
		t.Errorf("author delete: %v", err)
	}
	if err := service.DeletePost(context.Background(), "owner-1", "post-404"); !errors.Is(err, core_errors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
