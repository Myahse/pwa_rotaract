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

type RegistrationHandler struct {
	registration *service.RegistrationService
	clubs        *repository.ClubRepository
	requests     *repository.AccessRequestRepository
	publicURL    string
}

func NewRegistrationHandler(
	registration *service.RegistrationService,
	clubs *repository.ClubRepository,
	requests *repository.AccessRequestRepository,
	publicURL string,
) *RegistrationHandler {
	return &RegistrationHandler{
		registration: registration,
		clubs:        clubs,
		requests:     requests,
		publicURL:    publicURL,
	}
}

func (h *RegistrationHandler) PreviewInvite(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	preview, err := h.registration.PreviewInvite(r.Context(), code, h.publicURL)
	if errors.Is(err, service.ErrInvalidInviteCode) {
		WriteError(w, http.StatusNotFound, "invalid invite code")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load invite")
		return
	}
	WriteJSON(w, http.StatusOK, preview)
}

func (h *RegistrationHandler) PreviewEmailInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	preview, err := h.registration.PreviewEmailInvite(r.Context(), token)
	if errors.Is(err, service.ErrInvalidInviteToken) {
		WriteError(w, http.StatusNotFound, "invalid or expired invite link")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load invite")
		return
	}
	WriteJSON(w, http.StatusOK, preview)
}

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.registration.Register(r.Context(), input)
	if errors.Is(err, service.ErrInvalidInviteToken) {
		WriteError(w, http.StatusNotFound, "invalid or expired invite link")
		return
	}
	if errors.Is(err, service.ErrInvalidInviteCode) {
		WriteError(w, http.StatusNotFound, "invalid invite code")
		return
	}
	if errors.Is(err, service.ErrInviteEmailMismatch) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrEmailAlreadyUsed) {
		WriteError(w, http.StatusConflict, "email is already registered")
		return
	}
	if errors.Is(err, service.ErrRoleNotJoinable) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

func (h *RegistrationHandler) SendEmailInvite(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.SendEmailInviteInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	invite, err := h.registration.SendEmailInvite(r.Context(), clubID, user.ID, input)
	if errors.Is(err, service.ErrEmailAlreadyUsed) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrPendingEmailInvite) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusAccepted, map[string]any{
		"id":         invite.ID,
		"email":      invite.Email,
		"expires_at": invite.ExpiresAt,
		"message":    "registration invite sent by email",
	})
}

func (h *RegistrationHandler) ListEmailInvites(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	invites, err := h.registration.ListEmailInvites(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list email invites")
		return
	}
	WriteJSON(w, http.StatusOK, invites)
}

func (h *RegistrationHandler) RevokeEmailInvite(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	inviteID, err := InviteIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.registration.RevokeEmailInvite(r.Context(), clubID, inviteID); errors.Is(err, service.ErrInviteAlreadyUsed) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "invite not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (h *RegistrationHandler) SubmitAccessRequest(w http.ResponseWriter, r *http.Request) {
	var input service.AccessRequestInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req, err := h.registration.SubmitAccessRequest(r.Context(), input)
	if errors.Is(err, service.ErrPendingRequest) || errors.Is(err, service.ErrAlreadyClubMember) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrPasswordRequired) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusAccepted, req)
}

func (h *RegistrationHandler) GetClubInvite(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	club, err := h.clubs.GetByID(r.Context(), clubID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get club")
		return
	}

	WriteJSON(w, http.StatusOK, h.registration.GetClubInviteLink(club, h.publicURL))
}

func (h *RegistrationHandler) ListAccessRequests(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := domain.AccessRequestStatusPending
	requests, err := h.requests.List(r.Context(), &clubID, &status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list access requests")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

func (h *RegistrationHandler) ApproveAccessRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	requestID, err := requestIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	req, err := h.requests.GetByID(r.Context(), requestID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "access request not found")
		return
	}
	if req.ClubID != nil && *req.ClubID != clubID {
		WriteError(w, http.StatusForbidden, "request does not belong to this club")
		return
	}

	var input service.ApproveAccessRequestInput
	if r.ContentLength > 0 {
		if err := DecodeJSON(r, &input); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}
	input.ClubID = &clubID

	member, err := h.registration.ApproveAccessRequest(r.Context(), user.ID, requestID, input)
	if errors.Is(err, service.ErrRequestNotPending) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrClubNotMatched) {
		WriteError(w, http.StatusBadRequest, "club_id is required to approve this request")
		return
	}
	if errors.Is(err, service.ErrEmailAlreadyUsed) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, member)
}

func (h *RegistrationHandler) RejectAccessRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	requestID, err := requestIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Note *string `json:"note"`
	}
	_ = DecodeJSON(r, &body)

	if err := h.registration.RejectAccessRequest(r.Context(), user.ID, requestID, body.Note); err != nil {
		if errors.Is(err, service.ErrRequestNotPending) {
			WriteError(w, http.StatusConflict, err.Error())
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (h *RegistrationHandler) ListAdminAccessRequests(w http.ResponseWriter, r *http.Request) {
	status := domain.AccessRequestStatusPending
	requests, err := h.requests.List(r.Context(), nil, &status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list access requests")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

func (h *RegistrationHandler) ApproveAdminAccessRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	requestID, err := requestIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.ApproveAccessRequestInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member, err := h.registration.ApproveAccessRequest(r.Context(), user.ID, requestID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, member)
}

func requestIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "requestID")
	return uuid.Parse(raw)
}
