package handler

import (
	"net/http"

	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/service"
)

type BirthdayHandler struct {
	birthdays *service.BirthdayService
}

func NewBirthdayHandler(birthdays *service.BirthdayService) *BirthdayHandler {
	return &BirthdayHandler{birthdays: birthdays}
}

func (h *BirthdayHandler) VAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{
		"public_key": h.birthdays.VAPIDPublicKey(),
	})
}

func (h *BirthdayHandler) SavePushSubscription(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input service.PushSubscriptionInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.birthdays.SavePushSubscription(r.Context(), user.ID, input, r.UserAgent()); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "subscribed"})
}

func (h *BirthdayHandler) DeletePushSubscription(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var input struct {
		Endpoint string `json:"endpoint"`
	}
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.birthdays.RemovePushSubscription(r.Context(), user.ID, input.Endpoint); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "unsubscribed"})
}

func (h *BirthdayHandler) Widget(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	widget, err := h.birthdays.WidgetForUser(r.Context(), user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load birthday widget")
		return
	}

	w.Header().Set("Cache-Control", "no-cache, max-age=0")
	WriteJSON(w, http.StatusOK, widget)
}

func (h *BirthdayHandler) ClubBirthdaysToday(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.birthdays.ClubBirthdaysToday(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load club birthdays")
		return
	}

	WriteJSON(w, http.StatusOK, items)
}
