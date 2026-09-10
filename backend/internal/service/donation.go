package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

const (
	maxDonationReceiptImage = int64(5 << 20)
	maxDonationReceiptPDF   = int64(10 << 20)
)

type DonationService struct {
	donations    *repository.DonationRepository
	users        *repository.UserRepository
	store        *storage.LocalStore
	apiPublicURL string
}

func NewDonationService(
	donations *repository.DonationRepository,
	users *repository.UserRepository,
	store *storage.LocalStore,
	apiPublicURL string,
) *DonationService {
	return &DonationService{
		donations:    donations,
		users:        users,
		store:        store,
		apiPublicURL: strings.TrimRight(apiPublicURL, "/"),
	}
}

type DonationInput struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AmountXOF int    `json:"amount_xof"`
}

func (s *DonationService) Submit(
	ctx context.Context,
	input DonationInput,
	filename string,
	size int64,
	contentType string,
	file io.Reader,
) (*domain.Donation, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if name == "" || email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("name and a valid email are required")
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

	donation := &domain.Donation{
		Name:      name,
		Email:     email,
		AmountXOF: input.AmountXOF,
	}
	if user, err := s.users.GetByEmail(ctx, email); err == nil {
		donation.UserID = &user.ID
	} else if err != repository.ErrNotFound {
		return nil, err
	}

	if err := s.donations.Create(ctx, donation); err != nil {
		return nil, err
	}

	path, err := s.store.SaveDonationReceipt(donation.ID.String(), ext, io.LimitReader(file, maxSize+1))
	if err != nil {
		_ = s.donations.Delete(ctx, donation.ID)
		return nil, err
	}

	if err := s.donations.SetReceipt(ctx, donation.ID, path, mimeType); err != nil {
		_ = s.store.RemoveByRelativePath(path)
		_ = s.donations.Delete(ctx, donation.ID)
		return nil, err
	}

	updated, err := s.donations.Get(ctx, donation.ID)
	if err != nil {
		return nil, err
	}
	out := s.public(*updated)
	return &out, nil
}

func (s *DonationService) List(ctx context.Context) ([]domain.Donation, error) {
	items, err := s.donations.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Donation, 0, len(items))
	for _, item := range items {
		out = append(out, s.public(item))
	}
	return out, nil
}

func (s *DonationService) MarkReceived(ctx context.Context, id uuid.UUID) (*domain.Donation, error) {
	item, err := s.donations.MarkReceived(ctx, id)
	if err != nil {
		return nil, err
	}
	out := s.public(*item)
	return &out, nil
}

func (s *DonationService) public(item domain.Donation) domain.Donation {
	if item.ReceiptPath != "" {
		url := s.apiPublicURL + "/api/v1/uploads/" + strings.TrimPrefix(item.ReceiptPath, "/")
		item.ReceiptURL = &url
	}
	item.ReceiptPath = ""
	return item
}
