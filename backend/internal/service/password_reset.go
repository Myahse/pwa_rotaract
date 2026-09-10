package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/email"
	"github.com/rotaract-civ/backend/internal/repository"
)

var ErrInvalidPasswordResetToken = errors.New("invalid or expired password reset token")

type PasswordResetService struct {
	users     *repository.UserRepository
	tokens    *repository.PasswordResetRepository
	mailer    *email.Client
	publicURL string
	ttl       time.Duration
}

func NewPasswordResetService(users *repository.UserRepository, tokens *repository.PasswordResetRepository, mailer *email.Client, publicURL string, ttl time.Duration) *PasswordResetService {
	return &PasswordResetService{users: users, tokens: tokens, mailer: mailer, publicURL: publicURL, ttl: ttl}
}

type RequestPasswordResetInput struct {
	Email string `json:"email"`
}
type CompletePasswordResetInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (s *PasswordResetService) Request(ctx context.Context, input RequestPasswordResetInput) error {
	emailAddress := strings.ToLower(strings.TrimSpace(input.Email))
	if emailAddress == "" {
		return nil
	}
	user, err := s.users.GetByEmail(ctx, emailAddress)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil || !user.IsActive {
		return err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	expiresAt := time.Now().UTC().Add(s.ttl)
	if err := s.tokens.Create(ctx, user.ID, hex.EncodeToString(hash[:]), expiresAt); err != nil {
		return err
	}
	return s.mailer.SendPasswordReset(email.PasswordResetEmail{
		To: user.Email, FirstName: user.FirstName,
		ResetURL:  strings.TrimRight(s.publicURL, "/") + "/reset-password?token=" + token,
		ExpiresAt: expiresAt.Format("02/01/2006 15:04"),
	})
}

func (s *PasswordResetService) Complete(ctx context.Context, input CompletePasswordResetInput) error {
	if len(input.Password) < 8 || strings.TrimSpace(input.Token) == "" {
		return ErrInvalidPasswordResetToken
	}
	hash := sha256.Sum256([]byte(strings.TrimSpace(input.Token)))
	token, err := s.tokens.GetValid(ctx, hex.EncodeToString(hash[:]))
	if err != nil {
		return ErrInvalidPasswordResetToken
	}
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePasswordHash(ctx, token.UserID, passwordHash); err != nil {
		return err
	}
	if err := s.tokens.MarkUsed(ctx, token.ID); err != nil {
		return err
	}
	return nil
}
