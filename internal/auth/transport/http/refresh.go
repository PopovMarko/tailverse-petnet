package auth_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *AuthHttpHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("Refresh handler called")

	refreshDto := RefreshRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &refreshDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	tokens, err := h.authService.Refresh(ctx, refreshDto.RefreshToken)
	if err != nil {
		httpResponseHandler.ErrorResponse("Refresh handler: auth service", err)
		return
	}

	httpResponseHandler.JsonResponse(TokensToDto(tokens), http.StatusOK)
}
