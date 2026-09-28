package services_transport_http

import (
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// ListServices handles GET /services?lat=&lng=&radius_m=&category=
func (h *ServicesHttpHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("ListServices handler called")

	filter, err := filterFromRequest(r)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListServices handler: invalid query", err)
		return
	}

	offers, err := h.serviceOfferService.ListNearby(ctx, filter)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListServices handler: service offer service", err)
		return
	}

	servicesDto := make([]ServiceResponseDto, len(offers))
	for i, offer := range offers {
		servicesDto[i] = DomainToDto(offer)
	}
	httpResponseHandler.JsonResponse(ServicesListResponseDto{Services: servicesDto}, http.StatusOK)
}

func filterFromRequest(r *http.Request) (core_domain.ServiceOfferFilter, error) {
	lat, err := core_http_utils.GetFloatQueryParam(r, "lat")
	if err != nil {
		return core_domain.ServiceOfferFilter{}, err
	}
	lng, err := core_http_utils.GetFloatQueryParam(r, "lng")
	if err != nil {
		return core_domain.ServiceOfferFilter{}, err
	}
	radius, err := core_http_utils.GetIntQueryParam(r, "radius_m")
	if err != nil {
		return core_domain.ServiceOfferFilter{}, err
	}
	nearby, err := core_domain.NewNearbyQuery(lat, lng, radius)
	if err != nil {
		return core_domain.ServiceOfferFilter{}, err
	}

	filter := core_domain.ServiceOfferFilter{Nearby: nearby}
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = &category
	}
	return filter, nil
}
