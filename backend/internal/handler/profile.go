package handler

import (
	"errors"
	"net/http"

	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type ProfileHandler struct {
	profiles *service.ProfileService
	clubs    *repository.ClubRepository
}

func NewProfileHandler(profiles *service.ProfileService, clubs *repository.ClubRepository) *ProfileHandler {
	return &ProfileHandler{profiles: profiles, clubs: clubs}
}

func (h *ProfileHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profile, err := h.profiles.GetProfile(r.Context(), user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	WriteJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) ListMyClubs(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	memberships, err := h.clubs.ListMembershipsByUser(r.Context(), user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load clubs")
		return
	}

	WriteJSON(w, http.StatusOK, memberships)
}

func (h *ProfileHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input service.UpdateProfileInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	profile, err := h.profiles.UpdateProfile(r.Context(), user.ID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := r.ParseMultipartForm(6 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	profile, err := h.profiles.UploadAvatar(
		r.Context(),
		user.ID,
		header.Filename,
		header.Size,
		header.Header.Get("Content-Type"),
		file,
	)
	if errors.Is(err, service.ErrInvalidAvatarType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrAvatarTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

	WriteJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profile, err := h.profiles.DeleteAvatar(r.Context(), user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	WriteJSON(w, http.StatusOK, profile)
}
