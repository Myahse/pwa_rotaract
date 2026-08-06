package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/domain"
)

type UploadedImage struct {
	Filename    string
	Size        int64
	ContentType string
	Reader      io.Reader
}

type CreatePresidentInput struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
	Phone       *string
	BirthDate   *time.Time
	Profession  *string
	MemberSince *time.Time
	Avatar      *UploadedImage
}

type CreateClubFullInput struct {
	Name        string
	Slug        string
	Description *string
	Country     string
	City        string
	Commune     string
	FoundedAt   *time.Time
	Logo        *UploadedImage
	President   CreatePresidentInput
}

type CreateClubResult struct {
	Club      domain.Club `json:"club"`
	President domain.User `json:"president"`
}

func (s *AdminService) CreateClubFull(ctx context.Context, adminID uuid.UUID, input CreateClubFullInput) (*CreateClubResult, error) {
	if input.Logo == nil {
		return nil, fmt.Errorf("club logo is required")
	}
	if strings.TrimSpace(input.President.Email) == "" {
		return nil, fmt.Errorf("president email is required")
	}
	if strings.TrimSpace(input.President.Password) == "" {
		return nil, fmt.Errorf("president password is required")
	}
	if strings.TrimSpace(input.President.FirstName) == "" || strings.TrimSpace(input.President.LastName) == "" {
		return nil, fmt.Errorf("president first and last name are required")
	}

	club, err := s.createClubRecord(ctx, adminID, CreateClubInput{
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		Country:     input.Country,
		City:        input.City,
		Commune:     input.Commune,
		FoundedAt:   input.FoundedAt,
	})
	if err != nil {
		return nil, err
	}

	logoPath, err := s.saveClubLogo(club.ID, input.Logo)
	if err != nil {
		return nil, err
	}
	club, err = s.clubs.UpdateLogoPath(ctx, club.ID, logoPath)
	if err != nil {
		_ = s.store.RemoveByRelativePath(logoPath)
		return nil, err
	}

	president, err := s.createPresidentUser(ctx, adminID, club.ID, input.President)
	if err != nil {
		if club.LogoPath != nil {
			_ = s.store.RemoveByRelativePath(*club.LogoPath)
		}
		return nil, err
	}

	return &CreateClubResult{
		Club:      *s.profiles.PublicClub(club),
		President: *s.profiles.PublicUser(president),
	}, nil
}

func (s *AdminService) createClubRecord(ctx context.Context, adminID uuid.UUID, input CreateClubInput) (*domain.Club, error) {
	country := strings.TrimSpace(input.Country)
	city := strings.TrimSpace(input.City)
	commune := strings.TrimSpace(input.Commune)
	if country == "" {
		return nil, fmt.Errorf("country is required")
	}
	if city == "" {
		return nil, fmt.Errorf("city is required")
	}
	if commune == "" {
		return nil, fmt.Errorf("commune is required")
	}

	club := &domain.Club{
		Name:        strings.TrimSpace(input.Name),
		Slug:        slugify(input.Slug, input.Name),
		Description: input.Description,
		Country:     &country,
		City:        &city,
		Commune:     &commune,
		FoundedAt:   input.FoundedAt,
		CreatedBy:   &adminID,
	}
	if club.Name == "" {
		return nil, fmt.Errorf("club name is required")
	}

	inviteCode, err := GenerateInviteCode(club.Slug)
	if err != nil {
		return nil, err
	}
	club.InviteCode = inviteCode

	if err := s.clubs.Create(ctx, club); err != nil {
		return nil, err
	}

	if _, err := s.clubs.SeedHeadRole(ctx, club.ID); err != nil {
		return nil, err
	}
	if _, err := s.clubs.SeedMemberRole(ctx, club.ID); err != nil {
		return nil, err
	}

	group := &domain.ChatGroup{
		ClubID:    club.ID,
		GroupType: domain.ChatGroupTypeClub,
		Name:      club.Name + " - Groupe général",
		CreatedBy: &adminID,
	}
	if err := s.chat.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	if s.commissions != nil {
		if err := s.seedDefaultCommissions(ctx, adminID, club.ID); err != nil {
			return nil, err
		}
	}

	return club, nil
}

func (s *AdminService) saveClubLogo(clubID uuid.UUID, logo *UploadedImage) (string, error) {
	ext, err := imageExtension(logo.Filename, logo.ContentType)
	if err != nil {
		return "", err
	}
	if logo.Size > s.maxImageSize {
		return "", ErrImageTooLarge
	}
	limited := io.LimitReader(logo.Reader, s.maxImageSize+1)
	return s.store.SaveClubLogo(clubID.String(), ext, limited)
}

type CreateGoogleHeadInput struct {
	Email     string
	GoogleSub string
	FirstName string
	LastName  string
	Phone     *string
}

func (s *AdminService) CreateClubHeadGoogle(ctx context.Context, adminID, clubID uuid.UUID, input CreateGoogleHeadInput) (*domain.User, error) {
	if _, err := s.clubs.GetByID(ctx, clubID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.GoogleSub) == "" {
		return nil, fmt.Errorf("google_sub is required")
	}
	if strings.TrimSpace(input.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if strings.TrimSpace(input.FirstName) == "" || strings.TrimSpace(input.LastName) == "" {
		return nil, fmt.Errorf("first and last name are required")
	}

	sub := strings.TrimSpace(input.GoogleSub)
	user := &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: auth.OAuthPasswordMarker,
		GoogleSub:    &sub,
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Phone:        input.Phone,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.assignClubHead(ctx, adminID, clubID, user.ID); err != nil {
		return nil, err
	}

	if s.diary != nil {
		parentID := s.diary.FindParrainParentID(ctx, clubID)
		now := time.Now()
		_ = s.diary.RecordPresidentFromUser(ctx, adminID, clubID, user, parentID, &now)
	}
	if s.mandates != nil {
		_ = s.mandates.RecordPresidentInCurrentMandate(ctx, clubID, user)
	}

	user.PasswordHash = ""
	return s.profiles.PublicUser(user), nil
}

func (s *AdminService) createPresidentUser(ctx context.Context, adminID, clubID uuid.UUID, input CreatePresidentInput) (*domain.User, error) {
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
		return nil, err
	}

	if input.Avatar != nil {
		ext, err := imageExtension(input.Avatar.Filename, input.Avatar.ContentType)
		if err != nil {
			return nil, err
		}
		if input.Avatar.Size > s.maxImageSize {
			return nil, ErrImageTooLarge
		}
		limited := io.LimitReader(input.Avatar.Reader, s.maxImageSize+1)
		path, err := s.store.SaveAvatar(user.ID.String(), ext, limited)
		if err != nil {
			return nil, err
		}
		user, err = s.users.UpdateAvatarPath(ctx, user.ID, path)
		if err != nil {
			_ = s.store.RemoveByRelativePath(path)
			return nil, err
		}
	}

	if err := s.assignClubHead(ctx, adminID, clubID, user.ID); err != nil {
		return nil, err
	}

	if s.diary != nil {
		parentID := s.diary.FindParrainParentID(ctx, clubID)
		now := time.Now()
		_ = s.diary.RecordPresidentFromUser(ctx, adminID, clubID, user, parentID, &now)
	}
	if s.mandates != nil {
		_ = s.mandates.RecordPresidentInCurrentMandate(ctx, clubID, user)
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AdminService) assignClubHead(ctx context.Context, adminID, clubID, userID uuid.UUID) error {
	members, err := s.clubs.ListMembers(ctx, clubID)
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.MemberRole == domain.ClubMemberRoleHead {
			return ErrHeadAlreadyExists
		}
	}

	membership := &domain.ClubMembership{
		ClubID:     clubID,
		UserID:     userID,
		MemberRole: domain.ClubMemberRoleHead,
	}
	if err := s.clubs.AddMembership(ctx, membership); err != nil {
		return err
	}

	headRole, err := s.findRoleByName(ctx, clubID, "Responsable de club")
	if err != nil {
		return err
	}
	if err := s.clubs.AssignRole(ctx, &domain.ClubMemberRoleAssignment{
		ClubID:     clubID,
		UserID:     userID,
		ClubRoleID: headRole.ID,
		AssignedBy: &adminID,
	}); err != nil {
		return err
	}

	groups, err := s.chat.ListByClub(ctx, clubID)
	if err != nil {
		return err
	}
	for _, group := range groups {
		if group.GroupType == domain.ChatGroupTypeClub {
			if err := s.chat.AddMember(ctx, group.ID, userID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *AdminService) PublicClubs(ctx context.Context) ([]domain.Club, error) {
	clubs, err := s.clubs.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range clubs {
		s.profiles.PublicClub(&clubs[i])
	}
	return clubs, nil
}
