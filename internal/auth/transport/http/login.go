package auth_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *AuthHttpHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("Login handler called")

	loginDto := LoginRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &loginDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	tokens, err := h.authService.Login(ctx, loginDto.Email, loginDto.Password)
	if err != nil {
		httpResponseHandler.ErrorResponse("Login handler: auth service", err)
		return
	}

	httpResponseHandler.JsonResponse(TokensToDto(tokens), http.StatusOK)
}
