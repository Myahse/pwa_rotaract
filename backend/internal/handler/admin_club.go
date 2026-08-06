package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/service"
)

func (h *AdminHandler) CreateClub(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		h.createClubMultipart(w, r, user.ID)
		return
	}

	var input service.CreateClubInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	club, err := h.admin.CreateClub(r.Context(), user.ID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, club)
}

func (h *AdminHandler) createClubMultipart(w http.ResponseWriter, r *http.Request, adminID uuid.UUID) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	input, err := parseCreateClubMultipart(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.admin.CreateClubFull(r.Context(), adminID, input)
	if errors.Is(err, service.ErrHeadAlreadyExists) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrInvalidImageType) {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrImageTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, result)
}

func parseCreateClubMultipart(r *http.Request) (service.CreateClubFullInput, error) {
	var input service.CreateClubFullInput
	input.Name = strings.TrimSpace(r.FormValue("name"))
	input.Slug = strings.TrimSpace(r.FormValue("slug"))
	input.Country = strings.TrimSpace(r.FormValue("country"))
	input.City = strings.TrimSpace(r.FormValue("city"))
	input.Commune = strings.TrimSpace(r.FormValue("commune"))
	if description := strings.TrimSpace(r.FormValue("description")); description != "" {
		input.Description = &description
	}
	if foundedAt, err := parseOptionalDate(r.FormValue("founded_at")); err != nil {
		return input, err
	} else if foundedAt != nil {
		input.FoundedAt = foundedAt
	}

	logo, err := readUploadedImage(r, "logo", true)
	if err != nil {
		return input, err
	}
	input.Logo = logo

	input.President.Email = strings.TrimSpace(r.FormValue("president_email"))
	input.President.Password = r.FormValue("president_password")
	input.President.FirstName = strings.TrimSpace(r.FormValue("president_first_name"))
	input.President.LastName = strings.TrimSpace(r.FormValue("president_last_name"))
	if phone := strings.TrimSpace(r.FormValue("president_phone")); phone != "" {
		input.President.Phone = &phone
	}
	if profession := strings.TrimSpace(r.FormValue("president_profession")); profession != "" {
		input.President.Profession = &profession
	}
	if birthDate, err := parseOptionalDate(r.FormValue("president_birth_date")); err != nil {
		return input, err
	} else if birthDate != nil {
		input.President.BirthDate = birthDate
	}
	if memberSince, err := parseOptionalDate(r.FormValue("president_member_since")); err != nil {
		return input, err
	} else if memberSince != nil {
		input.President.MemberSince = memberSince
	}

	avatar, err := readUploadedImage(r, "president_avatar", false)
	if err != nil {
		return input, err
	}
	input.President.Avatar = avatar

	return input, nil
}

func readUploadedImage(r *http.Request, field string, required bool) (*service.UploadedImage, error) {
	file, header, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		if required {
			return nil, errors.New(field + " file is required")
		}
		return nil, nil
	}
	if err != nil {
		return nil, errors.New("invalid " + field + " file")
	}
	defer file.Close()

	const maxBytes = 6 << 20
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, errors.New("failed to read " + field)
	}
	if int64(len(data)) > maxBytes {
		return nil, service.ErrImageTooLarge
	}

	return &service.UploadedImage{
		Filename:    header.Filename,
		Size:        int64(len(data)),
		ContentType: header.Header.Get("Content-Type"),
		Reader:      bytes.NewReader(data),
	}, nil
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}
	return &parsed, nil
}
