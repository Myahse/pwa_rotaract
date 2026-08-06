package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/service"
)

type SocialHandler struct {
	social *service.SocialService
}

func NewSocialHandler(social *service.SocialService) *SocialHandler {
	return &SocialHandler{social: social}
}

func (h *SocialHandler) optionalUser(r *http.Request) *domain.User {
	user, _ := authctx.UserFromContext(r.Context())
	return user
}

func (h *SocialHandler) writeSocialErr(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, service.ErrSocialLoginRequired):
		WriteError(w, http.StatusUnauthorized, "connexion requise")
	case errors.Is(err, service.ErrSocialPostNotFound),
		errors.Is(err, service.ErrSocialCommentNotFound),
		errors.Is(err, service.ErrSocialUserNotFound),
		errors.Is(err, service.ErrSocialConversationNF),
		errors.Is(err, service.ErrSocialGroupNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrSocialForbidden),
		errors.Is(err, service.ErrSocialNotFriends),
		errors.Is(err, service.ErrSocialGroupForbidden),
		errors.Is(err, service.ErrSocialGroupNotMember):
		WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrInvalidSocialMedia), errors.Is(err, service.ErrSocialMediaTooLarge):
		WriteError(w, http.StatusBadRequest, err.Error())
	default:
		return false
	}
	return true
}

func (h *SocialHandler) ListFeed(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	followingOnly := r.URL.Query().Get("following") == "1" || r.URL.Query().Get("following") == "true"
	var before *time.Time
	if raw := r.URL.Query().Get("before"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid before timestamp")
			return
		}
		before = &t
	}
	posts, err := h.social.ListFeed(r.Context(), user, limit, before, followingOnly)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load feed")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"posts": posts})
}

func (h *SocialHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var input service.CreateSocialPostInput

	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(55 << 20); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}
		input.Body = r.FormValue("body")
		if raw := r.FormValue("club_id"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "invalid club_id")
				return
			}
			input.ClubID = &id
		}
		if r.MultipartForm != nil {
			for _, header := range r.MultipartForm.File["media"] {
				file, err := header.Open()
				if err != nil {
					WriteError(w, http.StatusBadRequest, "invalid media file")
					return
				}
				input.Media = append(input.Media, service.SocialMediaUpload{
					Filename:    header.Filename,
					ContentType: header.Header.Get("Content-Type"),
					Size:        header.Size,
					Reader:      file,
				})
			}
		}
		defer func() {
			for _, m := range input.Media {
				if c, ok := m.Reader.(io.Closer); ok {
					_ = c.Close()
				}
			}
		}()
	} else {
		var body struct {
			Body   string     `json:"body"`
			ClubID *uuid.UUID `json:"club_id"`
		}
		if err := DecodeJSON(r, &body); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		input.Body = body.Body
		input.ClubID = body.ClubID
	}

	post, err := h.social.CreatePost(r.Context(), user, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, post)
}

func (h *SocialHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	postID, err := SocialPostIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	post, err := h.social.GetPost(r.Context(), user, postID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load post")
		return
	}
	WriteJSON(w, http.StatusOK, post)
}

func (h *SocialHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	postID, err := SocialPostIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.DeletePost(r.Context(), user, postID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	postID, err := SocialPostIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	comments, err := h.social.ListComments(r.Context(), user, postID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load comments")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"comments": comments})
}

func (h *SocialHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	postID, err := SocialPostIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.CreateSocialCommentInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	comment, err := h.social.CreateComment(r.Context(), user, postID, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, comment)
}

func (h *SocialHandler) React(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	postID, err := SocialPostIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.ReactInput
	if err := DecodeJSON(r, &input); err != nil {
		input.Kind = "like"
	}
	err = h.social.React(r.Context(), user, postID, input)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "reacted"})
}

func (h *SocialHandler) Unreact(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	postID, err := SocialPostIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.social.Unreact(r.Context(), user, postID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to remove reaction")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	user := h.optionalUser(r)
	profiles, err := h.social.Suggestions(r.Context(), user)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load suggestions")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"users": profiles})
}

func (h *SocialHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "connexion requise")
		return
	}
	profiles, err := h.social.ListFollowing(r.Context(), user)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load following")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"users": profiles})
}

func (h *SocialHandler) Follow(w http.ResponseWriter, r *http.Request) {
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
	err = h.social.Follow(r.Context(), user, targetID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "following"})
}

func (h *SocialHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
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
	err = h.social.Unfollow(r.Context(), user, targetID)
	if h.writeSocialErr(w, err) {
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to unfollow")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func SocialPostIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "postID")
	if raw == "" {
		return uuid.Nil, errors.New("post id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid post id")
	}
	return id, nil
}

func SocialUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "userID")
	if raw == "" {
		return uuid.Nil, errors.New("user id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}
	return id, nil
}
