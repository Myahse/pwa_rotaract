package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rotaract-civ/backend/internal/domain"
)

type Hub struct {
	logger  *slog.Logger
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{}
}

type Client struct {
	Hub     *Hub
	UserID  uuid.UUID
	GroupID uuid.UUID
	Conn    *websocket.Conn
	Send    chan []byte
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[uuid.UUID]map[*Client]struct{}),
	}
}

type Event struct {
	Type      string              `json:"type"`
	GroupID   uuid.UUID           `json:"group_id"`
	Message   *domain.ChatMessage `json:"message,omitempty"`
	MessageID *uuid.UUID          `json:"message_id,omitempty"`
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.GroupID] == nil {
		h.clients[client.GroupID] = make(map[*Client]struct{})
	}
	h.clients[client.GroupID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if members, ok := h.clients[client.GroupID]; ok {
		delete(members, client)
		if len(members) == 0 {
			delete(h.clients, client.GroupID)
		}
	}
	close(client.Send)
}

func (h *Hub) BroadcastMessage(groupID uuid.UUID, message domain.ChatMessage) {
	msg := message
	h.BroadcastEvent(groupID, Event{
		Type:    "message",
		GroupID: groupID,
		Message: &msg,
	})
}

func (h *Hub) BroadcastEvent(groupID uuid.UUID, event Event) {
	event.GroupID = groupID
	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("marshal ws event", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients[groupID] {
		select {
		case client.Send <- payload:
		default:
			go func(c *Client) {
				h.Unregister(c)
				_ = c.Conn.Close()
			}(client)
		}
	}
}

const (
	WriteWait  = 10
	PongWait   = 60
	PingPeriod = 54
	MaxMsgSize = 4096
)
