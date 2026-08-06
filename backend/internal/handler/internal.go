package handler

import (
	"net/http"

	"github.com/rotaract-civ/backend/internal/scheduler"
)

type InternalHandler struct {
	birthdayScheduler *scheduler.BirthdayScheduler
	cronSecret        string
}

func NewInternalHandler(birthdayScheduler *scheduler.BirthdayScheduler, cronSecret string) *InternalHandler {
	return &InternalHandler{
		birthdayScheduler: birthdayScheduler,
		cronSecret:        cronSecret,
	}
}

func (h *InternalHandler) RunBirthdays(w http.ResponseWriter, r *http.Request) {
	if h.cronSecret == "" || r.Header.Get("X-Cron-Secret") != h.cronSecret {
		WriteError(w, http.StatusUnauthorized, "invalid cron secret")
		return
	}

	count, err := h.birthdayScheduler.RunNow(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "birthday job failed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"sent":   count,
	})
}
