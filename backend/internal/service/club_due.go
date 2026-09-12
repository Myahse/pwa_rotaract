package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

type ClubDueService struct {
	dues   *repository.ClubDueRepository
	clubs  *repository.ClubRepository
	users  *repository.UserRepository
	store  *storage.LocalStore
	apiURL string
}

func NewClubDueService(
	dues *repository.ClubDueRepository,
	clubs *repository.ClubRepository,
	users *repository.UserRepository,
	store *storage.LocalStore,
	apiPublicURL string,
) *ClubDueService {
	return &ClubDueService{
		dues:   dues,
		clubs:  clubs,
		users:  users,
		store:  store,
		apiURL: strings.TrimRight(apiPublicURL, "/"),
	}
}

type ClubDueInput struct {
	DueMonth  string `json:"due_month"`
	AmountXOF int    `json:"amount_xof"`
}

func (s *ClubDueService) Submit(
	ctx context.Context,
	user *domain.User,
	clubID uuid.UUID,
	input ClubDueInput,
	filename string,
	size int64,
	contentType string,
	file io.Reader,
) (*domain.ClubDuePayment, error) {
	if _, err := s.clubs.GetMembership(ctx, clubID, user.ID); err != nil {
		return nil, err
	}
	dueMonth, err := parseDueMonth(input.DueMonth)
	if err != nil {
		return nil, err
	}
	if input.AmountXOF <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	if file == nil {
		return nil, fmt.Errorf("receipt file is required")
	}

	ext, mimeType, err := receiptExtension(filename, contentType)
	if err != nil {
		return nil, err
	}
	maxSize := maxDonationReceiptImage
	if mimeType == "application/pdf" {
		maxSize = maxDonationReceiptPDF
	}
	if size > maxSize {
		return nil, ErrReceiptTooLarge
	}

	due := &domain.ClubDuePayment{
		ClubID:    clubID,
		UserID:    user.ID,
		DueMonth:  dueMonth,
		AmountXOF: input.AmountXOF,
	}
	if err := s.dues.Create(ctx, due); err != nil {
		return nil, err
	}

	path, err := s.store.SaveClubDueReceipt(due.ID.String(), ext, io.LimitReader(file, maxSize+1))
	if err != nil {
		_ = s.dues.Delete(ctx, due.ID)
		return nil, err
	}
	if err := s.dues.SetReceipt(ctx, due.ID, path, mimeType); err != nil {
		_ = s.store.RemoveByRelativePath(path)
		_ = s.dues.Delete(ctx, due.ID)
		return nil, err
	}

	updated, err := s.dues.Get(ctx, due.ID)
	if err != nil {
		return nil, err
	}
	out := s.public(*updated)
	return &out, nil
}

func (s *ClubDueService) List(ctx context.Context, user *domain.User, clubID uuid.UUID) ([]domain.ClubDuePayment, error) {
	if _, err := s.clubs.GetMembership(ctx, clubID, user.ID); err != nil {
		return nil, err
	}
	canManage, err := s.canManageDues(ctx, user, clubID)
	if err != nil {
		return nil, err
	}

	var items []domain.ClubDuePayment
	if canManage {
		items, err = s.dues.ListByClub(ctx, clubID)
	} else {
		items, err = s.dues.ListByClubAndUser(ctx, clubID, user.ID)
	}
	if err != nil {
		return nil, err
	}
	out := make([]domain.ClubDuePayment, 0, len(items))
	for _, item := range items {
		public := s.public(item)
		if public.User != nil {
			public.User = s.publicUser(public.User)
		}
		out = append(out, public)
	}
	return out, nil
}

func (s *ClubDueService) MarkReceived(ctx context.Context, user *domain.User, clubID, dueID uuid.UUID) (*domain.ClubDuePayment, error) {
	canManage, err := s.canManageDues(ctx, user, clubID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, repository.ErrNotFound
	}
	item, err := s.dues.Get(ctx, dueID)
	if err != nil {
		return nil, err
	}
	if item.ClubID != clubID {
		return nil, repository.ErrNotFound
	}
	updated, err := s.dues.MarkReceived(ctx, dueID, user.ID)
	if err != nil {
		return nil, err
	}
	out := s.public(*updated)
	if out.User != nil {
		out.User = s.publicUser(out.User)
	}
	return &out, nil
}

func (s *ClubDueService) public(item domain.ClubDuePayment) domain.ClubDuePayment {
	if item.ReceiptPath != "" {
		url := s.apiURL + "/api/v1/uploads/" + strings.TrimPrefix(item.ReceiptPath, "/")
		item.ReceiptURL = &url
	}
	item.ReceiptPath = ""
	if item.User != nil {
		item.User.PasswordHash = ""
	}
	return item
}

func (s *ClubDueService) publicUser(user *domain.User) *domain.User {
	user.PasswordHash = ""
	return user
}

func (s *ClubDueService) canManageDues(ctx context.Context, user *domain.User, clubID uuid.UUID) (bool, error) {
	if user.IsAdmin {
		return true, nil
	}
	head, err := s.clubs.IsHead(ctx, clubID, user.ID)
	if err != nil {
		return false, err
	}
	if head {
		return true, nil
	}
	assignments, err := s.clubs.ListUserRoleAssignments(ctx, clubID, user.ID)
	if err != nil {
		return false, err
	}
	for _, item := range assignments {
		if item.Role != nil && isTreasurerRoleName(item.Role.Name) {
			return true, nil
		}
	}
	return false, nil
}

func isTreasurerRoleName(name string) bool {
	normalized := strings.Map(func(r rune) rune {
		switch r {
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		default:
			return unicode.ToLower(r)
		}
	}, strings.TrimSpace(name))
	return strings.Contains(normalized, "tresorier") || strings.Contains(normalized, "treasurer")
}

func parseDueMonth(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if len(value) == 7 {
		if _, err := time.Parse("2006-01", value); err != nil {
			return "", fmt.Errorf("due_month must be YYYY-MM")
		}
		return value + "-01", nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", fmt.Errorf("due_month must be YYYY-MM")
	}
	return value, nil
}
