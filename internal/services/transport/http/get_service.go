package services_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *ServicesHttpHandler) GetService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetService handler called")

	serviceId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("GetService handler: invalid service id", err)
		return
	}

	offer, err := h.serviceOfferService.GetServiceOffer(ctx, serviceId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetService handler: service offer service", err)
		return
	}

	httpResponseHandler.JsonResponse(DomainToDto(offer), http.StatusOK)
}
