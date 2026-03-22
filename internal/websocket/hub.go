package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Event는 WebSocket으로 전송되는 이벤트입니다.
type Event struct {
	Type      string      `json:"type"`
	ClusterID string      `json:"cluster_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Subscription은 클라이언트의 클러스터 구독 정보입니다.
type Subscription struct {
	ClusterID string
	Client    *Client
}

// Hub는 WebSocket 연결을 관리합니다.
type Hub struct {
	mu       sync.RWMutex
	// 클러스터ID → 클라이언트 set
	clusters map[string]map[*Client]bool

	Register   chan *Subscription
	Unregister chan *Subscription
	Broadcast  chan *Event

	logger *zap.Logger
}

// NewHub는 새로운 Hub를 생성합니다.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clusters:   make(map[string]map[*Client]bool),
		Register:   make(chan *Subscription, 64),
		Unregister: make(chan *Subscription, 64),
		Broadcast:  make(chan *Event, 256),
		logger:     logger,
	}
}

// Run은 Hub의 메인 루프를 실행합니다.
func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.Register:
			h.mu.Lock()
			if _, ok := h.clusters[sub.ClusterID]; !ok {
				h.clusters[sub.ClusterID] = make(map[*Client]bool)
			}
			h.clusters[sub.ClusterID][sub.Client] = true
			h.mu.Unlock()

		case sub := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clusters[sub.ClusterID]; ok {
				delete(clients, sub.Client)
				if len(clients) == 0 {
					delete(h.clusters, sub.ClusterID)
				}
			}
			h.mu.Unlock()

		case event := <-h.Broadcast:
			h.broadcastToCluster(event)
		}
	}
}

func (h *Hub) broadcastToCluster(event *Event) {
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event", zap.Error(err))
		return
	}

	h.mu.RLock()
	clients := h.clusters[event.ClusterID]
	// 전체 스트림 구독자도 포함 (빈 ClusterID = 전체)
	allClients := h.clusters[""]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			h.logger.Warn("client send buffer full, disconnecting", zap.String("cluster_id", event.ClusterID))
			h.Unregister <- &Subscription{ClusterID: event.ClusterID, Client: client}
		}
	}

	for client := range allClients {
		select {
		case client.Send <- data:
		default:
			h.Unregister <- &Subscription{ClusterID: "", Client: client}
		}
	}
}

// BroadcastEvent는 특정 클러스터에 이벤트를 브로드캐스트합니다.
func (h *Hub) BroadcastEvent(eventType, clusterID string, data interface{}) {
	h.Broadcast <- &Event{
		Type:      eventType,
		ClusterID: clusterID,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}
