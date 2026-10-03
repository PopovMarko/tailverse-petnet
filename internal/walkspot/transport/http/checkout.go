package walkspot_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// CheckOut handles DELETE /walkspots/{id}/checkin: the pet has left the spot.
func (h *WalkSpotHttpHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CheckOut handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("CheckOut handler: no owner", err)
		return
	}

	spotId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("CheckOut handler: invalid spot id", err)
		return
	}

	checkOutDto := CheckInRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &checkOutDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	if err := h.walkSpotService.CheckOut(ctx, ownerId, spotId, checkOutDto.PetId); err != nil {
		httpResponseHandler.ErrorResponse("CheckOut handler: walk spot service", err)
		return
	}

	httpResponseHandler.NoContentResponse()
}
