package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	ws "github.com/k8s-cmp/k8s-cmp/internal/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 운영환경에서는 origin 검증 필요
		return true
	},
}

// WebSocketHandler는 WebSocket 연결을 처리합니다.
type WebSocketHandler struct {
	hub     *ws.Hub
	authSvc *service.AuthService
	logger  *zap.Logger
}

// NewWebSocketHandler는 새로운 WebSocketHandler를 생성합니다.
func NewWebSocketHandler(hub *ws.Hub, authSvc *service.AuthService, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, authSvc: authSvc, logger: logger}
}

// StreamCluster godoc
// GET /ws/v1/clusters/:id/stream
func (h *WebSocketHandler) StreamCluster(c *gin.Context) {
	// JWT 토큰 검증 (쿼리 파라미터)
	token := c.Query("token")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	if _, err := h.authSvc.ValidateAccessToken(token); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	clusterID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket", zap.Error(err))
		return
	}

	client := ws.NewClient(h.hub, conn, clusterID, h.logger)
	h.hub.Register <- &ws.Subscription{ClusterID: clusterID, Client: client}

	go client.WritePump()
	go client.ReadPump()
}

// StreamAll godoc
// GET /ws/v1/clusters/stream (모든 클러스터 스트림)
func (h *WebSocketHandler) StreamAll(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	if _, err := h.authSvc.ValidateAccessToken(token); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket", zap.Error(err))
		return
	}

	// 빈 ClusterID = 전체 클러스터 스트림
	client := ws.NewClient(h.hub, conn, "", h.logger)
	h.hub.Register <- &ws.Subscription{ClusterID: "", Client: client}

	go client.WritePump()
	go client.ReadPump()
}
