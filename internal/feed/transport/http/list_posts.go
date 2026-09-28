package feed_transport_http

import (
	"net/http"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_utils "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
)

// ListPosts handles GET /posts?spot_id=&cursor=&limit=
func (h *FeedHttpHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	logger.Debug("ListPosts handler called")

	filter, err := filterFromRequest(r)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListPosts handler: invalid query", err)
		return
	}

	page, err := h.feedService.ListPosts(ctx, filter)
	if err != nil {
		httpResponseHandler.ErrorResponse("ListPosts handler: feed service", err)
		return
	}

	httpResponseHandler.JsonResponse(PageToDto(page), http.StatusOK)
}

func filterFromRequest(r *http.Request) (core_domain.PostFilter, error) {
	spotId, err := core_http_utils.GetUUIDQueryParam(r, "spot_id")
	if err != nil {
		return core_domain.PostFilter{}, err
	}
	limit, err := core_http_utils.GetIntQueryParam(r, "limit")
	if err != nil {
		return core_domain.PostFilter{}, err
	}

	filter := core_domain.PostFilter{SpotId: spotId}
	if limit != nil {
		filter.Limit = *limit
		if *limit == 0 {
			filter.Limit = -1 // an explicit limit=0 is invalid, not "use the default"
		}
	}
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return core_domain.PostFilter{}, err
		}
		filter.After = &cursor
	}
	return filter, nil
}
