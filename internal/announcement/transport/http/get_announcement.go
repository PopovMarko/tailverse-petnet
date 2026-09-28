package announcement_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *AnnouncementHttpHandler) GetAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetAnnouncement handler called")

	announcementId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("GetAnnouncement handler: invalid announcement id", err)
		return
	}

	details, err := h.announcementService.GetAnnouncement(ctx, announcementId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetAnnouncement handler: announcement service", err)
		return
	}

	httpResponseHandler.JsonResponse(DetailsToDto(details), http.StatusOK)
}
