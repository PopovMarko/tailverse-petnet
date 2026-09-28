package feed_transport_http

import (
	"net/http"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *FeedHttpHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("GetPost handler called")

	postId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("GetPost handler: invalid post id", err)
		return
	}

	post, err := h.feedService.GetPost(ctx, postId)
	if err != nil {
		httpResponseHandler.ErrorResponse("GetPost handler: feed service", err)
		return
	}

	httpResponseHandler.JsonResponse(DomainToDto(post), http.StatusOK)
}
