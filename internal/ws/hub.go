package ws

import (
	"sync"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"go.uber.org/zap"
)

// Hub keeps the connected WebSocket clients and broadcasts events to all of them.
// It implements the EventPublisher interface of the presence and announcement services.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	closed  bool
	logger  *core_logger.Logger
}

func NewHub(logger *core_logger.Logger) *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		logger:  logger,
	}
}

// Publish never blocks: a client whose send buffer is full is disconnected.
func (h *Hub) Publish(event any) {
	message, err := encodeEvent(event)
	if err != nil {
		h.logger.Error("encode websocket event", zap.Error(err))
		return
	}

	h.mu.RLock()
	var slow []*Client
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			slow = append(slow, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range slow {
		h.logger.Warn("websocket client is too slow, disconnecting", zap.String("owner_id", client.ownerId))
		h.unregister(client)
	}
}

func (h *Hub) ClientsCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Close disconnects every client; called on server shutdown because hijacked connections are not closed by http.Server.
func (h *Hub) Close() {
	h.mu.Lock()
	h.closed = true
	clients := h.clients
	h.clients = make(map[*Client]struct{})
	h.mu.Unlock()

	for client := range clients {
		client.close()
	}
}

func (h *Hub) register(client *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return false
	}
	h.clients[client] = struct{}{}
	return true
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	_, ok := h.clients[client]
	delete(h.clients, client)
	h.mu.Unlock()

	if ok {
		client.close()
	}
}
