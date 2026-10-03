package core_http_middleware

import (
	"fmt"
	"net/http"
	"strings"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"go.uber.org/zap"
)

type AccessTokenParser interface {
	ParseAccessToken(token string) (string, error)
}

// Auth requires "Authorization: Bearer {token}" and puts the owner id into the request context.
// The token may also come from the "token" query parameter, which WebSocket clients use during the handshake.
func Auth(parser AccessTokenParser) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHttpResponseHandler(logger, w)

			token := bearerToken(r)
			if token == "" {
				responseHandler.ErrorResponse("authentication required", fmt.Errorf("missing bearer token: %w", core_errors.ErrUnauthorized))
				return
			}

			ownerId, err := parser.ParseAccessToken(token)
			if err != nil {
				responseHandler.ErrorResponse("invalid access token", err)
				return
			}

			ctx = core_auth.ToContext(ctx, ownerId)
			ctx = core_logger.ToContext(ctx, logger.With(zap.String("owner_id", ownerId)))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if scheme, token, ok := strings.Cut(header, " "); ok && strings.EqualFold(scheme, "Bearer") {
		return strings.TrimSpace(token)
	}
	return r.URL.Query().Get("token")
}
