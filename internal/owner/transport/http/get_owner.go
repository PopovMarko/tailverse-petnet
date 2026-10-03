package owner_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// GetOwner handles GET /owners/{id}: the public view of an owner's profile.
func (h *OwnerHttpHandler) GetOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetOwner handler called")

	ownerId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("GetOwner handler: invalid owner id", err)
		return
	}

	owner, err := h.ownerService.GetPublicOwner(ctx, ownerId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetOwner handler: owner service", err)
		return
	}

	httpResponseHandler.JsonResponse(PublicOwnerToDto(owner), http.StatusOK)
}
