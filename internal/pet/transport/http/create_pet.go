package pet_transport_http

import (
	"net/http"

	"github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	"github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"go.uber.org/zap"
)

func (h *PetHttpHandler) CreatePet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CreatePet handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("CreatePet handler: no owner", err)
		return
	}

	petRequestDto := PetRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &petRequestDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}
	petDomain := DtoToDomain(petRequestDto)
	pet, err := h.petService.CreatePet(ctx, ownerId, petDomain)
	if err != nil {
		httpResponseHandler.ErrorResponse("transport create pet", err)
		return
	}

	petResponseDto := DomainToDto(pet)

	logger.Debug("resposne from pet service", zap.Any("pet_response_dto", petResponseDto))

	httpResponseHandler.JsonResponse(petResponseDto, http.StatusCreated)
}
