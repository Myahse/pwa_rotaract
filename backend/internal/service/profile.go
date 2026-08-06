package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

var (
	ErrInvalidAvatarType = ErrInvalidImageType
	ErrAvatarTooLarge    = ErrImageTooLarge
)

const defaultMaxAvatarBytes = 5 << 20 // 5 MB

type ProfileService struct {
	users         *repository.UserRepository
	store         *storage.LocalStore
	apiPublicURL  string
	maxAvatarSize int64
}

func NewProfileService(users *repository.UserRepository, store *storage.LocalStore, apiPublicURL string, maxAvatarSize int64) *ProfileService {
	if maxAvatarSize <= 0 {
		maxAvatarSize = defaultMaxAvatarBytes
	}
	return &ProfileService{
		users:         users,
		store:         store,
		apiPublicURL:  strings.TrimRight(apiPublicURL, "/"),
		maxAvatarSize: maxAvatarSize,
	}
}

type UpdateProfileInput struct {
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Phone       *string    `json:"phone"`
	BirthDate   *time.Time `json:"birth_date"`
	Profession  *string    `json:"profession"`
	MemberSince *time.Time `json:"member_since"`
}

func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.PublicUser(user), nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*domain.User, error) {
	if input.FirstName != nil {
		trimmed := strings.TrimSpace(*input.FirstName)
		if trimmed == "" {
			return nil, fmt.Errorf("first_name cannot be empty")
		}
		input.FirstName = &trimmed
	}
	if input.LastName != nil {
		trimmed := strings.TrimSpace(*input.LastName)
		if trimmed == "" {
			return nil, fmt.Errorf("last_name cannot be empty")
		}
		input.LastName = &trimmed
	}

	user, err := s.users.UpdateProfile(ctx, userID, repository.UpdateProfileParams{
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
	return s.PublicUser(user), nil
}

func (s *ProfileService) UploadAvatar(ctx context.Context, userID uuid.UUID, filename string, size int64, contentType string, file io.Reader) (*domain.User, error) {
	ext, err := imageExtension(filename, contentType)
	if err != nil {
		return nil, err
	}
	if size > s.maxAvatarSize {
		return nil, ErrAvatarTooLarge
	}

	limited := io.LimitReader(file, s.maxAvatarSize+1)
	relativePath, err := s.store.SaveAvatar(userID.String(), ext, limited)
	if err != nil {
		return nil, err
	}

	user, err := s.users.UpdateAvatarPath(ctx, userID, relativePath)
	if err != nil {
		_ = s.store.RemoveByRelativePath(relativePath)
		return nil, err
	}

	return s.PublicUser(user), nil
}

func (s *ProfileService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.AvatarPath != nil {
		if err := s.store.RemoveByRelativePath(*user.AvatarPath); err != nil {
			return nil, err
		}
	}

	updated, err := s.users.UpdateAvatarPath(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	return s.PublicUser(updated), nil
}

func (s *ProfileService) PublicUser(user *domain.User) *domain.User {
	if user == nil {
		return nil
	}
	user.PasswordHash = ""
	if user.AvatarPath != nil && *user.AvatarPath != "" {
		user.AvatarURL = s.uploadURL(user.AvatarPath)
	} else {
		user.AvatarURL = nil
	}
	return user
}

func (s *ProfileService) PublicClub(club *domain.Club) *domain.Club {
	if club == nil {
		return nil
	}
	club.LogoURL = s.uploadURL(club.LogoPath)
	return club
}

func (s *ProfileService) UploadURL(relativePath *string) *string {
	return s.uploadURL(relativePath)
}

func (s *ProfileService) uploadURL(relativePath *string) *string {
	if relativePath == nil || *relativePath == "" {
		return nil
	}
	url := s.apiPublicURL + "/api/v1/uploads/" + strings.TrimPrefix(*relativePath, "/")
	return &url
}
