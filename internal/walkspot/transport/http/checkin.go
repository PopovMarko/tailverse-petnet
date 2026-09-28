package walkspot_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// CheckIn marks the owner's pet as present at the spot. Manual and geolocation-based check-ins use the same endpoint.
func (h *WalkSpotHttpHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CheckIn handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("CheckIn handler: no owner", err)
		return
	}

	spotId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("CheckIn handler: invalid spot id", err)
		return
	}

	checkInDto := CheckInRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &checkInDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	entry, err := h.walkSpotService.CheckIn(ctx, ownerId, spotId, checkInDto.PetId)
	if err != nil {
		httpResponseHandler.ErrorResponse("CheckIn handler: walk spot service", err)
		return
	}

	httpResponseHandler.JsonResponse(PresenceToDto(entry), http.StatusOK)
}
