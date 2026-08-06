package handler

import (
	"errors"
	"net/http"

	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type DiaryHandler struct {
	diary    *service.DiaryService
	mandates *service.MandateService
}

func NewDiaryHandler(diary *service.DiaryService, mandates *service.MandateService) *DiaryHandler {
	return &DiaryHandler{diary: diary, mandates: mandates}
}

func (h *DiaryHandler) GetDiary(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	diary, err := h.mandates.GetClubDiary(r.Context(), clubID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load club diary")
		return
	}

	WriteJSON(w, http.StatusOK, diary)
}

func (h *DiaryHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
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

	var input service.CreateDiaryEntryInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := h.diary.CreateEntry(r.Context(), user.ID, clubID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, entry)
}

func (h *DiaryHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	entryID, err := DiaryEntryIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.UpdateDiaryEntryInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := h.diary.UpdateEntry(r.Context(), clubID, entryID, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "entry not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, entry)
}

func (h *DiaryHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	entryID, err := DiaryEntryIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.diary.DeleteEntry(r.Context(), clubID, entryID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "entry not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete entry")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *DiaryHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	entryID, err := DiaryEntryIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := r.ParseMultipartForm(6 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "photo file is required")
		return
	}
	defer file.Close()

	entry, err := h.diary.UploadPhoto(r.Context(), clubID, entryID, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if errors.Is(err, service.ErrInvalidImageType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrImageTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "entry not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to upload photo")
		return
	}

	WriteJSON(w, http.StatusOK, entry)
}

func (h *DiaryHandler) CreateMandate(w http.ResponseWriter, r *http.Request) {
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
	var input service.CreateMandateInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	mandate, err := h.mandates.CreateMandate(r.Context(), user.ID, clubID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, mandate)
}

func (h *DiaryHandler) DeleteMandate(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	mandateID, err := MandateIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.mandates.DeleteMandate(r.Context(), clubID, mandateID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "mandate not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete mandate")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *DiaryHandler) CreateMandateAssignment(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	mandateID, err := MandateIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input service.CreateMandateAssignmentInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	assignment, err := h.mandates.CreateAssignment(r.Context(), clubID, mandateID, input)
	if errors.Is(err, service.ErrDuplicateExecutiveRole) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, assignment)
}

func (h *DiaryHandler) DeleteMandateAssignment(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	mandateID, err := MandateIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	assignmentID, err := AssignmentIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.mandates.DeleteAssignment(r.Context(), clubID, mandateID, assignmentID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "assignment not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete assignment")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
