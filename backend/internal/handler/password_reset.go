package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rotaract-civ/backend/internal/service"
)

type PasswordResetHandler struct {
	resets *service.PasswordResetService
	logger *slog.Logger
}

func NewPasswordResetHandler(resets *service.PasswordResetService, logger *slog.Logger) *PasswordResetHandler {
	return &PasswordResetHandler{resets: resets, logger: logger}
}

func (h *PasswordResetHandler) Request(w http.ResponseWriter, r *http.Request) {
	var input service.RequestPasswordResetInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.resets.Request(r.Context(), input); err != nil {
		h.logger.Error("password reset request failed", "err", err)
		WriteError(w, http.StatusInternalServerError, "failed to request password reset")
		return
	}
	WriteJSON(w, http.StatusAccepted, map[string]string{"message": "Si cette adresse existe, un lien de réinitialisation sera envoyé."})
}

func (h *PasswordResetHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var input service.CompletePasswordResetInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.resets.Complete(r.Context(), input); errors.Is(err, service.ErrInvalidPasswordResetToken) {
		WriteError(w, http.StatusBadRequest, "invalid or expired reset link")
		return
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Mot de passe mis à jour."})
}
