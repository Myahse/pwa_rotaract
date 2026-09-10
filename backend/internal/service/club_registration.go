package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/email"
	"github.com/rotaract-civ/backend/internal/repository"
)

var (
	ErrClubRegistrationNotPending   = errors.New("club registration request is not pending")
	ErrPendingClubRegistration      = errors.New("a pending club registration already exists for this email")
	ErrPendingClubName              = errors.New("a pending club registration already exists for this club name")
	ErrClubNameTaken                = errors.New("a club with this name already exists")
	ErrInvalidClubAccessToken       = errors.New("invalid or expired club registration link")
	ErrClubAccessAlreadyCompleted   = errors.New("club registration has already been completed")
	ErrGoogleEmailMismatch          = errors.New("google account email must match the approved contact email")
	ErrGoogleAuthNotConfigured      = errors.New("google sign-in is not configured")
)

type ClubRegistrationService struct {
	requests     *repository.ClubRegistrationRepository
	users        *repository.UserRepository
	clubs        *repository.ClubRepository
	admin        *AdminService
	mailer       *email.Client
	tokens       *auth.TokenManager
	publicURL    string
	inviteTTL    time.Duration
	googleClient string
}

func NewClubRegistrationService(
	requests *repository.ClubRegistrationRepository,
	users *repository.UserRepository,
	clubs *repository.ClubRepository,
	admin *AdminService,
	mailer *email.Client,
	tokens *auth.TokenManager,
	publicURL string,
	inviteTTL time.Duration,
	googleClientID string,
) *ClubRegistrationService {
	return &ClubRegistrationService{
		requests:     requests,
		users:        users,
		clubs:        clubs,
		admin:        admin,
		mailer:       mailer,
		tokens:       tokens,
		publicURL:    publicURL,
		inviteTTL:    inviteTTL,
		googleClient: strings.TrimSpace(googleClientID),
	}
}

type SubmitClubRegistrationInput struct {
	ClubName         string     `json:"club_name"`
	ContactEmail     string     `json:"contact_email"`
	ContactFirstName string     `json:"contact_first_name"`
	ContactLastName  string     `json:"contact_last_name"`
	Phone            *string    `json:"phone"`
	Description      *string    `json:"description"`
	Country          string     `json:"country"`
	City             string     `json:"city"`
	Commune          string     `json:"commune"`
	FoundedAt        *time.Time `json:"founded_at"`
	Message          *string    `json:"message"`
}

type ApproveClubRegistrationInput struct {
	Slug       *string `json:"slug"`
	Country    *string `json:"country"`
	City       *string `json:"city"`
	Commune    *string `json:"commune"`
	ReviewNote *string `json:"review_note"`
}

type ApproveClubRegistrationResult struct {
	Request   domain.ClubRegistrationRequest `json:"request"`
	Message   string                         `json:"message"`
	ExpiresAt time.Time                      `json:"expires_at"`
}

type CompleteClubRegistrationInput struct {
	Token         string `json:"token"`
	GoogleIDToken string `json:"google_id_token"`
	Password      string `json:"password"`
}

type CompleteClubRegistrationResult struct {
	AccessToken string      `json:"access_token"`
	ExpiresAt   time.Time   `json:"expires_at"`
	User        domain.User `json:"user"`
	Club        domain.Club `json:"club"`
}

func (s *ClubRegistrationService) Submit(ctx context.Context, input SubmitClubRegistrationInput) (*domain.ClubRegistrationRequest, error) {
	clubName := strings.TrimSpace(input.ClubName)
	emailAddr := strings.ToLower(strings.TrimSpace(input.ContactEmail))
	firstName := strings.TrimSpace(input.ContactFirstName)
	lastName := strings.TrimSpace(input.ContactLastName)
	country := strings.TrimSpace(input.Country)
	city := strings.TrimSpace(input.City)
	commune := strings.TrimSpace(input.Commune)

	if clubName == "" {
		return nil, fmt.Errorf("club_name is required")
	}
	if emailAddr == "" {
		return nil, fmt.Errorf("contact_email is required")
	}
	if firstName == "" || lastName == "" {
		return nil, fmt.Errorf("contact_first_name and contact_last_name are required")
	}
	if country == "" || city == "" || commune == "" {
		return nil, fmt.Errorf("country, city and commune are required")
	}

	if _, err := s.users.GetByEmail(ctx, emailAddr); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if _, err := s.clubs.FindByName(ctx, clubName); err == nil {
		return nil, ErrClubNameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	pendingEmail, err := s.requests.HasPendingByEmail(ctx, emailAddr)
	if err != nil {
		return nil, err
	}
	if pendingEmail {
		return nil, ErrPendingClubRegistration
	}

	pendingName, err := s.requests.HasPendingByClubName(ctx, clubName)
	if err != nil {
		return nil, err
	}
	if pendingName {
		return nil, ErrPendingClubName
	}

	req := &domain.ClubRegistrationRequest{
		ClubName:         clubName,
		ContactEmail:     emailAddr,
		ContactFirstName: firstName,
		ContactLastName:  lastName,
		Phone:            trimOptional(input.Phone),
		Description:      trimOptional(input.Description),
		Country:          country,
		City:             city,
		Commune:          commune,
		FoundedAt:        input.FoundedAt,
		Message:          trimOptional(input.Message),
	}
	if err := s.requests.Create(ctx, req); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrPendingClubRegistration
		}
		return nil, err
	}
	return req, nil
}

func (s *ClubRegistrationService) List(ctx context.Context, status *domain.ClubRegistrationStatus) ([]domain.ClubRegistrationRequest, error) {
	return s.requests.List(ctx, status)
}

func (s *ClubRegistrationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClubRegistrationRequest, error) {
	return s.requests.GetByID(ctx, id)
}

func (s *ClubRegistrationService) Approve(ctx context.Context, reviewerID, requestID uuid.UUID, input ApproveClubRegistrationInput) (*ApproveClubRegistrationResult, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != domain.ClubRegistrationStatusPending {
		return nil, ErrClubRegistrationNotPending
	}

	if _, err := s.users.GetByEmail(ctx, req.ContactEmail); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if _, err := s.clubs.FindByName(ctx, req.ClubName); err == nil {
		return nil, ErrClubNameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	country := req.Country
	city := req.City
	commune := req.Commune
	if input.Country != nil && strings.TrimSpace(*input.Country) != "" {
		country = strings.TrimSpace(*input.Country)
	}
	if input.City != nil && strings.TrimSpace(*input.City) != "" {
		city = strings.TrimSpace(*input.City)
	}
	if input.Commune != nil && strings.TrimSpace(*input.Commune) != "" {
		commune = strings.TrimSpace(*input.Commune)
	}

	var approvedSlug *string
	if input.Slug != nil {
		slug := strings.TrimSpace(*input.Slug)
		if slug != "" {
			approvedSlug = &slug
		}
	}

	token, err := repository.GenerateAccessToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(s.inviteTTL)

	if err := s.requests.MarkApproved(ctx, requestID, repository.ApproveRegistrationParams{
		ReviewerID: reviewerID,
		Note:       input.ReviewNote,
		Token:      token,
		ExpiresAt:  expiresAt,
		Slug:       approvedSlug,
		Country:    country,
		City:       city,
		Commune:    commune,
	}); err != nil {
		return nil, err
	}

	registerURL := strings.TrimRight(s.publicURL, "/") + "/register-club?token=" + token
	contactName := strings.TrimSpace(req.ContactFirstName + " " + req.ContactLastName)
	if err := s.mailer.SendClubAccess(email.ClubAccessEmail{
		To:          req.ContactEmail,
		ClubName:    req.ClubName,
		ContactName: contactName,
		RegisterURL: registerURL,
		ExpiresAt:   expiresAt.Format("02/01/2006 15:04"),
	}); err != nil {
		return nil, fmt.Errorf("send club access email: %w", err)
	}

	updated, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	return &ApproveClubRegistrationResult{
		Request:   *updated,
		Message:   "approval email sent — contact must finish registration with Google",
		ExpiresAt: expiresAt,
	}, nil
}

func (s *ClubRegistrationService) PreviewAccess(ctx context.Context, token string) (*domain.ClubRegistrationAccessPreview, error) {
	req, err := s.loadValidAccessRequest(ctx, token)
	if err != nil {
		return nil, err
	}
	return &domain.ClubRegistrationAccessPreview{
		ClubName:         req.ClubName,
		ContactEmail:     req.ContactEmail,
		ContactFirstName: req.ContactFirstName,
		ContactLastName:  req.ContactLastName,
		ExpiresAt:        *req.AccessTokenExpiresAt,
	}, nil
}

func (s *ClubRegistrationService) Complete(ctx context.Context, input CompleteClubRegistrationInput) (*CompleteClubRegistrationResult, error) {
	if strings.TrimSpace(input.GoogleIDToken) != "" {
		return s.CompleteWithGoogle(ctx, input)
	}
	if strings.TrimSpace(input.Password) != "" {
		return s.CompleteWithPassword(ctx, input)
	}
	return nil, fmt.Errorf("google_id_token or password is required")
}

func (s *ClubRegistrationService) CompleteWithPassword(ctx context.Context, input CompleteClubRegistrationInput) (*CompleteClubRegistrationResult, error) {
	req, err := s.loadValidAccessRequest(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	if len(input.Password) < 8 {
		return nil, ErrPasswordRequired
	}

	if _, err := s.users.GetByEmail(ctx, req.ContactEmail); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	club, reviewerID, err := s.createClubFromApprovedRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	president, err := s.admin.CreateClubHeadPassword(ctx, reviewerID, club.ID, CreatePresidentInput{
		Email:     req.ContactEmail,
		Password:  input.Password,
		FirstName: req.ContactFirstName,
		LastName:  req.ContactLastName,
		Phone:     req.Phone,
	})
	if err != nil {
		_ = s.admin.DeleteClub(ctx, club.ID)
		return nil, err
	}

	return s.finalizeClubRegistration(ctx, req, club, president)
}

func (s *ClubRegistrationService) CompleteWithGoogle(ctx context.Context, input CompleteClubRegistrationInput) (*CompleteClubRegistrationResult, error) {
	if s.googleClient == "" {
		return nil, ErrGoogleAuthNotConfigured
	}

	req, err := s.loadValidAccessRequest(ctx, input.Token)
	if err != nil {
		return nil, err
	}

	identity, err := auth.VerifyGoogleIDToken(ctx, input.GoogleIDToken, s.googleClient)
	if err != nil {
		return nil, err
	}
	if identity.Email != req.ContactEmail {
		return nil, ErrGoogleEmailMismatch
	}

	if _, err := s.users.GetByEmail(ctx, req.ContactEmail); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if existing, err := s.users.GetByGoogleSub(ctx, identity.Sub); err == nil && existing != nil {
		return nil, ErrEmailAlreadyUsed
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	club, reviewerID, err := s.createClubFromApprovedRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	firstName := req.ContactFirstName
	lastName := req.ContactLastName
	if identity.GivenName != "" {
		firstName = identity.GivenName
	}
	if identity.FamilyName != "" {
		lastName = identity.FamilyName
	}

	president, err := s.admin.CreateClubHeadGoogle(ctx, reviewerID, club.ID, CreateGoogleHeadInput{
		Email:     req.ContactEmail,
		GoogleSub: identity.Sub,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     req.Phone,
	})
	if err != nil {
		_ = s.admin.DeleteClub(ctx, club.ID)
		return nil, err
	}

	return s.finalizeClubRegistration(ctx, req, club, president)
}

func (s *ClubRegistrationService) createClubFromApprovedRequest(ctx context.Context, req *domain.ClubRegistrationRequest) (*domain.Club, uuid.UUID, error) {
	if _, err := s.clubs.FindByName(ctx, req.ClubName); err == nil {
		return nil, uuid.Nil, ErrClubNameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, uuid.Nil, err
	}

	baseSlug := ""
	if req.ApprovedSlug != nil {
		baseSlug = strings.TrimSpace(*req.ApprovedSlug)
	}
	if baseSlug == "" {
		baseSlug = slugify("", req.ClubName)
	}
	suffix, err := randomAlphaNum(4)
	if err != nil {
		return nil, uuid.Nil, err
	}
	slug := strings.Trim(baseSlug, "-") + "-" + strings.ToLower(suffix)

	if req.ReviewedBy == nil {
		return nil, uuid.Nil, fmt.Errorf("registration is missing reviewer")
	}
	reviewerID := *req.ReviewedBy

	club, err := s.admin.CreateClub(ctx, reviewerID, CreateClubInput{
		Name:        req.ClubName,
		Slug:        slug,
		Description: req.Description,
		Country:     req.Country,
		City:        req.City,
		Commune:     req.Commune,
		FoundedAt:   req.FoundedAt,
	})
	if err != nil {
		return nil, uuid.Nil, err
	}
	return club, reviewerID, nil
}

func (s *ClubRegistrationService) finalizeClubRegistration(ctx context.Context, req *domain.ClubRegistrationRequest, club *domain.Club, president *domain.User) (*CompleteClubRegistrationResult, error) {
	if err := s.requests.MarkCompleted(ctx, req.ID, club.ID); err != nil {
		return nil, err
	}

	accessToken, expiresAt, err := s.tokens.Generate(president.ID, president.Email, president.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &CompleteClubRegistrationResult{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
		User:        *president,
		Club:        *club,
	}, nil
}

func (s *ClubRegistrationService) Reject(ctx context.Context, reviewerID, requestID uuid.UUID, note *string) error {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ClubRegistrationStatusPending {
		return ErrClubRegistrationNotPending
	}
	return s.requests.MarkRejected(ctx, requestID, reviewerID, note)
}

func (s *ClubRegistrationService) GoogleClientID() string {
	return s.googleClient
}

func (s *ClubRegistrationService) loadValidAccessRequest(ctx context.Context, token string) (*domain.ClubRegistrationRequest, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidClubAccessToken
	}

	req, err := s.requests.GetByAccessToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidClubAccessToken
		}
		return nil, err
	}
	if req.Status != domain.ClubRegistrationStatusApproved {
		return nil, ErrInvalidClubAccessToken
	}
	if req.AccessTokenUsedAt != nil || req.CreatedClubID != nil {
		return nil, ErrClubAccessAlreadyCompleted
	}
	if req.AccessTokenExpiresAt == nil || time.Now().UTC().After(*req.AccessTokenExpiresAt) {
		return nil, ErrInvalidClubAccessToken
	}
	return req, nil
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
