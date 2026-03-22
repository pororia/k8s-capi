package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client는 개별 WebSocket 연결을 나타냅니다.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	clusterID string
	Send      chan []byte
	logger    *zap.Logger
}

// NewClient는 새로운 WebSocket 클라이언트를 생성합니다.
func NewClient(hub *Hub, conn *websocket.Conn, clusterID string, logger *zap.Logger) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		clusterID: clusterID,
		Send:      make(chan []byte, 256),
		logger:    logger,
	}
}

// ReadPump은 클라이언트로부터 메시지를 읽는 루프입니다.
// WebSocket 연결 유지를 위한 ping/pong 처리를 담당합니다.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- &Subscription{ClusterID: c.clusterID, Client: c}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Debug("websocket unexpected close", zap.Error(err))
			}
			break
		}
	}
}

// WritePump은 클라이언트로 메시지를 전송하는 루프입니다.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Debug("websocket write error", zap.Error(err))
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
