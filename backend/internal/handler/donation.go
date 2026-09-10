package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type DonationHandler struct {
	donations *service.DonationService
}

func NewDonationHandler(donations *service.DonationService) *DonationHandler {
	return &DonationHandler{donations: donations}
}

func (h *DonationHandler) Submit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "receipt file is required")
		return
	}
	defer file.Close()

	amount, err := strconv.Atoi(strings.TrimSpace(r.FormValue("amount_xof")))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	input := service.DonationInput{
		Name:      r.FormValue("name"),
		Email:     r.FormValue("email"),
		AmountXOF: amount,
	}

	donation, err := h.donations.Submit(r.Context(), input, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if errors.Is(err, service.ErrInvalidReceiptType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrReceiptTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, donation)
}

func (h *DonationHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.donations.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list donations")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *DonationHandler) MarkReceived(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "donationID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid donation id")
		return
	}
	item, err := h.donations.MarkReceived(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusBadRequest, "donation not found or already received")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to mark donation received")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
