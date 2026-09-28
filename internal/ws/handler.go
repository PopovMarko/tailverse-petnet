package ws

import (
	"net/http"

	core_auth "github.com/PopovMarko/tailverse-petnet/internal/core/auth"
	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	core_http_response "github.com/PopovMarko/tailverse-petnet/internal/core/transport/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Mobile clients send no Origin header, and every connection is authenticated by token.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WsHttpHandler struct {
	hub *Hub
}

func NewWsHttpHandler(hub *Hub) *WsHttpHandler {
	return &WsHttpHandler{hub: hub}
}

// NewWsRouter registers GET /presence. The token may be passed as "Authorization: Bearer" or "?token=".
func NewWsRouter(h *WsHttpHandler, authMiddleware func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.With(authMiddleware).Get("/presence", h.Presence)
	return r
}

func (h *WsHttpHandler) Presence(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	httpResponseHandler := core_http_response.NewHttpResponseHandler(logger, w)

	ownerId, err := core_auth.OwnerIdFromContext(ctx)
	if err != nil {
		httpResponseHandler.ErrorResponse("Presence handler: no owner", err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade has already written the HTTP error response.
		logger.Debug("websocket upgrade failed", zap.Error(err))
		return
	}

	client := newClient(h.hub, conn, ownerId, logger)
	if !h.hub.register(client) {
		conn.Close()
		return
	}
	logger.Debug("websocket client connected", zap.Int("clients", h.hub.ClientsCount()))

	go client.writePump()
	go client.readPump()
}
