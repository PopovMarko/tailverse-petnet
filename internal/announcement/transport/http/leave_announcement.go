package announcement_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// LeaveAnnouncement handles DELETE /announcements/{id}/join.
func (h *AnnouncementHttpHandler) LeaveAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("LeaveAnnouncement handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("LeaveAnnouncement handler: no owner", err)
		return
	}

	announcementId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("LeaveAnnouncement handler: invalid announcement id", err)
		return
	}

	leaveDto := JoinRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &leaveDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	if err := h.announcementService.Leave(ctx, ownerId, announcementId, leaveDto.PetId); err != nil {
		httpResponseHandler.ErrorResponse("LeaveAnnouncement handler: announcement service", err)
		return
	}

	httpResponseHandler.NoContentResponse()
}
