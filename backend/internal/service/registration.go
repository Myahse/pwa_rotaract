package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/email"
	"github.com/rotaract-civ/backend/internal/repository"
)

var (
	ErrInvalidInviteCode   = errors.New("invalid invite code")
	ErrRoleNotJoinable     = errors.New("role is not available for self registration")
	ErrEmailAlreadyUsed    = errors.New("email is already registered")
	ErrPendingRequest      = errors.New("a pending access request already exists for this email")
	ErrRequestNotPending   = errors.New("access request is not pending")
	ErrClubNotMatched      = errors.New("club could not be matched, admin review required")
)

type RegistrationService struct {
	users        *repository.UserRepository
	clubs        *repository.ClubRepository
	chat         *repository.ChatRepository
	requests     *repository.AccessRequestRepository
	emailInvites *repository.EmailInviteRepository
	mailer       *email.Client
	publicURL    string
	inviteTTL    time.Duration
}

func NewRegistrationService(
	users *repository.UserRepository,
	clubs *repository.ClubRepository,
	chat *repository.ChatRepository,
	requests *repository.AccessRequestRepository,
	emailInvites *repository.EmailInviteRepository,
	mailer *email.Client,
	publicURL string,
	inviteTTL time.Duration,
) *RegistrationService {
	return &RegistrationService{
		users:        users,
		clubs:        clubs,
		chat:         chat,
		requests:     requests,
		emailInvites: emailInvites,
		mailer:       mailer,
		publicURL:    publicURL,
		inviteTTL:    inviteTTL,
	}
}

type MemberProfileInput struct {
	Email       string     `json:"email"`
	Password    string     `json:"password"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Phone       *string    `json:"phone"`
	BirthDate   *time.Time `json:"birth_date"`
	Profession  *string    `json:"profession"`
	MemberSince *time.Time `json:"member_since"`
}

type RegisterWithInviteInput struct {
	MemberProfileInput
	InviteCode string    `json:"invite_code"`
	RoleID     uuid.UUID `json:"role_id"`
}

type AccessRequestInput struct {
	MemberProfileInput
	ClubName string     `json:"club_name"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type ApproveAccessRequestInput struct {
	ClubID *uuid.UUID `json:"club_id"`
	RoleID *uuid.UUID `json:"role_id"`
}

type InvitePreview struct {
	Club      domain.Club   `json:"club"`
	Roles     []domain.ClubRole `json:"roles"`
	InviteURL string        `json:"invite_url"`
}

func (s *RegistrationService) PreviewInvite(ctx context.Context, inviteCode, publicURL string) (*InvitePreview, error) {
	club, err := s.clubs.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, ErrInvalidInviteCode
	}

	roles, err := s.clubs.ListJoinableRoles(ctx, club.ID)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(publicURL, "/") + "/register?code=" + club.InviteCode
	return &InvitePreview{
		Club:      *club,
		Roles:     roles,
		InviteURL: url,
	}, nil
}

func (s *RegistrationService) RegisterWithInvite(ctx context.Context, input RegisterWithInviteInput) (*domain.User, error) {
	club, err := s.clubs.GetByInviteCode(ctx, input.InviteCode)
	if err != nil {
		return nil, ErrInvalidInviteCode
	}

	role, err := s.clubs.GetRoleByID(ctx, club.ID, input.RoleID)
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

	if err := s.joinClub(ctx, club.ID, user, role.ID, nil); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *RegistrationService) SubmitAccessRequest(ctx context.Context, input AccessRequestInput) (*domain.AccessRequest, error) {
	if strings.TrimSpace(input.ClubName) == "" {
		return nil, fmt.Errorf("club name is required")
	}

	if _, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email))); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	pending, err := s.requests.HasPendingByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, ErrPendingRequest
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	club, clubErr := s.clubs.FindByName(ctx, input.ClubName)
	var clubID *uuid.UUID
	if clubErr == nil {
		clubID = &club.ID
		if input.RoleID != nil {
			if _, err := s.clubs.GetRoleByID(ctx, club.ID, *input.RoleID); err != nil {
				return nil, err
			}
		}
	} else if !errors.Is(clubErr, repository.ErrNotFound) {
		return nil, clubErr
	}

	req := &domain.AccessRequest{
		ClubID:          clubID,
		ClubName:        strings.TrimSpace(input.ClubName),
		RequestedRoleID: input.RoleID,
		Email:           strings.ToLower(strings.TrimSpace(input.Email)),
		FirstName:       strings.TrimSpace(input.FirstName),
		LastName:        strings.TrimSpace(input.LastName),
		Phone:           input.Phone,
		BirthDate:       input.BirthDate,
		Profession:      input.Profession,
		MemberSince:     input.MemberSince,
	}

	if err := s.requests.Create(ctx, req, hash); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *RegistrationService) ApproveAccessRequest(ctx context.Context, reviewerID, requestID uuid.UUID, input ApproveAccessRequestInput) (*domain.User, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != domain.AccessRequestStatusPending {
		return nil, ErrRequestNotPending
	}

	targetClubID := req.ClubID
	if targetClubID == nil {
		if input.ClubID == nil {
			return nil, ErrClubNotMatched
		}
		targetClubID = input.ClubID
	}

	roleID := req.RequestedRoleID
	if roleID == nil {
		if input.RoleID == nil {
			return nil, fmt.Errorf("role_id is required to approve this request")
		}
		roleID = input.RoleID
	}

	if _, err := s.clubs.GetRoleByID(ctx, *targetClubID, *roleID); err != nil {
		return nil, err
	}

	if _, err := s.users.GetByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	passwordHash, err := s.requests.GetPasswordHash(ctx, requestID)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		BirthDate:    req.BirthDate,
		Profession:   req.Profession,
		MemberSince:  req.MemberSince,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.joinClub(ctx, *targetClubID, user, *roleID, &reviewerID); err != nil {
		return nil, err
	}

	if err := s.requests.UpdateStatus(ctx, requestID, domain.AccessRequestStatusApproved, reviewerID, nil); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *RegistrationService) RejectAccessRequest(ctx context.Context, reviewerID, requestID uuid.UUID, note *string) error {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.AccessRequestStatusPending {
		return ErrRequestNotPending
	}
	return s.requests.UpdateStatus(ctx, requestID, domain.AccessRequestStatusRejected, reviewerID, note)
}

func (s *RegistrationService) GetClubInviteLink(club *domain.Club, publicURL string) map[string]string {
	url := strings.TrimRight(publicURL, "/") + "/register?code=" + club.InviteCode
	return map[string]string{
		"invite_code": club.InviteCode,
		"invite_url":  url,
	}
}

func (s *RegistrationService) createMemberUser(ctx context.Context, input MemberProfileInput) (*domain.User, error) {
	if _, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email))); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: hash,
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Phone:        input.Phone,
		BirthDate:    input.BirthDate,
		Profession:   input.Profession,
		MemberSince:  input.MemberSince,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyUsed
		}
		return nil, err
	}
	return user, nil
}

func (s *RegistrationService) joinClub(ctx context.Context, clubID uuid.UUID, user *domain.User, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	if err := s.clubs.AddMembership(ctx, &domain.ClubMembership{
		ClubID:     clubID,
		UserID:     user.ID,
		MemberRole: domain.ClubMemberRoleMember,
	}); err != nil {
		return err
	}

	if err := s.clubs.AssignRole(ctx, &domain.ClubMemberRoleAssignment{
		ClubID:     clubID,
		UserID:     user.ID,
		ClubRoleID: roleID,
		AssignedBy: assignedBy,
	}); err != nil {
		return err
	}

	groups, err := s.chat.ListByClub(ctx, clubID)
	if err != nil {
		return err
	}
	for _, group := range groups {
		if group.GroupType == domain.ChatGroupTypeClub {
			if err := s.chat.AddMember(ctx, group.ID, user.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func GenerateInviteCode(slug string) (string, error) {
	prefix := strings.ToUpper(strings.ReplaceAll(slug, "-", ""))
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}

	suffix, err := randomAlphaNum(6)
	if err != nil {
		return "", err
	}

	return prefix + "-" + suffix, nil
}

func randomAlphaNum(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
