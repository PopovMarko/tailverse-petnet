package owner_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// GetMe handles GET /owners/me: the authenticated owner's full profile.
func (h *OwnerHttpHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetMe handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetMe handler: no owner", err)
		return
	}

	owner, err := h.ownerService.GetOwner(ctx, ownerId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetMe handler: owner service", err)
		return
	}

	httpResponseHandler.JsonResponse(OwnerToDto(owner), http.StatusOK)
}
