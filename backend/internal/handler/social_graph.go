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

func (h *SocialHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	userID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	profile, err := h.social.GetUserProfile(r.Context(), user, userID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

func (h *SocialHandler) ListUserPosts(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	userID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	posts, err := h.social.ListUserPosts(r.Context(), user, userID, limit)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load posts")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"posts": posts})
}

func (h *SocialHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	friends, err := h.social.ListFriends(r.Context(), user)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load friends")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"friends": friends})
}

func (h *SocialHandler) ListFriendRequests(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	reqs, err := h.social.ListFriendRequests(r.Context(), user)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load requests")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"requests": reqs})
}

func (h *SocialHandler) RequestFriend(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	targetID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.RequestFriend(r.Context(), user, targetID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "pending"})
}

func (h *SocialHandler) AcceptFriend(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	requesterID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.AcceptFriend(r.Context(), user, requesterID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func (h *SocialHandler) DeclineFriend(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	requesterID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.DeclineFriend(r.Context(), user, requesterID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

func (h *SocialHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	otherID, err := SocialUserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.RemoveFriend(r.Context(), user, otherID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to remove friend")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	convs, err := h.social.ListConversations(r.Context(), user)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load conversations")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"conversations": convs})
}

func (h *SocialHandler) OpenConversation(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	var body struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := DecodeJSON(r, &body); err != nil || body.UserID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	conv, err := h.social.OpenConversation(r.Context(), user, body.UserID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, conv)
}

func (h *SocialHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	conversationID, err := SocialConversationIDFromRequest(r)
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
	msgs, err := h.social.ListMessages(r.Context(), user, conversationID, limit, before)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load messages")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"messages": msgs})
}

func (h *SocialHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	conversationID, err := SocialConversationIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.SendSocialMessageInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	msg, err := h.social.SendMessage(r.Context(), user, conversationID, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, msg)
}

func (h *SocialHandler) ShareComment(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	commentID, err := SocialCommentIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.ShareCommentInput
	if err := DecodeJSON(r, &input); err != nil || input.FriendUserID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "friend_user_id is required")
		return
	}
	msg, err := h.social.ShareCommentWithFriend(r.Context(), user, commentID, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, msg)
}

func SocialConversationIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "conversationID")
	if raw == "" {
		return uuid.Nil, errors.New("conversation id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid conversation id")
	}
	return id, nil
}

func SocialCommentIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "commentID")
	if raw == "" {
		return uuid.Nil, errors.New("comment id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid comment id")
	}
	return id, nil
}
