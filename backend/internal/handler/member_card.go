package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type MemberCardHandler struct {
	cards *service.MemberCardService
}

func NewMemberCardHandler(cards *service.MemberCardService) *MemberCardHandler {
	return &MemberCardHandler{cards: cards}
}

func (h *MemberCardHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := service.MemberCardListFilter{Limit: 200}
	if raw := r.URL.Query().Get("club_id"); raw != "" {
		clubID, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid club_id")
			return
		}
		filter.ClubID = &clubID
	}
	if r.URL.Query().Get("pending_only") == "true" {
		filter.PendingOnly = true
	}

	items, err := h.cards.List(r.Context(), filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list member cards")
		return
	}
	pending, err := h.cards.CountPending(r.Context(), filter.ClubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to count pending cards")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"items":          items,
		"pending_count":  pending,
	})
}

func (h *MemberCardHandler) IssueMissing(w http.ResponseWriter, r *http.Request) {
	var clubID *uuid.UUID
	if raw := r.URL.Query().Get("club_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid club_id")
			return
		}
		clubID = &id
	}
	created, err := h.cards.IssueMissing(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"created": created})
}

func (h *MemberCardHandler) SendOne(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	cardID, err := uuid.Parse(chi.URLParam(r, "cardID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid card id")
		return
	}
	card, err := h.cards.SendOne(r.Context(), cardID, user.ID)
	if err != nil {
		if err == repository.ErrNotFound {
			WriteError(w, http.StatusNotFound, "member card not found")
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, card)
}

func (h *MemberCardHandler) SendPending(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var clubID *uuid.UUID
	if raw := r.URL.Query().Get("club_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid club_id")
			return
		}
		clubID = &id
	}
	result, err := h.cards.SendPending(r.Context(), clubID, user.ID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, result)
}
