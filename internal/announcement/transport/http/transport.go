package announcement_transport_http

import (
	"context"
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/go-chi/chi/v5"
)

type AnnouncementService interface {
	CreateAnnouncement(ctx context.Context, ownerId string, a core_domain.Announcement) (core_domain.Announcement, error)
	ListAnnouncements(ctx context.Context, filter core_domain.AnnouncementFilter) ([]core_domain.Announcement, error)
	GetAnnouncement(ctx context.Context, id string) (core_domain.AnnouncementDetails, error)
	Join(ctx context.Context, ownerId, announcementId, petId string) (core_domain.Participation, error)
	Leave(ctx context.Context, ownerId, announcementId, petId string) error
}

type AnnouncementHttpHandler struct {
	announcementService AnnouncementService
}

func NewAnnouncementHttpHandler(announcementService AnnouncementService) *AnnouncementHttpHandler {
	return &AnnouncementHttpHandler{
		announcementService: announcementService,
	}
}

func NewAnnouncementsRouter(h *AnnouncementHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", h.ListAnnouncements)
	r.Get("/{id}", h.GetAnnouncement)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", h.CreateAnnouncement)
		r.Post("/{id}/join", h.JoinAnnouncement)
		r.Delete("/{id}/join", h.LeaveAnnouncement)
	})

	return r
}
