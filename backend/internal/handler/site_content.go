package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type SiteContentHandler struct {
	site *service.SiteContentService
}

func NewSiteContentHandler(site *service.SiteContentService) *SiteContentHandler {
	return &SiteContentHandler{site: site}
}

func (h *SiteContentHandler) ListGalleryPublic(w http.ResponseWriter, r *http.Request) {
	items, err := h.site.ListGallery(r.Context(), true)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list gallery")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *SiteContentHandler) GetFeaturedPostulantPublic(w http.ResponseWriter, r *http.Request) {
	item, err := h.site.GetFeaturedPostulantForDisplay(r.Context(), true)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "featured postulant not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load featured postulant")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *SiteContentHandler) ListGalleryAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.site.ListGallery(r.Context(), false)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list gallery")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *SiteContentHandler) CreateGallery(w http.ResponseWriter, r *http.Request) {
	var input service.SiteGalleryInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.site.CreateGallery(r.Context(), input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *SiteContentHandler) UpdateGallery(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseGalleryID(w, r)
	if !ok {
		return
	}
	var input service.SiteGalleryInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.site.UpdateGallery(r.Context(), id, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "gallery image not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *SiteContentHandler) UploadGalleryImage(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseGalleryID(w, r)
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

	item, err := h.site.UploadGalleryImage(r.Context(), id, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if errors.Is(err, service.ErrInvalidImageType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrImageTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "gallery image not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *SiteContentHandler) DeleteGallery(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseGalleryID(w, r)
	if !ok {
		return
	}
	if err := h.site.DeleteGallery(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "gallery image not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to delete gallery image")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SiteContentHandler) ListFeaturedPostulantsAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.site.ListFeaturedPostulants(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list featured postulants")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *SiteContentHandler) CreateFeaturedPostulant(w http.ResponseWriter, r *http.Request) {
	var input service.FeaturedPostulantInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.site.CreateFeaturedPostulant(r.Context(), input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *SiteContentHandler) UpdateFeaturedPostulant(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parsePostulantID(w, r)
	if !ok {
		return
	}
	var input service.FeaturedPostulantInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.site.UpdateFeaturedPostulant(r.Context(), id, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "featured postulant not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *SiteContentHandler) DeleteFeaturedPostulant(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parsePostulantID(w, r)
	if !ok {
		return
	}
	if err := h.site.DeleteFeaturedPostulant(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "featured postulant not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to delete featured postulant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SiteContentHandler) UploadPostulantFlyer(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parsePostulantID(w, r)
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

	item, err := h.site.UploadPostulantFlyer(r.Context(), id, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if errors.Is(err, service.ErrInvalidImageType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrImageTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "featured postulant not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *SiteContentHandler) parseGalleryID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "imageID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid gallery image id")
		return uuid.Nil, false
	}
	return id, true
}

func (h *SiteContentHandler) parsePostulantID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "postulantID"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid featured postulant id")
		return uuid.Nil, false
	}
	return id, true
}
