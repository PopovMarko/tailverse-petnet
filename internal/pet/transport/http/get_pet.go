package pet_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"go.uber.org/zap"
)

func (h *PetHttpHandler) GetPet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetPet handler called")

	petId, err := core_http_utils.GetStringPathParams(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("GetPet handler:", err)
		return
	}

	petDomain, err := h.petService.GetPet(ctx, petId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetPet handler:", err)
		return
	}

	petResponseDto := DomainToDto(petDomain)
	logger.Debug("GetPet returns:", zap.Any("pet_resposne_dto", petResponseDto))
	httpResponseHandler.JsonResponse(petResponseDto, http.StatusOK)
}
