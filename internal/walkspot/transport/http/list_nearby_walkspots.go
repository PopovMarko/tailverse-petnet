package walkspot_transport_http

import (
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// ListNearbyWalkSpots handles GET /walkspots/nearby?lat=&lng=&radius_m= — the spot picker of the "Иду гулять" form.
// radius_m defaults to 500 and may not exceed 500; spots come closest first with distance_m.
func (h *WalkSpotHttpHandler) ListNearbyWalkSpots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("ListNearbyWalkSpots handler called")

	query, err := spotPickerQueryFromRequest(r)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListNearbyWalkSpots handler: invalid query", err)
		return
	}

	spots, err := h.walkSpotService.ListForPicker(ctx, query)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListNearbyWalkSpots handler: walk spot service", err)
		return
	}

	spotsDto := make([]NearbyWalkSpotDto, len(spots))
	for i, spot := range spots {
		spotsDto[i] = NearbySummaryToDto(spot)
	}
	httpResponseHandler.JsonResponse(NearbyWalkSpotsResponseDto{Spots: spotsDto}, http.StatusOK)
}

func spotPickerQueryFromRequest(r *http.Request) (core_domain.NearbyQuery, error) {
	lat, err := core_http_utils.GetFloatQueryParam(r, "lat")
	if err != nil {
		return core_domain.NearbyQuery{}, err
	}
	lng, err := core_http_utils.GetFloatQueryParam(r, "lng")
	if err != nil {
		return core_domain.NearbyQuery{}, err
	}
	radius, err := core_http_utils.GetIntQueryParam(r, "radius_m")
	if err != nil {
		return core_domain.NearbyQuery{}, err
	}
	return core_domain.NewSpotPickerQuery(lat, lng, radius)
}
