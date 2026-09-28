package announcement_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *AnnouncementHttpHandler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CreateAnnouncement handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("CreateAnnouncement handler: no owner", err)
		return
	}

	announcementDto := AnnouncementRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &announcementDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	announcement, err := h.announcementService.CreateAnnouncement(ctx, ownerId, DtoToDomain(announcementDto))
	if err != nil {
		httpResponseHandler.ErrorResponse("CreateAnnouncement handler: announcement service", err)
		return
	}

	httpResponseHandler.JsonResponse(DomainToDto(announcement), http.StatusCreated)
}
