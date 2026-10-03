package ws

import (
	"encoding/json"
	"sync"
	"time"

	core_logger "github.com/PopovMarko/tailverse-petnet/internal/core/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = pongWait * 9 / 10
	maxMessageSize = 4096
	sendBufferSize = 64
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	ownerId   string
	logger    *core_logger.Logger
	closeOnce sync.Once
	done      chan struct{}
}

func newClient(hub *Hub, conn *websocket.Conn, ownerId string, logger *core_logger.Logger) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, sendBufferSize),
		ownerId: ownerId,
		logger:  logger,
		done:    make(chan struct{}),
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// readPump handles pongs and client messages until the connection breaks.
func (c *Client) readPump() {
	defer c.hub.unregister(c)

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.logger.Debug("websocket read", zap.Error(err))
			}
			return
		}

		var message clientMessage
		if err := json.Unmarshal(data, &message); err != nil {
			c.logger.Debug("websocket: invalid client message", zap.Error(err))
			continue
		}
		switch message.Type {
		case "location_update":
			// Live location during a walk is optional in the API design; it is accepted but not stored yet.
			c.logger.Debug("websocket location_update", zap.Float64("lat", message.Lat), zap.Float64("lng", message.Lng))
		default:
			c.logger.Debug("websocket: unknown client message type", zap.String("type", message.Type))
		}
	}
}

// writePump sends queued events and keepalive pings. It is the only goroutine that writes to conn.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.hub.unregister(c)
	}()

	for {
		select {
		case message := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
