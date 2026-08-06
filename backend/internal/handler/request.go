package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ClubIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "clubID")
	if raw == "" {
		return uuid.Nil, errors.New("club id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid club id")
	}
	return id, nil
}

func CommissionIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "commissionID")
	if raw == "" {
		return uuid.Nil, errors.New("commission id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid commission id")
	}
	return id, nil
}

func UserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "userID")
	if raw == "" {
		return uuid.Nil, errors.New("user id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}
	return id, nil
}

func GroupIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "groupID")
	if raw == "" {
		return uuid.Nil, errors.New("group id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid group id")
	}
	return id, nil
}

func RoleIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "roleID")
	if raw == "" {
		return uuid.Nil, errors.New("role id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid role id")
	}
	return id, nil
}

func MessageIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "messageID")
	if raw == "" {
		return uuid.Nil, errors.New("message id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid message id")
	}
	return id, nil
}

func InviteIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "inviteID")
	if raw == "" {
		return uuid.Nil, errors.New("invite id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid invite id")
	}
	return id, nil
}

func DiaryEntryIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "entryID")
	if raw == "" {
		return uuid.Nil, errors.New("entry id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid entry id")
	}
	return id, nil
}

func MandateIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "mandateID")
	if raw == "" {
		return uuid.Nil, errors.New("mandate id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid mandate id")
	}
	return id, nil
}

func AssignmentIDFromRequest(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "assignmentID")
	if raw == "" {
		return uuid.Nil, errors.New("assignment id is required")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid assignment id")
	}
	return id, nil
}

func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(dst)
}
