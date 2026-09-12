package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/service"
)

type ClubDueHandler struct {
	dues *service.ClubDueService
}

func NewClubDueHandler(dues *service.ClubDueService) *ClubDueHandler {
	return &ClubDueHandler{dues: dues}
}

func (h *ClubDueHandler) Submit(w http.ResponseWriter, r *http.Request) {
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
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	amount, err := parseAmountField(r.FormValue("amount_xof"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "receipt file is required")
		return
	}
	defer file.Close()

	input := service.ClubDueInput{
		DueMonth:  r.FormValue("due_month"),
		AmountXOF: amount,
	}
	due, err := h.dues.Submit(r.Context(), user, clubID, input, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if errors.Is(err, service.ErrReceiptTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, due)
}

func (h *ClubDueHandler) List(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.dues.List(r.Context(), user, clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list dues")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ClubDueHandler) MarkReceived(w http.ResponseWriter, r *http.Request) {
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
	dueID, err := uuid.Parse(chi.URLParam(r, "dueID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid due id")
		return
	}
	due, err := h.dues.MarkReceived(r.Context(), user, clubID, dueID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, due)
}

func parseAmountField(raw string) (int, error) {
	amount, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || amount <= 0 {
		return 0, errors.New("invalid amount")
	}
	return amount, nil
}
