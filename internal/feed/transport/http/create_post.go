package feed_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_request "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/request"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *FeedHttpHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("CreatePost handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("CreatePost handler: no owner", err)
		return
	}

	postDto := PostRequestDto{}
	if err := core_http_request.DecodeAndValidate(r, &postDto); err != nil {
		httpResponseHandler.ErrorResponse("failed to decode or validate json", err)
		return
	}

	post, err := h.feedService.CreatePost(ctx, ownerId, DtoToDomain(postDto))
	if err != nil {
		httpResponseHandler.ErrorResponse("CreatePost handler: feed service", err)
		return
	}

	httpResponseHandler.JsonResponse(DomainToDto(post), http.StatusCreated)
}
