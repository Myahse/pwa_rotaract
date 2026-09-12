package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrHeadAlreadyExists  = errors.New("club already has a head member")
	ErrDuplicatePresident = errors.New("commission already has a president")
	ErrDuplicateSecretary = errors.New("commission already has a secretary")
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	AccessToken string       `json:"access_token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        domain.User  `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}
	if err := auth.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := s.tokens.Generate(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &LoginResult{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		User:        *user,
	}, nil
}

type AdminService struct {
	users        *repository.UserRepository
	clubs        *repository.ClubRepository
	commissions  *repository.CommissionRepository
	chat         *repository.ChatRepository
	diary        *DiaryService
	mandates     *MandateService
	profiles     *ProfileService
	store        *storage.LocalStore
	maxImageSize int64
}

func NewAdminService(
	users *repository.UserRepository,
	clubs *repository.ClubRepository,
	commissions *repository.CommissionRepository,
	chat *repository.ChatRepository,
	diary *DiaryService,
	mandates *MandateService,
	profiles *ProfileService,
	store *storage.LocalStore,
	maxImageSize int64,
) *AdminService {
	if maxImageSize <= 0 {
		maxImageSize = defaultMaxAvatarBytes
	}
	return &AdminService{
		users:        users,
		clubs:        clubs,
		commissions:  commissions,
		chat:         chat,
		diary:        diary,
		mandates:     mandates,
		profiles:     profiles,
		store:        store,
		maxImageSize: maxImageSize,
	}
}

type CreateClubInput struct {
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description"`
	Country     string     `json:"country"`
	City        string     `json:"city"`
	Commune     string     `json:"commune"`
	FoundedAt   *time.Time `json:"founded_at"`
}

func (s *AdminService) CreateClub(ctx context.Context, adminID uuid.UUID, input CreateClubInput) (*domain.Club, error) {
	club, err := s.createClubRecord(ctx, adminID, input)
	if err != nil {
		return nil, err
	}
	return s.profiles.PublicClub(club), nil
}

type CreateHeadMemberInput struct {
	Email       string     `json:"email"`
	Password    string     `json:"password"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Phone       *string    `json:"phone"`
	BirthDate   *time.Time `json:"birth_date"`
	Profession  *string    `json:"profession"`
	MemberSince *time.Time `json:"member_since"`
}

func (s *AdminService) CreateClubHead(ctx context.Context, adminID, clubID uuid.UUID, input CreateHeadMemberInput) (*domain.User, error) {
	if _, err := s.clubs.GetByID(ctx, clubID); err != nil {
		return nil, err
	}

	president, err := s.createPresidentUser(ctx, adminID, clubID, CreatePresidentInput{
		Email:       input.Email,
		Password:    input.Password,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Phone:       input.Phone,
		BirthDate:   input.BirthDate,
		Profession:  input.Profession,
		MemberSince: input.MemberSince,
	})
	if err != nil {
		return nil, err
	}
	return s.profiles.PublicUser(president), nil
}

func (s *AdminService) ListClubs(ctx context.Context) ([]domain.Club, error) {
	return s.PublicClubs(ctx)
}

type ClubService struct {
	users       *repository.UserRepository
	clubs       *repository.ClubRepository
	commissions *repository.CommissionRepository
	chat        *repository.ChatRepository
}

func NewClubService(
	users *repository.UserRepository,
	clubs *repository.ClubRepository,
	commissions *repository.CommissionRepository,
	chat *repository.ChatRepository,
) *ClubService {
	return &ClubService{
		users:       users,
		clubs:       clubs,
		commissions: commissions,
		chat:        chat,
	}
}

type CreateRoleInput struct {
	Name           string   `json:"name"`
	Description    *string  `json:"description"`
	PermissionKeys []string `json:"permission_keys"`
}

func (s *ClubService) CreateRole(ctx context.Context, clubID uuid.UUID, input CreateRoleInput) (*domain.ClubRole, error) {
	role := &domain.ClubRole{
		ClubID:      clubID,
		Name:        strings.TrimSpace(input.Name),
		Description: input.Description,
	}
	if err := s.clubs.CreateRole(ctx, role, input.PermissionKeys); err != nil {
		return nil, err
	}
	return role, nil
}

type AssignRoleInput struct {
	RoleID uuid.UUID `json:"role_id"`
}

func (s *ClubService) GetClubAccess(ctx context.Context, clubID, userID uuid.UUID) (*domain.ClubAccessSummary, error) {
	membership, err := s.clubs.GetMembership(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}
	assignments, err := s.clubs.ListUserRoleAssignments(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.clubs.ListUserPermissions(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}
	roles := make([]domain.ClubRole, 0, len(assignments))
	for _, item := range assignments {
		if item.Role != nil {
			roles = append(roles, *item.Role)
		}
	}
	commissions, err := s.commissions.ListByUserInClub(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}
	return &domain.ClubAccessSummary{
		MemberRole:  membership.MemberRole,
		Roles:       roles,
		Permissions: permissions,
		Commissions: commissions,
	}, nil
}

func (s *ClubService) ListRoleAssignments(ctx context.Context, clubID uuid.UUID) ([]domain.ClubMemberRoleAssignment, error) {
	return s.clubs.ListRoleAssignments(ctx, clubID)
}

func (s *ClubService) AssignRole(ctx context.Context, actorID, clubID, userID uuid.UUID, input AssignRoleInput) error {
	if _, err := s.clubs.GetMembership(ctx, clubID, userID); err != nil {
		return err
	}
	return s.clubs.AssignRole(ctx, &domain.ClubMemberRoleAssignment{
		ClubID:     clubID,
		UserID:     userID,
		ClubRoleID: input.RoleID,
		AssignedBy: &actorID,
	})
}

type CreateCommissionInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (s *ClubService) CreateCommission(ctx context.Context, actorID, clubID uuid.UUID, input CreateCommissionInput) (*domain.Commission, error) {
	commission := &domain.Commission{
		ClubID:      clubID,
		Name:        strings.TrimSpace(input.Name),
		Description: input.Description,
		CreatedBy:   &actorID,
	}
	if err := s.commissions.Create(ctx, commission); err != nil {
		return nil, err
	}
	return commission, nil
}

func (s *ClubService) ListCommissions(ctx context.Context, clubID uuid.UUID) ([]domain.Commission, error) {
	return s.commissions.ListByClub(ctx, clubID)
}

type AddCommissionMemberInput struct {
	UserID     uuid.UUID                   `json:"user_id"`
	MemberRole domain.CommissionMemberRole `json:"member_role"`
}

func (s *ClubService) AddCommissionMember(ctx context.Context, clubID, commissionID uuid.UUID, input AddCommissionMemberInput) (*domain.CommissionMembership, error) {
	commission, err := s.commissions.GetByID(ctx, commissionID)
	if err != nil {
		return nil, err
	}
	if commission.ClubID != clubID {
		return nil, repository.ErrNotFound
	}

	if _, err := s.clubs.GetMembership(ctx, clubID, input.UserID); err != nil {
		return nil, fmt.Errorf("user must belong to the club")
	}

	if input.MemberRole == domain.CommissionMemberRolePresident {
		count, err := s.commissions.CountPresidents(ctx, commissionID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicatePresident
		}
	}
	if input.MemberRole == domain.CommissionMemberRoleSecretary {
		count, err := s.commissions.CountSecretaries(ctx, commissionID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicateSecretary
		}
	}

	membership := &domain.CommissionMembership{
		CommissionID: commissionID,
		UserID:       input.UserID,
		MemberRole:   input.MemberRole,
	}
	if err := s.commissions.AddMember(ctx, membership); err != nil {
		return nil, err
	}
	return membership, nil
}

func (s *AdminService) findRoleByName(ctx context.Context, clubID uuid.UUID, name string) (*domain.ClubRole, error) {
	roles, err := s.clubs.ListRoles(ctx, clubID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.Name == name {
			return &role, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (s *ClubService) findRoleByName(ctx context.Context, clubID uuid.UUID, name string) (*domain.ClubRole, error) {
	roles, err := s.clubs.ListRoles(ctx, clubID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.Name == name {
			return &role, nil
		}
	}
	return nil, repository.ErrNotFound
}

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(rawSlug, fallbackName string) string {
	source := strings.TrimSpace(rawSlug)
	if source == "" {
		source = fallbackName
	}
	slug := strings.ToLower(source)
	slug = slugRegexp.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

type BootstrapService struct {
	users *repository.UserRepository
}

func NewBootstrapService(users *repository.UserRepository) *BootstrapService {
	return &BootstrapService{users: users}
}

func (s *BootstrapService) EnsureAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil
	}

	count, err := s.users.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	return s.users.Create(ctx, &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hash,
		FirstName:    "System",
		LastName:     "Admin",
		IsAdmin:      true,
		IsActive:     true,
	})
}
