package auth_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *AuthHttpHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("Register handler called")

	registerDto := RegisterRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &registerDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	owner, tokens, err := h.authService.Register(ctx, registerDto.Email, registerDto.Password, RegisterDtoToProfile(registerDto))
	if err != nil {
		httpResponseHandler.ErrorResponse("Register handler: auth service", err)
		return
	}

	httpResponseHandler.JsonResponse(RegisterDomainToDto(owner, tokens), http.StatusCreated)
}
