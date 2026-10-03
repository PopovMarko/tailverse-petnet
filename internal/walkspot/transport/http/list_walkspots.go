package walkspot_transport_http

import (
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// ListWalkSpots handles GET /walkspots?lat=&lng=&radius_m=
func (h *WalkSpotHttpHandler) ListWalkSpots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("ListWalkSpots handler called")

	query, err := nearbyQueryFromRequest(r)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListWalkSpots handler: invalid query", err)
		return
	}

	spots, err := h.walkSpotService.ListNearby(ctx, query)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListWalkSpots handler: walk spot service", err)
		return
	}

	spotsDto := make([]WalkSpotSummaryDto, len(spots))
	for i, spot := range spots {
		spotsDto[i] = SummaryToDto(spot)
	}
	httpResponseHandler.JsonResponse(WalkSpotsListResponseDto{Spots: spotsDto}, http.StatusOK)
}

func nearbyQueryFromRequest(r *http.Request) (core_domain.NearbyQuery, error) {
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
	return core_domain.NewNearbyQuery(lat, lng, radius)
}
