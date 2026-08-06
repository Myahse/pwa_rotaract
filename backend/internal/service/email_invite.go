package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/email"
	"github.com/rotaract-civ/backend/internal/repository"
)

var (
	ErrInvalidInviteToken  = errors.New("invalid or expired invite token")
	ErrInviteEmailMismatch = errors.New("email does not match the invitation")
	ErrPendingEmailInvite  = errors.New("a pending email invite already exists for this address")
)

type RegisterInput struct {
	MemberProfileInput
	Token      string    `json:"token"`
	InviteCode string    `json:"invite_code"`
	RoleID     uuid.UUID `json:"role_id"`
}

type SendEmailInviteInput struct {
	Email  string     `json:"email"`
	RoleID *uuid.UUID `json:"role_id"`
}

func (s *RegistrationService) SendEmailInvite(ctx context.Context, clubID, invitedBy uuid.UUID, input SendEmailInviteInput) (*domain.EmailInvite, error) {
	emailAddr := strings.ToLower(strings.TrimSpace(input.Email))
	if emailAddr == "" {
		return nil, fmt.Errorf("email is required")
	}

	if _, err := s.users.GetByEmail(ctx, emailAddr); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	pending, err := s.emailInvites.HasPending(ctx, clubID, emailAddr)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, ErrPendingEmailInvite
	}

	club, err := s.clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	if input.RoleID != nil {
		role, err := s.clubs.GetRoleByID(ctx, clubID, *input.RoleID)
		if err != nil {
			return nil, err
		}
		if role.Name == "Responsable de club" {
			return nil, ErrRoleNotJoinable
		}
	}

	expiresAt := time.Now().UTC().Add(s.inviteTTL)
	invite := &domain.EmailInvite{
		ClubID:    clubID,
		Email:     emailAddr,
		RoleID:    input.RoleID,
		InvitedBy: invitedBy,
		ExpiresAt: expiresAt,
	}
	if err := s.emailInvites.Create(ctx, invite); err != nil {
		return nil, err
	}

	registerURL := strings.TrimRight(s.publicURL, "/") + "/register?token=" + invite.Token
	if err := s.mailer.SendRegistrationInvite(email.RegistrationInvite{
		To:          emailAddr,
		ClubName:    club.Name,
		RegisterURL: registerURL,
		ExpiresAt:   expiresAt.Format("02/01/2006 15:04"),
	}); err != nil {
		return nil, fmt.Errorf("send invite email: %w", err)
	}

	return invite, nil
}

func (s *RegistrationService) PreviewEmailInvite(ctx context.Context, token string) (*domain.EmailInvitePreview, error) {
	invite, err := s.emailInvites.GetByToken(ctx, token)
	if err != nil {
		return nil, ErrInvalidInviteToken
	}
	if !isInviteValid(invite, time.Now().UTC()) {
		return nil, ErrInvalidInviteToken
	}

	club, err := s.clubs.GetByID(ctx, invite.ClubID)
	if err != nil {
		return nil, err
	}

	roles, err := s.clubs.ListJoinableRoles(ctx, club.ID)
	if err != nil {
		return nil, err
	}

	preview := &domain.EmailInvitePreview{
		Club:      *club,
		Email:     invite.Email,
		Roles:     roles,
		ExpiresAt: invite.ExpiresAt,
	}

	if invite.RoleID != nil {
		role, err := s.clubs.GetRoleByID(ctx, club.ID, *invite.RoleID)
		if err == nil {
			preview.Role = role
		}
	}

	return preview, nil
}

func (s *RegistrationService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if strings.TrimSpace(input.Token) != "" {
		return s.registerWithToken(ctx, input)
	}
	if strings.TrimSpace(input.InviteCode) != "" {
		return s.RegisterWithInvite(ctx, RegisterWithInviteInput{
			MemberProfileInput: input.MemberProfileInput,
			InviteCode:         input.InviteCode,
			RoleID:             input.RoleID,
		})
	}
	return nil, fmt.Errorf("token or invite_code is required")
}

func (s *RegistrationService) registerWithToken(ctx context.Context, input RegisterInput) (*domain.User, error) {
	invite, err := s.emailInvites.GetByToken(ctx, input.Token)
	if err != nil {
		return nil, ErrInvalidInviteToken
	}
	if !isInviteValid(invite, time.Now().UTC()) {
		return nil, ErrInvalidInviteToken
	}

	emailAddr := strings.ToLower(strings.TrimSpace(input.Email))
	if emailAddr != invite.Email {
		return nil, ErrInviteEmailMismatch
	}

	roleID := input.RoleID
	if invite.RoleID != nil {
		if roleID != uuid.Nil && roleID != *invite.RoleID {
			return nil, fmt.Errorf("role is fixed by the invitation")
		}
		roleID = *invite.RoleID
	}
	if roleID == uuid.Nil {
		return nil, fmt.Errorf("role_id is required")
	}

	role, err := s.clubs.GetRoleByID(ctx, invite.ClubID, roleID)
	if err != nil {
		return nil, err
	}
	if role.Name == "Responsable de club" {
		return nil, ErrRoleNotJoinable
	}

	user, err := s.createMemberUser(ctx, input.MemberProfileInput)
	if err != nil {
		return nil, err
	}

	if err := s.joinClub(ctx, invite.ClubID, user, roleID, &invite.InvitedBy); err != nil {
		return nil, err
	}

	if err := s.emailInvites.MarkUsed(ctx, invite.ID); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func isInviteValid(invite *domain.EmailInvite, now time.Time) bool {
	if invite.UsedAt != nil {
		return false
	}
	return now.Before(invite.ExpiresAt)
}

func (s *RegistrationService) ListEmailInvites(ctx context.Context, clubID uuid.UUID) ([]domain.EmailInvite, error) {
	return s.emailInvites.ListByClub(ctx, clubID)
}
