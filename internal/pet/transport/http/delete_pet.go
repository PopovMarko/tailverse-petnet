package pet_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *PetHttpHandler) DeletePet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("DeletePet handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("DeletePet handler: no owner", err)
		return
	}

	petId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("DeletePet handler: invalid pet id", err)
		return
	}

	if err := h.petService.DeletePet(ctx, ownerId, petId); err != nil {
		httpResponseHandler.ErrorResponse("DeletePet handler: pet service", err)
		return
	}

	httpResponseHandler.NoContentResponse()
}
