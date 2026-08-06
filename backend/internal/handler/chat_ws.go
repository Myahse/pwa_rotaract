package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/service"
	"github.com/rotaract-civ/backend/internal/ws"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatWSHandler struct {
	chat  *service.ChatService
	hub   *ws.Hub
	token *auth.TokenManager
}

func NewChatWSHandler(chat *service.ChatService, hub *ws.Hub, token *auth.TokenManager) *ChatWSHandler {
	return &ChatWSHandler{chat: chat, hub: hub, token: token}
}

func (h *ChatWSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	if token == "" {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	claims, err := h.token.Parse(token)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	groupID, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("group_id")))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid group_id")
		return
	}

	if err := h.chat.EnsureGroupAccess(r.Context(), groupID, claims.UserID, claims.IsAdmin); err != nil {
		WriteError(w, http.StatusForbidden, "access denied to this group")
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Hub:     h.hub,
		Conn:    conn,
		GroupID: groupID,
		UserID:  claims.UserID,
		Send:    make(chan []byte, 256),
	}
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
