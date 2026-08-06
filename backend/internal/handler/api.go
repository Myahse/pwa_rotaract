package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.Login(r.Context(), input)
	if errors.Is(err, service.ErrInvalidCredentials) {
		WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := authctx.UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	user.PasswordHash = ""
	WriteJSON(w, http.StatusOK, user)
}

type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

func (h *AdminHandler) CreateClubHead(w http.ResponseWriter, r *http.Request) {
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

	var input service.CreateHeadMemberInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	head, err := h.admin.CreateClubHead(r.Context(), user.ID, clubID, input)
	if errors.Is(err, service.ErrHeadAlreadyExists) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, head)
}

func (h *AdminHandler) ListClubs(w http.ResponseWriter, r *http.Request) {
	clubs, err := h.admin.ListClubs(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list clubs")
		return
	}
	WriteJSON(w, http.StatusOK, clubs)
}

type ClubHandler struct {
	clubs    *service.ClubService
	repo     *repository.ClubRepository
	profiles *service.ProfileService
}

func NewClubHandler(clubs *service.ClubService, repo *repository.ClubRepository, profiles *service.ProfileService) *ClubHandler {
	return &ClubHandler{clubs: clubs, repo: repo, profiles: profiles}
}

func (h *ClubHandler) GetClub(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	club, err := h.repo.GetByID(r.Context(), clubID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "club not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get club")
		return
	}

	WriteJSON(w, http.StatusOK, h.profiles.PublicClub(club))
}

func (h *ClubHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	members, err := h.repo.ListMembers(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	for i := range members {
		if members[i].User != nil {
			members[i].User = h.profiles.PublicUser(members[i].User)
		}
	}

	WriteJSON(w, http.StatusOK, members)
}

func (h *ClubHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	roles, err := h.repo.ListRoles(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}

	WriteJSON(w, http.StatusOK, roles)
}

func (h *ClubHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.CreateRoleInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := h.clubs.CreateRole(r.Context(), clubID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, role)
}

func (h *ClubHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
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

	userID, err := UserIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.AssignRoleInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.clubs.AssignRole(r.Context(), user.ID, clubID, userID, input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

func (h *ClubHandler) CreateCommission(w http.ResponseWriter, r *http.Request) {
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

	var input service.CreateCommissionInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	commission, err := h.clubs.CreateCommission(r.Context(), user.ID, clubID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, commission)
}

func (h *ClubHandler) ListCommissions(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	commissions, err := h.clubs.ListCommissions(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list commissions")
		return
	}

	WriteJSON(w, http.StatusOK, commissions)
}

func (h *ClubHandler) AddCommissionMember(w http.ResponseWriter, r *http.Request) {
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

	var input service.AddCommissionMemberInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member, err := h.clubs.AddCommissionMember(r.Context(), clubID, commissionID, input)
	if errors.Is(err, service.ErrDuplicatePresident) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, service.ErrDuplicateSecretary) {
		WriteError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, member)
}

type ChatHandler struct {
	chat     *service.ChatService
	chatRepo *repository.ChatRepository
	clubRepo *repository.ClubRepository
}

func NewChatHandler(chat *service.ChatService, chatRepo *repository.ChatRepository, clubRepo *repository.ClubRepository) *ChatHandler {
	return &ChatHandler{chat: chat, chatRepo: chatRepo, clubRepo: clubRepo}
}

func (h *ChatHandler) ListClubGroups(w http.ResponseWriter, r *http.Request) {
	clubID, err := ClubIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	groups, err := h.chatRepo.ListByClub(r.Context(), clubID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list chat groups")
		return
	}

	WriteJSON(w, http.StatusOK, groups)
}

func (h *ChatHandler) CreateCommissionGroup(w http.ResponseWriter, r *http.Request) {
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

	commissionID, err := CommissionIDFromRequest(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.CreateCommissionGroupInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.chat.CreateCommissionGroup(r.Context(), user.ID, clubID, commissionID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, group)
}

func (h *ChatHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
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

	group, err := h.chatRepo.GetByID(r.Context(), groupID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "group not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get group")
		return
	}

	canAccess, err := h.chatRepo.UserCanAccessGroup(r.Context(), groupID, user.ID, user.IsAdmin)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to verify group access")
		return
	}
	if !canAccess && user.IsAdmin {
		canAccess = true
	}
	if !canAccess {
		isHead, headErr := h.clubRepo.IsHead(r.Context(), group.ClubID, user.ID)
		if headErr == nil && isHead {
			canAccess = true
		}
	}
	if !canAccess {
		WriteError(w, http.StatusForbidden, "access denied to this group")
		return
	}

	members, err := h.chatRepo.ListMembers(r.Context(), groupID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list group members")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"group":   group,
		"members": members,
	})
}

func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
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

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, parseErr := strconv.Atoi(raw); parseErr == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	var before *time.Time
	if raw := r.URL.Query().Get("before"); raw != "" {
		t, parseErr := time.Parse(time.RFC3339, raw)
		if parseErr != nil {
			WriteError(w, http.StatusBadRequest, "invalid before timestamp")
			return
		}
		before = &t
	}

	messages, err := h.chat.ListMessages(r.Context(), user, groupID, limit, before)
	if err != nil {
		WriteError(w, http.StatusForbidden, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
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

	var input service.SendMessageInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	msg, err := h.chat.SendMessage(r.Context(), user, groupID, input)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, msg)
}
