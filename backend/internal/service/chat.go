package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/ws"
)

var (
	ErrChatAccessDenied = errors.New("access denied to chat group")
	ErrEmptyMessage     = errors.New("message content is required")
)

type ChatService struct {
	clubs       *repository.ClubRepository
	commissions *repository.CommissionRepository
	chat        *repository.ChatRepository
	profiles    *ProfileService
	hub         *ws.Hub
}

func NewChatService(
	clubs *repository.ClubRepository,
	commissions *repository.CommissionRepository,
	chat *repository.ChatRepository,
	profiles *ProfileService,
	hub *ws.Hub,
) *ChatService {
	return &ChatService{
		clubs:       clubs,
		commissions: commissions,
		chat:        chat,
		profiles:    profiles,
		hub:         hub,
	}
}

func (s *ChatService) Hub() *ws.Hub {
	return s.hub
}

type SendMessageInput struct {
	Content string `json:"content"`
}

func (s *ChatService) SendMessage(ctx context.Context, user *domain.User, groupID uuid.UUID, input SendMessageInput) (*domain.ChatMessage, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrEmptyMessage
	}
	if len(content) > 4000 {
		return nil, fmt.Errorf("message is too long")
	}

	if err := s.ensureGroupAccess(ctx, user, groupID); err != nil {
		return nil, err
	}

	msg := &domain.ChatMessage{
		GroupID: groupID,
		UserID:  user.ID,
		Content: content,
	}
	if err := s.chat.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	if s.profiles != nil {
		msg.User = s.profiles.PublicUser(user)
	} else {
		user.PasswordHash = ""
		msg.User = user
	}

	if s.hub != nil {
		s.hub.BroadcastMessage(groupID, *msg)
	}

	return msg, nil
}

func (s *ChatService) broadcastUpdated(groupID uuid.UUID, message domain.ChatMessage) {
	if s.hub != nil {
		s.hub.BroadcastEvent(groupID, ws.Event{
			Type:    "message_updated",
			GroupID: groupID,
			Message: &message,
		})
	}
}

func (s *ChatService) broadcastDeleted(groupID, messageID uuid.UUID) {
	if s.hub != nil {
		s.hub.BroadcastEvent(groupID, ws.Event{
			Type:      "message_deleted",
			GroupID:   groupID,
			MessageID: &messageID,
		})
	}
}

func (s *ChatService) ListMessages(ctx context.Context, user *domain.User, groupID uuid.UUID, limit int, before *time.Time) ([]domain.ChatMessage, error) {
	if err := s.ensureGroupAccess(ctx, user, groupID); err != nil {
		return nil, err
	}

	messages, err := s.chat.ListMessages(ctx, groupID, limit, before)
	if err != nil {
		return nil, err
	}

	if s.profiles != nil {
		for i := range messages {
			if messages[i].User != nil {
				messages[i].User = s.profiles.PublicUser(messages[i].User)
			}
		}
	}

	return messages, nil
}

func (s *ChatService) EnsureGroupAccess(ctx context.Context, groupID, userID uuid.UUID, isAdmin bool) error {
	return s.ensureGroupAccess(ctx, &domain.User{ID: userID, IsAdmin: isAdmin}, groupID)
}

func (s *ChatService) ensureGroupAccess(ctx context.Context, user *domain.User, groupID uuid.UUID) error {
	if user.IsAdmin {
		return nil
	}

	canAccess, err := s.chat.UserCanAccessGroup(ctx, groupID, user.ID, user.IsAdmin)
	if err != nil {
		return err
	}
	if canAccess {
		return nil
	}

	group, err := s.chat.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	isHead, err := s.clubs.IsHead(ctx, group.ClubID, user.ID)
	if err != nil {
		return err
	}
	if isHead {
		return nil
	}

	return ErrChatAccessDenied
}

type CreateCommissionGroupInput struct {
	Name string `json:"name"`
}

func (s *ChatService) CreateCommissionGroup(ctx context.Context, actorID, clubID, commissionID uuid.UUID, input CreateCommissionGroupInput) (*domain.ChatGroup, error) {
	commission, err := s.commissions.GetByID(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if commission.ClubID != clubID {
		return nil, repository.ErrNotFound
	}

	if _, err := s.chat.GetCommissionGroup(ctx, commissionID); err == nil {
		return nil, fmt.Errorf("commission group already exists")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = commission.Name + " - Groupe"
	}

	group := &domain.ChatGroup{
		ClubID:       clubID,
		CommissionID: &commissionID,
		GroupType:    domain.ChatGroupTypeCommission,
		Name:         name,
		CreatedBy:    &actorID,
	}
	if err := s.chat.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	members, err := s.commissions.ListMembers(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if err := s.chat.AddMember(ctx, group.ID, member.UserID); err != nil {
			return nil, err
		}
	}

	return group, nil
}

type CreateClubGroupInput struct {
	Name      string      `json:"name"`
	MemberIDs []uuid.UUID `json:"member_ids"`
}

func (s *ChatService) ListInbox(ctx context.Context, user *domain.User, clubID uuid.UUID) ([]domain.ChatInboxGroup, error) {
	if _, err := s.clubs.GetMembership(ctx, clubID, user.ID); err != nil && !user.IsAdmin {
		return nil, ErrChatAccessDenied
	}
	return s.chat.ListInboxForUser(ctx, clubID, user.ID)
}

type CreateDirectChatInput struct {
	UserID uuid.UUID `json:"user_id"`
}

func (s *ChatService) CreateDirectChat(ctx context.Context, actorID, clubID uuid.UUID, input CreateDirectChatInput) (*domain.ChatGroup, error) {
	if input.UserID == actorID {
		return nil, fmt.Errorf("cannot start a direct chat with yourself")
	}
	if _, err := s.clubs.GetMembership(ctx, clubID, actorID); err != nil {
		return nil, ErrChatAccessDenied
	}
	if _, err := s.clubs.GetMembership(ctx, clubID, input.UserID); err != nil {
		return nil, fmt.Errorf("user must belong to the club")
	}

	if existing, err := s.chat.FindDirectGroup(ctx, clubID, actorID, input.UserID); err == nil {
		return existing, nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	name := "Conversation"
	if s.profiles != nil {
		if otherUser, err := s.profiles.GetProfile(ctx, input.UserID); err == nil {
			name = strings.TrimSpace(otherUser.FirstName + " " + otherUser.LastName)
		}
	}
	group := &domain.ChatGroup{
		ClubID:    clubID,
		GroupType: domain.ChatGroupTypeCustom,
		Name:      name,
		CreatedBy: &actorID,
	}
	if err := s.chat.CreateGroup(ctx, group); err != nil {
		return nil, err
	}
	for _, userID := range []uuid.UUID{actorID, input.UserID} {
		if err := s.chat.AddMember(ctx, group.ID, userID); err != nil {
			return nil, err
		}
	}
	return group, nil
}

type AddGroupMembersInput struct {
	MemberIDs []uuid.UUID `json:"member_ids"`
}

func (s *ChatService) AddGroupMembers(ctx context.Context, actor *domain.User, groupID uuid.UUID, input AddGroupMembersInput) (*domain.ChatGroup, error) {
	group, err := s.chat.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.GroupType != domain.ChatGroupTypeCustom {
		return nil, fmt.Errorf("members can only be added to custom groups")
	}
	if err := s.ensureGroupManageAccess(ctx, actor, group); err != nil {
		return nil, err
	}

	memberSet := map[uuid.UUID]struct{}{}
	for _, id := range input.MemberIDs {
		memberSet[id] = struct{}{}
	}
	for userID := range memberSet {
		if _, err := s.clubs.GetMembership(ctx, group.ClubID, userID); err != nil {
			continue
		}
		if err := s.chat.AddMember(ctx, group.ID, userID); err != nil {
			return nil, err
		}
	}
	return group, nil
}

func (s *ChatService) ensureGroupManageAccess(ctx context.Context, user *domain.User, group *domain.ChatGroup) error {
	if user.IsAdmin {
		return nil
	}
	isHead, err := s.clubs.IsHead(ctx, group.ClubID, user.ID)
	if err != nil {
		return err
	}
	if isHead {
		return nil
	}
	if group.CreatedBy != nil && *group.CreatedBy == user.ID {
		return nil
	}
	return ErrChatAccessDenied
}

func (s *ChatService) CreateClubGroup(ctx context.Context, actorID, clubID uuid.UUID, input CreateClubGroupInput) (*domain.ChatGroup, error) {
	if _, err := s.clubs.GetMembership(ctx, clubID, actorID); err != nil {
		return nil, ErrChatAccessDenied
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	group := &domain.ChatGroup{
		ClubID:    clubID,
		GroupType: domain.ChatGroupTypeCustom,
		Name:      name,
		CreatedBy: &actorID,
	}
	if err := s.chat.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	memberSet := map[uuid.UUID]struct{}{actorID: {}}
	for _, id := range input.MemberIDs {
		memberSet[id] = struct{}{}
	}
	for userID := range memberSet {
		if _, err := s.clubs.GetMembership(ctx, clubID, userID); err != nil {
			continue
		}
		if err := s.chat.AddMember(ctx, group.ID, userID); err != nil {
			return nil, err
		}
	}
	return group, nil
}
