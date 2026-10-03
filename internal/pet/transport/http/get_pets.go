package pet_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"go.uber.org/zap"
)

func (h *PetHttpHandler) GetPets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetPets caled")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetPets handler: no owner", err)
		return
	}

	petsDomainList, err := h.petService.GetPets(ctx, ownerId)
	if err != nil {
		httpResponseHandler.ErrorResponse("pet service response", err)
		return
	}

	petsDtoList := make([]PetResponseDto, len(petsDomainList))
	for i, pet := range petsDomainList {
		petsDtoList[i] = DomainToDto(pet)
	}

	logger.Debug("GetPets", zap.Any("pets_dto_list", petsDtoList))

	httpResponseHandler.JsonResponse(PetsListResponseDto{Pets: petsDtoList}, http.StatusOK)
}
