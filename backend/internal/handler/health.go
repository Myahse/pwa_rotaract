package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rotaract-civ/backend/internal/config"
	"github.com/rotaract-civ/backend/internal/database"
)

type HealthHandler struct {
	cfg       *config.Config
	store     *database.Store
	startedAt time.Time
}

func NewHealthHandler(cfg *config.Config, store *database.Store) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		store:     store,
		startedAt: time.Now().UTC(),
	}
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Env       string `json:"env"`
	Uptime    string `json:"uptime"`
	Timestamp string `json:"timestamp"`
}

type statusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Version string `json:"version"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   "rotaract-civ-api",
		Env:       h.cfg.Env,
		Uptime:    time.Since(h.startedAt).Round(time.Second).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not_ready",
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (h *HealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, statusResponse{
		Status:  "ok",
		Message: "Rotaract CIV API is running",
		Version: "0.2.0",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
