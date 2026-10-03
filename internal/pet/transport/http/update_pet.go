package pet_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *PetHttpHandler) UpdatePet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("UpdatePet handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("UpdatePet handler: no owner", err)
		return
	}

	petId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("UpdatePet handler: invalid pet id", err)
		return
	}

	petUpdateDto := PetUpdateDto{}
	if err := core_http_request.DecodeAndValidate(r, &petUpdateDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	pet, err := h.petService.UpdatePet(ctx, ownerId, petId, UpdateDtoToDomain(petUpdateDto))
	if err != nil {
		httpResponseHandler.ErrorResponse("UpdatePet handler: pet service", err)
		return
	}

	httpResponseHandler.JsonResponse(DomainToDto(pet), http.StatusOK)
}
