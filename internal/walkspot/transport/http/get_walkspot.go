package walkspot_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *WalkSpotHttpHandler) GetWalkSpot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetWalkSpot handler called")

	spotId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("GetWalkSpot handler: invalid spot id", err)
		return
	}

	spot, err := h.walkSpotService.GetWalkSpot(ctx, spotId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetWalkSpot handler: walk spot service", err)
		return
	}

	httpResponseHandler.JsonResponse(DetailsToDto(spot), http.StatusOK)
}
