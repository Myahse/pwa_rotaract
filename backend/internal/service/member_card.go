package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/email"
	"github.com/rotaract-civ/backend/internal/membercard"
	"github.com/rotaract-civ/backend/internal/repository"
)

type MemberCardService struct {
	cards  *repository.MemberCardRepository
	clubs  *repository.ClubRepository
	users  *repository.UserRepository
	mailer *email.Client
	logger *slog.Logger
}

func NewMemberCardService(
	cards *repository.MemberCardRepository,
	clubs *repository.ClubRepository,
	users *repository.UserRepository,
	mailer *email.Client,
	logger *slog.Logger,
) *MemberCardService {
	if logger == nil {
		logger = slog.Default()
	}
	return &MemberCardService{
		cards:  cards,
		clubs:  clubs,
		users:  users,
		mailer: mailer,
		logger: logger,
	}
}

type MemberCardListFilter struct {
	ClubID      *uuid.UUID
	PendingOnly bool
	Limit       int
}

type MemberCardSendResult struct {
	Sent   int      `json:"sent"`
	Failed int      `json:"failed"`
	Errors []string `json:"errors,omitempty"`
}

func (s *MemberCardService) List(ctx context.Context, filter MemberCardListFilter) ([]domain.MemberCard, error) {
	items, err := s.cards.List(ctx, repository.MemberCardListFilter{
		ClubID:      filter.ClubID,
		PendingOnly: filter.PendingOnly,
		Limit:       filter.Limit,
	})
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i] = s.public(items[i])
	}
	return items, nil
}

func (s *MemberCardService) CountPending(ctx context.Context, clubID *uuid.UUID) (int, error) {
	return s.cards.CountPending(ctx, clubID)
}

func (s *MemberCardService) IssueAndSend(ctx context.Context, clubID, userID uuid.UUID, sentBy *uuid.UUID) error {
	card, err := s.ensureCard(ctx, clubID, userID)
	if err != nil {
		return err
	}
	return s.sendCard(ctx, card, sentBy)
}

func (s *MemberCardService) IssueMissing(ctx context.Context, clubID *uuid.UUID) (int, error) {
	memberships, err := s.cards.ListMembershipsWithoutCard(ctx, clubID)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, membership := range memberships {
		if _, err := s.ensureCard(ctx, membership.ClubID, membership.UserID); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func (s *MemberCardService) SendPending(ctx context.Context, clubID *uuid.UUID, sentBy uuid.UUID) (*MemberCardSendResult, error) {
	items, err := s.cards.List(ctx, repository.MemberCardListFilter{
		ClubID:      clubID,
		PendingOnly: true,
		Limit:       500,
	})
	if err != nil {
		return nil, err
	}
	result := &MemberCardSendResult{}
	for _, item := range items {
		if err := s.sendCard(ctx, &item, &sentBy); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", item.User.Email, err.Error()))
			continue
		}
		result.Sent++
	}
	return result, nil
}

func (s *MemberCardService) SendOne(ctx context.Context, cardID, sentBy uuid.UUID) (*domain.MemberCard, error) {
	card, err := s.cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := s.sendCard(ctx, card, &sentBy); err != nil {
		return nil, err
	}
	updated, err := s.cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	out := s.public(*updated)
	return &out, nil
}

func (s *MemberCardService) ensureCard(ctx context.Context, clubID, userID uuid.UUID) (*domain.MemberCard, error) {
	existing, err := s.cards.GetByClubAndUser(ctx, clubID, userID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if _, err := s.clubs.GetMembership(ctx, clubID, userID); err != nil {
		return nil, err
	}
	club, err := s.clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	seq, err := s.cards.NextSequence(ctx, clubID)
	if err != nil {
		return nil, err
	}
	prefix := strings.ToUpper(strings.ReplaceAll(club.InviteCode, "-", ""))
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	if prefix == "" {
		prefix = strings.ToUpper(strings.ReplaceAll(club.Slug, "-", ""))
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
	}

	card := &domain.MemberCard{
		ClubID:     clubID,
		UserID:     userID,
		CardNumber: fmt.Sprintf("RCI-%s-%04d", prefix, seq),
	}
	if err := s.cards.Create(ctx, card); err != nil {
		return nil, err
	}
	return s.cards.GetByClubAndUser(ctx, clubID, userID)
}

func (s *MemberCardService) sendCard(ctx context.Context, card *domain.MemberCard, sentBy *uuid.UUID) error {
	if card.User == nil || card.Club == nil {
		loaded, err := s.cards.GetByID(ctx, card.ID)
		if err != nil {
			return err
		}
		card = loaded
	}
	if card.User == nil || card.Club == nil {
		return fmt.Errorf("member card data incomplete")
	}

	memberSince := ""
	if card.User.MemberSince != nil {
		memberSince = card.User.MemberSince.Format("01/2006")
	} else {
		memberSince = card.IssuedAt.Format("01/2006")
	}

	pdfBytes, err := membercard.RenderPDF(membercard.CardData{
		FullName:    card.User.FullName(),
		ClubName:    card.Club.Name,
		CardNumber:  card.CardNumber,
		MemberSince: memberSince,
		IssuedAt:    card.IssuedAt,
	})
	if err != nil {
		return err
	}

	if err := s.mailer.SendMemberCard(email.MemberCardEmail{
		To:         card.User.Email,
		FirstName:  card.User.FirstName,
		ClubName:   card.Club.Name,
		CardNumber: card.CardNumber,
		PDF:        pdfBytes,
	}); err != nil {
		return err
	}

	if _, err := s.cards.MarkSent(ctx, card.ID, sentBy); err != nil {
		return err
	}
	return nil
}

func (s *MemberCardService) public(card domain.MemberCard) domain.MemberCard {
	if card.User != nil {
		card.User.PasswordHash = ""
	}
	return card
}
