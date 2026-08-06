package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rotaract-civ/backend/internal/domain"
)

func (r *ChatRepository) CreateMessage(ctx context.Context, msg *domain.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (group_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, msg.GroupID, msg.UserID, msg.Content).
		Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("create chat message: %w", err)
	}
	return nil
}

func (r *ChatRepository) ListMessages(ctx context.Context, groupID uuid.UUID, limit int, before *time.Time) ([]domain.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT m.id, m.group_id, m.user_id, m.content, m.created_at,
		       ` + userSelectColumnsAliased + `
		FROM chat_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.group_id = $1
	`
	args := []any{groupID}
	if before != nil {
		query += ` AND m.created_at < $2`
		args = append(args, *before)
	}
	query += ` ORDER BY m.created_at DESC LIMIT ` + fmt.Sprintf("%d", limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.ChatMessage, 0)
	for rows.Next() {
		var msg domain.ChatMessage
		var user domain.User
		if err := rows.Scan(
			&msg.ID, &msg.GroupID, &msg.UserID, &msg.Content, &msg.CreatedAt,
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
			&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
			&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		user.PasswordHash = ""
		msg.User = &user
		messages = append(messages, msg)
	}

	// Return chronological order (oldest first) for UI.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

func (r *ChatRepository) GetMessage(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error) {
	var msg domain.ChatMessage
	var user domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT m.id, m.group_id, m.user_id, m.content, m.created_at,
		       `+userSelectColumnsAliased+`
		FROM chat_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1
	`, id).Scan(
		&msg.ID, &msg.GroupID, &msg.UserID, &msg.Content, &msg.CreatedAt,
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
		&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
		&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chat message: %w", err)
	}
	user.PasswordHash = ""
	msg.User = &user
	return &msg, nil
}

func (r *ChatRepository) UpdateMessageContent(ctx context.Context, id uuid.UUID, content string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE chat_messages SET content = $2 WHERE id = $1`, id, content)
	if err != nil {
		return fmt.Errorf("update chat message: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ChatRepository) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM chat_messages WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete chat message: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
