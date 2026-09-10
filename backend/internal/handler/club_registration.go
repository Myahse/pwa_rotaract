package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type ClubRegistrationHandler struct {
	registrations *service.ClubRegistrationService
}

func NewClubRegistrationHandler(registrations *service.ClubRegistrationService) *ClubRegistrationHandler {
	return &ClubRegistrationHandler{registrations: registrations}
}

func (h *ClubRegistrationHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var input service.SubmitClubRegistrationInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req, err := h.registrations.Submit(r.Context(), input)
	if errors.Is(err, service.ErrEmailAlreadyUsed) {
		WriteError(w, http.StatusConflict, "email is already registered")
		return
	}
	if errors.Is(err, service.ErrClubNameTaken) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrPendingClubRegistration) || errors.Is(err, service.ErrPendingClubName) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusAccepted, req)
}

func (h *ClubRegistrationHandler) List(w http.ResponseWriter, r *http.Request) {
	var statusFilter *domain.ClubRegistrationStatus
	if raw := r.URL.Query().Get("status"); raw != "" {
		status := domain.ClubRegistrationStatus(raw)
		switch status {
		case domain.ClubRegistrationStatusPending,
			domain.ClubRegistrationStatusApproved,
			domain.ClubRegistrationStatusRejected:
			statusFilter = &status
		default:
			WriteError(w, http.StatusBadRequest, "invalid status filter")
			return
		}
	} else {
		pending := domain.ClubRegistrationStatusPending
		statusFilter = &pending
	}

	requests, err := h.registrations.List(r.Context(), statusFilter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list club registration requests")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

func (h *ClubRegistrationHandler) Get(w http.ResponseWriter, r *http.Request) {
	requestID, err := clubRegistrationIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	req, err := h.registrations.GetByID(r.Context(), requestID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club registration request not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load club registration request")
		return
	}
	WriteJSON(w, http.StatusOK, req)
}

func (h *ClubRegistrationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	requestID, err := clubRegistrationIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.ApproveClubRegistrationInput
	if r.ContentLength > 0 {
		if err := DecodeJSON(r, &input); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	result, err := h.registrations.Approve(r.Context(), user.ID, requestID, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club registration request not found")
		return
	}
	if errors.Is(err, service.ErrClubRegistrationNotPending) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrEmailAlreadyUsed) || errors.Is(err, service.ErrClubNameTaken) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *ClubRegistrationHandler) Reject(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	requestID, err := clubRegistrationIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Note *string `json:"note"`
	}
	_ = DecodeJSON(r, &body)

	if err := h.registrations.Reject(r.Context(), user.ID, requestID, body.Note); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "club registration request not found")
			return
		}
		if errors.Is(err, service.ErrClubRegistrationNotPending) {
			WriteError(w, http.StatusConflict, err.Error())
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (h *ClubRegistrationHandler) PreviewAccess(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	preview, err := h.registrations.PreviewAccess(r.Context(), token)
	if errors.Is(err, service.ErrInvalidClubAccessToken) {
		WriteError(w, http.StatusNotFound, "invalid or expired registration link")
		return
	}
	if errors.Is(err, service.ErrClubAccessAlreadyCompleted) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load registration link")
		return
	}
	WriteJSON(w, http.StatusOK, preview)
}

func (h *ClubRegistrationHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var input service.CompleteClubRegistrationInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.registrations.Complete(r.Context(), input)
	if errors.Is(err, service.ErrGoogleAuthNotConfigured) {
		WriteError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if errors.Is(err, service.ErrPasswordRequired) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrInvalidClubAccessToken) {
		WriteError(w, http.StatusNotFound, "invalid or expired registration link")
		return
	}
	if errors.Is(err, service.ErrClubAccessAlreadyCompleted) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrGoogleEmailMismatch) {
		WriteError(w, http.StatusForbidden, err.Error())
		return
	}
	if errors.Is(err, service.ErrEmailAlreadyUsed) || errors.Is(err, service.ErrClubNameTaken) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, result)
}

func (h *ClubRegistrationHandler) GoogleClientID(w http.ResponseWriter, r *http.Request) {
	clientID := h.registrations.GoogleClientID()
	if clientID == "" {
		WriteError(w, http.StatusServiceUnavailable, "google sign-in is not configured")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"client_id": clientID})
}

func clubRegistrationIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "requestID")
	if raw == "" {
		return uuid.Nil, errors.New("request id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid request id")
	}
	return id, nil
}
