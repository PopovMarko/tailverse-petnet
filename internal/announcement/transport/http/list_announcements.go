package announcement_transport_http

import (
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// ListAnnouncements handles GET /announcements?lat=&lng=&radius_m=&from=&to= (from/to are RFC 3339).
func (h *AnnouncementHttpHandler) ListAnnouncements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("ListAnnouncements handler called")

	filter, err := filterFromRequest(r)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListAnnouncements handler: invalid query", err)
		return
	}

	announcements, err := h.announcementService.ListAnnouncements(ctx, filter)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListAnnouncements handler: announcement service", err)
		return
	}

	announcementsDto := make([]AnnouncementResponseDto, len(announcements))
	for i, announcement := range announcements {
		announcementsDto[i] = DomainToDto(announcement)
	}
	httpResponseHandler.JsonResponse(AnnouncementsListResponseDto{Announcements: announcementsDto}, http.StatusOK)
}

func filterFromRequest(r *http.Request) (core_domain.AnnouncementFilter, error) {
	lat, err := core_http_utils.GetFloatQueryParam(r, "lat")
	if err != nil {
		return core_domain.AnnouncementFilter{}, err
	}
	lng, err := core_http_utils.GetFloatQueryParam(r, "lng")
	if err != nil {
		return core_domain.AnnouncementFilter{}, err
	}
	radius, err := core_http_utils.GetIntQueryParam(r, "radius_m")
	if err != nil {
		return core_domain.AnnouncementFilter{}, err
	}
	nearby, err := core_domain.NewNearbyQuery(lat, lng, radius)
	if err != nil {
		return core_domain.AnnouncementFilter{}, err
	}

	from, err := core_http_utils.GetTimeQueryParam(r, "from")
	if err != nil {
		return core_domain.AnnouncementFilter{}, err
	}
	to, err := core_http_utils.GetTimeQueryParam(r, "to")
	if err != nil {
		return core_domain.AnnouncementFilter{}, err
	}

	return core_domain.AnnouncementFilter{Nearby: nearby, From: from, To: to}, nil
}
