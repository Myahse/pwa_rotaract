package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/service"
)

func (h *SocialHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	mine := r.URL.Query().Get("mine") == "1" || r.URL.Query().Get("mine") == "true"
	groups, err := h.social.ListGroups(r.Context(), user, mine)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load groups")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"groups": groups})
}

func (h *SocialHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	var input service.CreateSocialGroupInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	group, err := h.social.CreateGroup(r.Context(), user, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, group)
}

func (h *SocialHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	group, err := h.social.GetGroup(r.Context(), user, groupID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load group")
		return
	}
	WriteJSON(w, http.StatusOK, group)
}

func (h *SocialHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.JoinSocialGroupInput
	_ = DecodeJSON(r, &input)
	group, err := h.social.JoinGroup(r.Context(), user, groupID, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, group)
}

func (h *SocialHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.LeaveGroup(r.Context(), user, groupID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	members, err := h.social.ListGroupMembers(r.Context(), user, groupID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load members")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (h *SocialHandler) ListJoinRequests(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	reqs, err := h.social.ListJoinRequests(r.Context(), user, groupID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load requests")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"requests": reqs})
}

func (h *SocialHandler) ApproveJoinRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	requesterID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.ApproveJoinRequest(r.Context(), user, groupID, requesterID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (h *SocialHandler) RejectJoinRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	requesterID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.RejectJoinRequest(r.Context(), user, groupID, requesterID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (h *SocialHandler) ListGroupMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	var before *time.Time
	if raw := r.URL.Query().Get("before"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid before timestamp")
			return
		}
		before = &t
	}
	msgs, err := h.social.ListGroupMessages(r.Context(), user, groupID, limit, before)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load messages")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"messages": msgs})
}

func (h *SocialHandler) SendGroupMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	groupID, err := SocialGroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.SendGroupMessageInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	msg, err := h.social.SendGroupMessage(r.Context(), user, groupID, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, msg)
}

func SocialGroupIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "groupID")
	if raw == "" {
		return uuid.Nil, errors.New("group id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid group id")
	}
	return id, nil
}
