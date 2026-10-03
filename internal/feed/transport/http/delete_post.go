package feed_transport_http

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

func (h *FeedHttpHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("DeletePost handler called")

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("DeletePost handler: no owner", err)
		return
	}

	postId, err := core_http_utils.GetUUIDPathParam(r, "id")
	if err != nil {
		httpResponseHandler.ErrorResponse("DeletePost handler: invalid post id", err)
		return
	}

	if err := h.feedService.DeletePost(ctx, ownerId, postId); err != nil {
		httpResponseHandler.ErrorResponse("DeletePost handler: feed service", err)
		return
	}

	httpResponseHandler.NoContentResponse()
}
