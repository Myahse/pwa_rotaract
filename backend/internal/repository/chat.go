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

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) CreateGroup(ctx context.Context, group *domain.ChatGroup) error {
	query := `
		INSERT INTO chat_groups (club_id, commission_id, group_type, name, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, group.ClubID, group.CommissionID, group.GroupType, group.Name, group.CreatedBy).
		Scan(&group.ID, &group.CreatedAt)
	if err != nil {
		return fmt.Errorf("create chat group: %w", err)
	}
	return nil
}

func (r *ChatRepository) AddMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_group_members (group_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("add chat member: %w", err)
	}
	return nil
}

func (r *ChatRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]domain.ChatGroup, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, club_id, commission_id, group_type, name, created_by, created_at
		FROM chat_groups WHERE club_id = $1 ORDER BY created_at ASC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list chat groups: %w", err)
	}
	defer rows.Close()

	groups := make([]domain.ChatGroup, 0)
	for rows.Next() {
		var group domain.ChatGroup
		if err := rows.Scan(&group.ID, &group.ClubID, &group.CommissionID, &group.GroupType, &group.Name, &group.CreatedBy, &group.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat group: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (r *ChatRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChatGroup, error) {
	var group domain.ChatGroup
	err := r.pool.QueryRow(ctx, `
		SELECT id, club_id, commission_id, group_type, name, created_by, created_at
		FROM chat_groups WHERE id = $1
	`, id).Scan(&group.ID, &group.ClubID, &group.CommissionID, &group.GroupType, &group.Name, &group.CreatedBy, &group.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chat group: %w", err)
	}
	return &group, nil
}

func (r *ChatRepository) GetCommissionGroup(ctx context.Context, commissionID uuid.UUID) (*domain.ChatGroup, error) {
	var group domain.ChatGroup
	err := r.pool.QueryRow(ctx, `
		SELECT id, club_id, commission_id, group_type, name, created_by, created_at
		FROM chat_groups WHERE commission_id = $1 AND group_type = 'commission'
	`, commissionID).Scan(&group.ID, &group.ClubID, &group.CommissionID, &group.GroupType, &group.Name, &group.CreatedBy, &group.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get commission chat group: %w", err)
	}
	return &group, nil
}

func (r *ChatRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.ChatGroupMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cgm.id, cgm.group_id, cgm.user_id, cgm.joined_at,
		       `+userSelectColumnsAliased+`
		FROM chat_group_members cgm
		JOIN users u ON u.id = cgm.user_id
		WHERE cgm.group_id = $1
		ORDER BY cgm.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list chat members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.ChatGroupMember, 0)
	for rows.Next() {
		var member domain.ChatGroupMember
		var user domain.User
		if err := rows.Scan(
			&member.ID, &member.GroupID, &member.UserID, &member.JoinedAt,
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
			&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
			&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat member: %w", err)
		}
		member.User = &user
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *ChatRepository) UserCanAccessGroup(ctx context.Context, groupID, userID uuid.UUID, isAdmin bool) (bool, error) {
	if isAdmin {
		return true, nil
	}

	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM chat_group_members WHERE group_id = $1 AND user_id = $2
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}

func (r *ChatRepository) CountMembers(ctx context.Context, groupID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM chat_group_members WHERE group_id = $1
	`, groupID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count chat members: %w", err)
	}
	return count, nil
}

func (r *ChatRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM chat_group_members WHERE group_id = $1 AND user_id = $2
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("remove chat member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ChatRepository) FindDirectGroup(ctx context.Context, clubID, userA, userB uuid.UUID) (*domain.ChatGroup, error) {
	var group domain.ChatGroup
	err := r.pool.QueryRow(ctx, `
		SELECT g.id, g.club_id, g.commission_id, g.group_type, g.name, g.created_by, g.created_at
		FROM chat_groups g
		WHERE g.club_id = $1
		  AND g.group_type = 'custom'
		  AND EXISTS (SELECT 1 FROM chat_group_members WHERE group_id = g.id AND user_id = $2)
		  AND EXISTS (SELECT 1 FROM chat_group_members WHERE group_id = g.id AND user_id = $3)
		  AND (SELECT COUNT(*) FROM chat_group_members WHERE group_id = g.id) = 2
		LIMIT 1
	`, clubID, userA, userB).Scan(
		&group.ID, &group.ClubID, &group.CommissionID, &group.GroupType, &group.Name, &group.CreatedBy, &group.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find direct chat group: %w", err)
	}
	return &group, nil
}

func (r *ChatRepository) ListInboxForUser(ctx context.Context, clubID, userID uuid.UUID) ([]domain.ChatInboxGroup, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.club_id, g.commission_id, g.group_type, g.name, g.created_by, g.created_at,
		       lm.content, lm.created_at,
		       (SELECT COUNT(*)::int FROM chat_group_members WHERE group_id = g.id) AS member_count
		FROM chat_groups g
		LEFT JOIN LATERAL (
			SELECT content, created_at
			FROM chat_messages
			WHERE group_id = g.id
			ORDER BY created_at DESC
			LIMIT 1
		) lm ON true
		WHERE g.club_id = $1
		  AND (
		    g.group_type = 'club'
		    OR EXISTS (
		      SELECT 1 FROM chat_group_members cgm
		      WHERE cgm.group_id = g.id AND cgm.user_id = $2
		    )
		  )
		ORDER BY COALESCE(lm.created_at, g.created_at) DESC
	`, clubID, userID)
	if err != nil {
		return nil, fmt.Errorf("list chat inbox: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ChatInboxGroup, 0)
	for rows.Next() {
		var item domain.ChatInboxGroup
		var lastContent *string
		var lastAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.ClubID, &item.CommissionID, &item.GroupType, &item.Name, &item.CreatedBy, &item.CreatedAt,
			&lastContent, &lastAt,
			&item.MemberCount,
		); err != nil {
			return nil, fmt.Errorf("scan chat inbox row: %w", err)
		}
		item.LastMessageContent = lastContent
		item.LastMessageAt = lastAt
		item.IsDirect = item.GroupType == domain.ChatGroupTypeCustom && item.MemberCount == 2
		items = append(items, item)
	}
	return items, rows.Err()
}
