package services_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *ServicesHttpHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CreateService handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("CreateService handler: no owner", err)
		return
	}

	serviceDto := ServiceRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &serviceDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	offer, err := h.serviceOfferService.CreateServiceOffer(ctx, ownerId, DtoToDomain(serviceDto))
	if err != nil {
		httpResponseHandler.ErrorResponse("CreateService handler: service offer service", err)
		return
	}

	httpResponseHandler.JsonResponse(DomainToDto(offer), http.StatusCreated)
}
