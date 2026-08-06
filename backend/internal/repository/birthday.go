package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type PushSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewPushSubscriptionRepository(pool *pgxpool.Pool) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{pool: pool}
}

func (r *PushSubscriptionRepository) Upsert(ctx context.Context, sub *domain.PushSubscription) error {
	query := `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (endpoint) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			p256dh = EXCLUDED.p256dh,
			auth = EXCLUDED.auth,
			user_agent = EXCLUDED.user_agent,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth, sub.UserAgent).
		Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2
	`, userID, endpoint)
	return err
}

func (r *PushSubscriptionRepository) DeleteByEndpointGlobal(ctx context.Context, endpoint string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}

func (r *PushSubscriptionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.PushSubscription, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, endpoint, p256dh, auth, user_agent, created_at, updated_at
		FROM push_subscriptions WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := make([]domain.PushSubscription, 0)
	for rows.Next() {
		var sub domain.PushSubscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.P256dh, &sub.Auth, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

type BirthdayRepository struct {
	pool *pgxpool.Pool
}

func NewBirthdayRepository(pool *pgxpool.Pool) *BirthdayRepository {
	return &BirthdayRepository{pool: pool}
}

type BirthdayUser struct {
	User   domain.User
	ClubID *uuid.UUID
	ClubName *string
}

func (r *BirthdayRepository) UsersWithBirthdayOn(ctx context.Context, month, day int) ([]BirthdayUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+userSelectColumnsAliased+`, c.id, c.name
		FROM users u
		LEFT JOIN LATERAL (
			SELECT cm.club_id FROM club_memberships cm WHERE cm.user_id = u.id ORDER BY cm.joined_at ASC LIMIT 1
		) primary_club ON TRUE
		LEFT JOIN clubs c ON c.id = primary_club.club_id
		WHERE u.is_active = TRUE
		  AND u.birth_date IS NOT NULL
		  AND EXTRACT(MONTH FROM u.birth_date) = $1
		  AND EXTRACT(DAY FROM u.birth_date) = $2
	`, month, day)
	if err != nil {
		return nil, fmt.Errorf("list birthday users: %w", err)
	}
	defer rows.Close()

	users := make([]BirthdayUser, 0)
	for rows.Next() {
		var item BirthdayUser
		if err := rows.Scan(
			&item.User.ID, &item.User.Email, &item.User.PasswordHash, &item.User.FirstName, &item.User.LastName, &item.User.Phone,
			&item.User.BirthDate, &item.User.Profession, &item.User.MemberSince, &item.User.AvatarPath,
			&item.User.IsAdmin, &item.User.IsActive, &item.User.CreatedAt, &item.User.UpdatedAt,
			&item.ClubID, &item.ClubName,
		); err != nil {
			return nil, fmt.Errorf("scan birthday user: %w", err)
		}
		users = append(users, item)
	}
	return users, rows.Err()
}

func (r *BirthdayRepository) ClubBirthdaysOn(ctx context.Context, clubID uuid.UUID, month, day int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+userSelectColumnsAliased+`
		FROM users u
		JOIN club_memberships cm ON cm.user_id = u.id
		WHERE cm.club_id = $1
		  AND u.is_active = TRUE
		  AND u.birth_date IS NOT NULL
		  AND EXTRACT(MONTH FROM u.birth_date) = $2
		  AND EXTRACT(DAY FROM u.birth_date) = $3
		ORDER BY u.first_name ASC
	`, clubID, month, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
			&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
			&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan club birthday user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *BirthdayRepository) ClubmatesBirthdaysOn(ctx context.Context, userID uuid.UUID, month, day int) ([]BirthdayUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+userSelectColumnsAliased+`, c.id, c.name
		FROM users u
		JOIN club_memberships cm ON cm.user_id = u.id
		JOIN clubs c ON c.id = cm.club_id
		WHERE u.is_active = TRUE
		  AND u.id != $1
		  AND u.birth_date IS NOT NULL
		  AND EXTRACT(MONTH FROM u.birth_date) = $2
		  AND EXTRACT(DAY FROM u.birth_date) = $3
		  AND cm.club_id IN (
			SELECT club_id FROM club_memberships WHERE user_id = $1
		  )
		ORDER BY u.first_name ASC
	`, userID, month, day)
	if err != nil {
		return nil, fmt.Errorf("list clubmate birthdays: %w", err)
	}
	defer rows.Close()

	users := make([]BirthdayUser, 0)
	seen := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var item BirthdayUser
		if err := rows.Scan(
			&item.User.ID, &item.User.Email, &item.User.PasswordHash, &item.User.FirstName, &item.User.LastName, &item.User.Phone,
			&item.User.BirthDate, &item.User.Profession, &item.User.MemberSince, &item.User.AvatarPath,
			&item.User.IsAdmin, &item.User.IsActive, &item.User.CreatedAt, &item.User.UpdatedAt,
			&item.ClubID, &item.ClubName,
		); err != nil {
			return nil, fmt.Errorf("scan clubmate birthday: %w", err)
		}
		if _, ok := seen[item.User.ID]; ok {
			continue
		}
		seen[item.User.ID] = struct{}{}
		users = append(users, item)
	}
	return users, rows.Err()
}

func (r *BirthdayRepository) WasNotified(ctx context.Context, userID uuid.UUID, year int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM birthday_notifications WHERE user_id = $1 AND year = $2
		)
	`, userID, year).Scan(&exists)
	return exists, err
}

func (r *BirthdayRepository) MarkNotified(ctx context.Context, userID uuid.UUID, year int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO birthday_notifications (user_id, year)
		VALUES ($1, $2)
		ON CONFLICT (user_id, year) DO NOTHING
	`, userID, year)
	return err
}

func (r *BirthdayRepository) UserHasBirthdayOn(ctx context.Context, userID uuid.UUID, month, day int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE id = $1 AND birth_date IS NOT NULL
			  AND EXTRACT(MONTH FROM birth_date) = $2
			  AND EXTRACT(DAY FROM birth_date) = $3
		)
	`, userID, month, day).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, err
}

func TodayInLocation(loc *time.Location) (month int, day int, year int) {
	now := time.Now().In(loc)
	return int(now.Month()), now.Day(), now.Year()
}
