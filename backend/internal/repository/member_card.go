package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type MemberCardRepository struct {
	pool *pgxpool.Pool
}

func NewMemberCardRepository(pool *pgxpool.Pool) *MemberCardRepository {
	return &MemberCardRepository{pool: pool}
}

func (r *MemberCardRepository) Create(ctx context.Context, card *domain.MemberCard) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO member_cards (club_id, user_id, card_number)
		VALUES ($1, $2, $3)
		RETURNING id, issued_at
	`, card.ClubID, card.UserID, card.CardNumber).Scan(&card.ID, &card.IssuedAt)
	if err != nil {
		return fmt.Errorf("create member card: %w", err)
	}
	return nil
}

func (r *MemberCardRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MemberCard, error) {
	return r.scanOne(ctx, `
		SELECT mc.id, mc.club_id, mc.user_id, mc.card_number, mc.issued_at, mc.sent_at, mc.sent_by,
		       `+userSelectColumnsAliased+`,
		       c.id, c.name, c.slug, c.invite_code, c.description, c.country, c.city, c.commune,
		       c.logo_path, c.founded_at, c.is_active, c.created_by, c.created_at, c.updated_at
		FROM member_cards mc
		JOIN users u ON u.id = mc.user_id
		JOIN clubs c ON c.id = mc.club_id
		WHERE mc.id = $1
	`, id)
}

func (r *MemberCardRepository) GetByClubAndUser(ctx context.Context, clubID, userID uuid.UUID) (*domain.MemberCard, error) {
	return r.scanOne(ctx, `
		SELECT mc.id, mc.club_id, mc.user_id, mc.card_number, mc.issued_at, mc.sent_at, mc.sent_by,
		       `+userSelectColumnsAliased+`,
		       c.id, c.name, c.slug, c.invite_code, c.description, c.country, c.city, c.commune,
		       c.logo_path, c.founded_at, c.is_active, c.created_by, c.created_at, c.updated_at
		FROM member_cards mc
		JOIN users u ON u.id = mc.user_id
		JOIN clubs c ON c.id = mc.club_id
		WHERE mc.club_id = $1 AND mc.user_id = $2
	`, clubID, userID)
}

func (r *MemberCardRepository) MarkSent(ctx context.Context, id uuid.UUID, sentBy *uuid.UUID) (*domain.MemberCard, error) {
	card, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	err = r.pool.QueryRow(ctx, `
		UPDATE member_cards SET sent_at = now(), sent_by = $2
		WHERE id = $1
		RETURNING sent_at
	`, id, sentBy).Scan(&card.SentAt)
	if err != nil {
		return nil, fmt.Errorf("mark member card sent: %w", err)
	}
	card.SentBy = sentBy
	return card, nil
}

func (r *MemberCardRepository) NextSequence(ctx context.Context, clubID uuid.UUID) (int, error) {
	var next int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(
			CASE WHEN card_number ~ '-[0-9]+$'
			THEN CAST(substring(card_number from '[0-9]+$') AS INT)
			ELSE 0 END
		), 0) + 1
		FROM member_cards WHERE club_id = $1
	`, clubID).Scan(&next)
	return next, err
}

type MemberCardListFilter struct {
	ClubID      *uuid.UUID
	PendingOnly bool
	Limit       int
}

func (r *MemberCardRepository) List(ctx context.Context, filter MemberCardListFilter) ([]domain.MemberCard, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	query := `
		SELECT mc.id, mc.club_id, mc.user_id, mc.card_number, mc.issued_at, mc.sent_at, mc.sent_by,
		       ` + userSelectColumnsAliased + `,
		       c.id, c.name, c.slug, c.invite_code, c.description, c.country, c.city, c.commune,
		       c.logo_path, c.founded_at, c.is_active, c.created_by, c.created_at, c.updated_at
		FROM member_cards mc
		JOIN users u ON u.id = mc.user_id
		JOIN clubs c ON c.id = mc.club_id
		WHERE 1=1
	`
	args := make([]any, 0, 3)
	if filter.ClubID != nil {
		args = append(args, *filter.ClubID)
		query += fmt.Sprintf(" AND mc.club_id = $%d", len(args))
	}
	if filter.PendingOnly {
		query += " AND mc.sent_at IS NULL"
	}
	query += " ORDER BY mc.issued_at DESC LIMIT " + fmt.Sprintf("%d", limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list member cards: %w", err)
	}
	defer rows.Close()

	items := make([]domain.MemberCard, 0)
	for rows.Next() {
		card, err := scanMemberCardRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, card)
	}
	return items, rows.Err()
}

func (r *MemberCardRepository) ListMembershipsWithoutCard(ctx context.Context, clubID *uuid.UUID) ([]domain.ClubMembership, error) {
	query := `
		SELECT cm.id, cm.club_id, cm.user_id, cm.member_role, cm.joined_at,
		       ` + userSelectColumnsAliased + `
		FROM club_memberships cm
		JOIN users u ON u.id = cm.user_id
		LEFT JOIN member_cards mc ON mc.club_id = cm.club_id AND mc.user_id = cm.user_id
		WHERE mc.id IS NULL
	`
	args := make([]any, 0, 1)
	if clubID != nil {
		args = append(args, *clubID)
		query += " AND cm.club_id = $1"
	}
	query += " ORDER BY cm.joined_at ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list memberships without card: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ClubMembership, 0)
	for rows.Next() {
		var membership domain.ClubMembership
		var user domain.User
		if err := rows.Scan(
			&membership.ID, &membership.ClubID, &membership.UserID, &membership.MemberRole, &membership.JoinedAt,
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
			&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
			&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan membership without card: %w", err)
		}
		membership.User = &user
		items = append(items, membership)
	}
	return items, rows.Err()
}

func (r *MemberCardRepository) CountPending(ctx context.Context, clubID *uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM member_cards WHERE sent_at IS NULL`
	args := make([]any, 0, 1)
	if clubID != nil {
		args = append(args, *clubID)
		query += " AND club_id = $1"
	}
	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *MemberCardRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.MemberCard, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	card, err := scanMemberCardRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &card, nil
}

type memberCardScanner interface {
	Scan(dest ...any) error
}

func scanMemberCardRow(row memberCardScanner) (domain.MemberCard, error) {
	var card domain.MemberCard
	var user domain.User
	var club domain.Club
	if err := row.Scan(
		&card.ID, &card.ClubID, &card.UserID, &card.CardNumber, &card.IssuedAt, &card.SentAt, &card.SentBy,
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
		&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
		&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&club.ID, &club.Name, &club.Slug, &club.InviteCode, &club.Description, &club.Country, &club.City, &club.Commune,
		&club.LogoPath, &club.FoundedAt, &club.IsActive, &club.CreatedBy, &club.CreatedAt, &club.UpdatedAt,
	); err != nil {
		return domain.MemberCard{}, err
	}
	card.User = &user
	card.Club = &club
	return card, nil
}
