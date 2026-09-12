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
)

var (
	ErrMessageNotOwned            = errors.New("can only modify your own messages")
	ErrMessageForbidden           = errors.New("not allowed to modify this message")
	ErrInviteAlreadyUsed          = errors.New("invite has already been used")
	ErrSystemCommissionProtected  = errors.New("system commission cannot be deleted")
)

type UpdateClubInput struct {
	Name        *string    `json:"name"`
	Slug        *string    `json:"slug"`
	Description *string    `json:"description"`
	Country     *string    `json:"country"`
	City        *string    `json:"city"`
	Commune     *string    `json:"commune"`
	FoundedAt   *time.Time `json:"founded_at"`
	IsActive    *bool      `json:"is_active"`
}

func (s *AdminService) UpdateClub(ctx context.Context, clubID uuid.UUID, input UpdateClubInput) (*domain.Club, error) {
	club, err := s.clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, fmt.Errorf("club name cannot be empty")
		}
		club.Name = trimmed
	}
	if input.Slug != nil {
		club.Slug = slugify(*input.Slug, club.Name)
	}
	if input.Description != nil {
		club.Description = input.Description
	}
	if input.Country != nil {
		trimmed := strings.TrimSpace(*input.Country)
		if trimmed == "" {
			return nil, fmt.Errorf("country cannot be empty")
		}
		club.Country = &trimmed
	}
	if input.City != nil {
		trimmed := strings.TrimSpace(*input.City)
		if trimmed == "" {
			return nil, fmt.Errorf("city cannot be empty")
		}
		club.City = &trimmed
	}
	if input.Commune != nil {
		trimmed := strings.TrimSpace(*input.Commune)
		if trimmed == "" {
			return nil, fmt.Errorf("commune cannot be empty")
		}
		club.Commune = &trimmed
	}
	if input.FoundedAt != nil {
		club.FoundedAt = input.FoundedAt
	}
	if input.IsActive != nil {
		club.IsActive = *input.IsActive
	}
	updated, err := s.clubs.Update(ctx, club)
	if err != nil {
		return nil, err
	}
	return s.profiles.PublicClub(updated), nil
}

func (s *AdminService) DeleteClub(ctx context.Context, clubID uuid.UUID) error {
	club, err := s.clubs.GetByID(ctx, clubID)
	if err != nil {
		return err
	}
	club.IsActive = false
	_, err = s.clubs.Update(ctx, club)
	return err
}

type UpdateCommissionInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (s *ClubService) GetCommission(ctx context.Context, clubID, commissionID uuid.UUID) (*domain.Commission, []domain.CommissionMembership, error) {
	commission, err := s.commissions.GetByID(ctx, commissionID)
	if err != nil {
		return nil, nil, err
	}
	if commission.ClubID != clubID {
		return nil, nil, repository.ErrNotFound
	}
	members, err := s.commissions.ListMembers(ctx, commissionID)
	if err != nil {
		return nil, nil, err
	}
	return commission, members, nil
}

func (s *ClubService) UpdateCommission(ctx context.Context, clubID, commissionID uuid.UUID, input UpdateCommissionInput) (*domain.Commission, error) {
	commission, err := s.commissions.GetByID(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if commission.ClubID != clubID {
		return nil, repository.ErrNotFound
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, fmt.Errorf("commission name cannot be empty")
		}
		commission.Name = trimmed
	}
	if input.Description != nil {
		commission.Description = input.Description
	}
	return s.commissions.Update(ctx, commission)
}

func (s *ClubService) DeleteCommission(ctx context.Context, clubID, commissionID uuid.UUID) error {
	commission, err := s.commissions.GetByID(ctx, commissionID)
	if err != nil {
		return err
	}
	if commission.ClubID != clubID {
		return repository.ErrNotFound
	}
	if commission.IsSystem {
		return ErrSystemCommissionProtected
	}
	return s.commissions.Delete(ctx, commissionID)
}

func (s *ClubService) RemoveCommissionMember(ctx context.Context, clubID, commissionID, userID uuid.UUID) error {
	commission, err := s.commissions.GetByID(ctx, commissionID)
	if err != nil {
		return err
	}
	if commission.ClubID != clubID {
		return repository.ErrNotFound
	}
	return s.commissions.RemoveMember(ctx, commissionID, userID)
}

func (s *ClubService) UnassignRole(ctx context.Context, clubID, userID, roleID uuid.UUID) error {
	if _, err := s.clubs.GetRoleByID(ctx, clubID, roleID); err != nil {
		return err
	}
	return s.clubs.UnassignRole(ctx, clubID, userID, roleID)
}

func (s *RegistrationService) RevokeEmailInvite(ctx context.Context, clubID, inviteID uuid.UUID) error {
	invite, err := s.emailInvites.GetByID(ctx, clubID, inviteID)
	if err != nil {
		return err
	}
	if invite.UsedAt != nil {
		return ErrInviteAlreadyUsed
	}
	return s.emailInvites.RevokePending(ctx, clubID, inviteID)
}

type UpdateMessageInput struct {
	Content string `json:"content"`
}

func (s *ChatService) UpdateMessage(ctx context.Context, user *domain.User, groupID, messageID uuid.UUID, input UpdateMessageInput) (*domain.ChatMessage, error) {
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

	msg, err := s.chat.GetMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg.GroupID != groupID {
		return nil, repository.ErrNotFound
	}
	if msg.UserID != user.ID && !user.IsAdmin {
		return nil, ErrMessageNotOwned
	}

	if err := s.chat.UpdateMessageContent(ctx, messageID, content); err != nil {
		return nil, err
	}
	updated, err := s.chat.GetMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if s.profiles != nil && updated.User != nil {
		updated.User = s.profiles.PublicUser(updated.User)
	}
	s.broadcastUpdated(groupID, *updated)
	return updated, nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, user *domain.User, groupID, messageID uuid.UUID) error {
	if err := s.ensureGroupAccess(ctx, user, groupID); err != nil {
		return err
	}

	msg, err := s.chat.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.GroupID != groupID {
		return repository.ErrNotFound
	}
	if msg.UserID == user.ID || user.IsAdmin {
		if err := s.chat.DeleteMessage(ctx, messageID); err != nil {
			return err
		}
		s.broadcastDeleted(groupID, messageID)
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
		if err := s.chat.DeleteMessage(ctx, messageID); err != nil {
			return err
		}
		s.broadcastDeleted(groupID, messageID)
		return nil
	}
	return ErrMessageForbidden
}
