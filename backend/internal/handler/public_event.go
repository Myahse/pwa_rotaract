package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type PublicEventHandler struct {
	events *service.PublicEventService
}

func NewPublicEventHandler(events *service.PublicEventService) *PublicEventHandler {
	return &PublicEventHandler{events: events}
}

func (h *PublicEventHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	items, err := h.events.List(r.Context(), true)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *PublicEventHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	item, err := h.events.Get(r.Context(), id, true)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load event")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicEventHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.events.List(r.Context(), false)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *PublicEventHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	h.writeEvent(w, r, false)
}

func (h *PublicEventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.PublicEventInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.events.Create(r.Context(), input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *PublicEventHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	var input service.PublicEventInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.events.Update(r.Context(), id, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicEventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	if err := h.events.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "event not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *PublicEventHandler) UploadFlyer(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, true)
}

func (h *PublicEventHandler) AddImage(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, false)
}

func (h *PublicEventHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	eventID, ok := h.parseID(w, r)
	if !ok {
		return
	}
	imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid image id")
		return
	}
	item, err := h.events.DeleteImage(r.Context(), eventID, imageID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "image not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete image")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicEventHandler) upload(w http.ResponseWriter, r *http.Request, flyer bool) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	var item any
	if flyer {
		item, err = h.events.UploadFlyer(r.Context(), id, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	} else {
		item, err = h.events.AddImage(r.Context(), id, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	}
	if errors.Is(err, service.ErrInvalidImageType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrImageTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicEventHandler) writeEvent(w http.ResponseWriter, r *http.Request, publishedOnly bool) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	item, err := h.events.Get(r.Context(), id, publishedOnly)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load event")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicEventHandler) parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "eventID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid event id")
		return uuid.Nil, false
	}
	return id, true
}
