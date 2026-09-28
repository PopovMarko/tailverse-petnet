package feed_transport_http

import (
	"context"
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type FeedService interface {
	CreatePost(ctx context.Context, ownerId string, post core_domain.Post) (core_domain.Post, error)
	GetPost(ctx context.Context, id string) (core_domain.Post, error)
	ListPosts(ctx context.Context, filter core_domain.PostFilter) (core_domain.PostPage, error)
	DeletePost(ctx context.Context, ownerId string, id string) error
}

type FeedHttpHandler struct {
	feedService FeedService
}

func NewFeedHttpHandler(feedService FeedService) *FeedHttpHandler {
	return &FeedHttpHandler{
		feedService: feedService,
	}
}

func NewPostsRouter(h *FeedHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", h.ListPosts)
	r.Get("/{id}", h.GetPost)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", h.CreatePost)
		r.Delete("/{id}", h.DeletePost)
	})

	return r
}
