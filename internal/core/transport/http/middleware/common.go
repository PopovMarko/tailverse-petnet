package core_http_middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIdKey = "X-Request-ID"

var prefix string

type Middleware func(http.Handler) http.Handler

func init() {
	p, err := os.Hostname()
	if p == "" || err != nil {
		p = "localhost"
	}
	prefix = p
}

func RequestID() Middleware {
	previousHandler := func(next http.Handler) http.Handler {
		currentHandler := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID := r.Header.Get(requestIdKey)
			if requestID == "" {
				myid := uuid.NewString()
				requestID = fmt.Sprintf("%s-%s", prefix, myid)
			}

			w.Header().Set(requestIdKey, requestID)
			r.Header.Set(requestIdKey, requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(currentHandler)
	}
	return previousHandler
}

func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, UPDATE, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Logger(logger *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIdKey)

			// Logger preconfigure for particular request with request id and url
			log := logger.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.String()),
			)

			ctx := core_logger.ToContext(r.Context(), log)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := core_logger.FromContext(ctx)
			rw := core_http_response.NewResponseWriter(w)

			// On the way IN
			before := time.Now().UTC()
			logger.Debug(">>> Incoming http request",
				zap.String("url", r.URL.String()),
				zap.Time("came_at: ", before))

			next.ServeHTTP(rw, r)

			//On the way OUT
			logger.Debug("<<< Outcoming http response",
				zap.Int("status", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)))
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHttpResponseHandler(logger, w)

			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse("panic", p)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
