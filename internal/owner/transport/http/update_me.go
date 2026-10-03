package owner_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// UpdateMe handles PATCH /owners/me: a partial update of the authenticated owner's profile.
func (h *OwnerHttpHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("UpdateMe handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("UpdateMe handler: no owner", err)
		return
	}

	ownerUpdateDto := OwnerUpdateDto{}
	if err := core_http_request.DecodeAndValidate(r, &ownerUpdateDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	owner, err := h.ownerService.UpdateOwner(ctx, ownerId, UpdateDtoToDomain(ownerUpdateDto))
	if err != nil {
		httpResponseHandler.ErrorResponse("UpdateMe handler: owner service", err)
		return
	}

	httpResponseHandler.JsonResponse(OwnerToDto(owner), http.StatusOK)
}
