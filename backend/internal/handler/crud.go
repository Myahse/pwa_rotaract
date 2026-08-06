package handler

import (
	"errors"
	"net/http"

	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

func (h *AdminHandler) UpdateClub(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.UpdateClubInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	club, err := h.admin.UpdateClub(r.Context(), clubID, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, club)
}

func (h *AdminHandler) DeleteClub(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.admin.DeleteClub(r.Context(), clubID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

func (h *ClubHandler) GetCommission(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	commissionID, err := CommissionIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	commission, members, err := h.clubs.GetCommission(r.Context(), clubID, commissionID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "commission not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get commission")
		return
	}

	for i := range members {
		if members[i].User != nil {
			members[i].User = h.profiles.PublicUser(members[i].User)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"commission": commission,
		"members":    members,
	})
}

func (h *ClubHandler) UpdateCommission(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	commissionID, err := CommissionIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.UpdateCommissionInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	commission, err := h.clubs.UpdateCommission(r.Context(), clubID, commissionID, input)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "commission not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, commission)
}

func (h *ClubHandler) DeleteCommission(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	commissionID, err := CommissionIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.clubs.DeleteCommission(r.Context(), clubID, commissionID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "commission not found")
		return
	} else if errors.Is(err, service.ErrSystemCommissionProtected) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	} else if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ClubHandler) RemoveCommissionMember(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	commissionID, err := CommissionIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID, err := UserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.clubs.RemoveCommissionMember(r.Context(), clubID, commissionID, userID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "member not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *ClubHandler) UnassignRole(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID, err := UserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	roleID, err := RoleIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.clubs.UnassignRole(r.Context(), clubID, userID, roleID); errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "role assignment not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "unassigned"})
}

func (h *ChatHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	groupID, err := GroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	messageID, err := MessageIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.UpdateMessageInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	msg, err := h.chat.UpdateMessage(r.Context(), user, groupID, messageID, input)
	if errors.Is(err, service.ErrMessageNotOwned) || errors.Is(err, service.ErrMessageForbidden) {
		WriteError(w, http.StatusForbidden, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, msg)
}

func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	groupID, err := GroupIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	messageID, err := MessageIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.chat.DeleteMessage(r.Context(), user, groupID, messageID); errors.Is(err, service.ErrMessageForbidden) || errors.Is(err, service.ErrMessageNotOwned) {
		WriteError(w, http.StatusForbidden, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
