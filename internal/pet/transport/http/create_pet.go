package pet_transport_http

import (
	"net/http"

	"github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *PetHttpHandler) CreatePet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	petRequestDto := PetRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &petRequestDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}
	petDomain := DtoToDomain(petRequestDto)
	pet, err := h.petService.CreatePet(ctx, petDomain)
	if err != nil {
		httpResponseHandler.ErrorResponse("transport create pet", err)
		return
	}
	petResponseDto := DomainToDto(pet)

	httpResponseHandler.JsonResponse(petResponseDto, http.StatusCreated)

}
