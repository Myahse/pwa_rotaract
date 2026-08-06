package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rotaract-civ/backend/internal/domain"
)

func (r *SocialRepository) CreateGroup(ctx context.Context, g *domain.SocialGroup) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO social_groups (name, description, privacy, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, g.Name, g.Description, g.Privacy, g.CreatedBy).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
}

func (r *SocialRepository) AddGroupMember(ctx context.Context, groupID, userID uuid.UUID, role domain.SocialGroupMemberRole) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO social_group_members (group_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (group_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, groupID, userID, role)
	return err
}

func (r *SocialRepository) RemoveGroupMember(ctx context.Context, groupID, userID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `
		DELETE FROM social_group_members WHERE group_id = $1 AND user_id = $2
	`, groupID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SocialRepository) GetGroupMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.SocialGroupMemberRole, error) {
	var role domain.SocialGroupMemberRole
	err := r.pool.QueryRow(ctx, `
		SELECT role FROM social_group_members WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (r *SocialRepository) IsGroupMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM social_group_members WHERE group_id = $1 AND user_id = $2)
	`, groupID, userID).Scan(&ok)
	return ok, err
}

func (r *SocialRepository) GetGroup(ctx context.Context, groupID, viewerID uuid.UUID) (*domain.SocialGroup, error) {
	var g domain.SocialGroup
	var creator domain.SocialAuthor
	var avatar *string
	var myRole *string
	var joinStatus string
	err := r.pool.QueryRow(ctx, `
		SELECT g.id, g.name, g.description, g.privacy, g.created_by, g.created_at, g.updated_at,
		       (SELECT COUNT(*)::int FROM social_group_members m WHERE m.group_id = g.id),
		       u.first_name, u.last_name, u.avatar_path,
		       mem.role,
		       CASE
		         WHEN mem.user_id IS NOT NULL THEN 'member'
		         WHEN jr.status = 'pending' THEN 'pending'
		         WHEN jr.status = 'rejected' THEN 'rejected'
		         ELSE 'none'
		       END,
		       CASE WHEN mem.role = 'admin' THEN (
		         SELECT COUNT(*)::int FROM social_group_join_requests r
		         WHERE r.group_id = g.id AND r.status = 'pending'
		       ) ELSE 0 END
		FROM social_groups g
		JOIN users u ON u.id = g.created_by
		LEFT JOIN social_group_members mem ON mem.group_id = g.id AND mem.user_id = $2
		LEFT JOIN social_group_join_requests jr ON jr.group_id = g.id AND jr.user_id = $2
		WHERE g.id = $1
	`, groupID, viewerID).Scan(
		&g.ID, &g.Name, &g.Description, &g.Privacy, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt,
		&g.MemberCount,
		&creator.FirstName, &creator.LastName, &avatar,
		&myRole, &joinStatus, &g.PendingCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	creator.ID = g.CreatedBy
	creator.AvatarURL = avatar
	g.Creator = &creator
	g.JoinStatus = domain.SocialGroupJoinStatus(joinStatus)
	if myRole != nil {
		role := domain.SocialGroupMemberRole(*myRole)
		g.MyRole = &role
	}
	return &g, nil
}

func (r *SocialRepository) ListGroups(ctx context.Context, viewerID uuid.UUID, mineOnly bool, limit int) ([]domain.SocialGroup, error) {
	if limit <= 0 || limit > 50 {
		limit = 30
	}
	query := `
		SELECT g.id, g.name, g.description, g.privacy, g.created_by, g.created_at, g.updated_at,
		       (SELECT COUNT(*)::int FROM social_group_members m WHERE m.group_id = g.id),
		       u.first_name, u.last_name, u.avatar_path,
		       mem.role,
		       CASE
		         WHEN mem.user_id IS NOT NULL THEN 'member'
		         WHEN jr.status = 'pending' THEN 'pending'
		         WHEN jr.status = 'rejected' THEN 'rejected'
		         ELSE 'none'
		       END
		FROM social_groups g
		JOIN users u ON u.id = g.created_by
		LEFT JOIN social_group_members mem ON mem.group_id = g.id AND mem.user_id = $1
		LEFT JOIN social_group_join_requests jr ON jr.group_id = g.id AND jr.user_id = $1`
	if mineOnly {
		query += ` WHERE EXISTS (
			SELECT 1 FROM social_group_members mm WHERE mm.group_id = g.id AND mm.user_id = $1
		)`
	}
	query += ` ORDER BY g.updated_at DESC LIMIT $2`

	rows, err := r.pool.Query(ctx, query, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.SocialGroup, 0)
	for rows.Next() {
		var g domain.SocialGroup
		var creator domain.SocialAuthor
		var avatar *string
		var myRole *string
		var joinStatus string
		if err := rows.Scan(
			&g.ID, &g.Name, &g.Description, &g.Privacy, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt,
			&g.MemberCount,
			&creator.FirstName, &creator.LastName, &avatar,
			&myRole, &joinStatus,
		); err != nil {
			return nil, err
		}
		creator.ID = g.CreatedBy
		creator.AvatarURL = avatar
		g.Creator = &creator
		g.JoinStatus = domain.SocialGroupJoinStatus(joinStatus)
		if myRole != nil {
			role := domain.SocialGroupMemberRole(*myRole)
			g.MyRole = &role
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *SocialRepository) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]domain.SocialGroupMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.user_id, m.role, m.joined_at, u.first_name, u.last_name, u.avatar_path
		FROM social_group_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.group_id = $1
		ORDER BY CASE WHEN m.role = 'admin' THEN 0 ELSE 1 END, m.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SocialGroupMember, 0)
	for rows.Next() {
		var m domain.SocialGroupMember
		var user domain.SocialAuthor
		var avatar *string
		if err := rows.Scan(&m.UserID, &m.Role, &m.JoinedAt, &user.FirstName, &user.LastName, &avatar); err != nil {
			return nil, err
		}
		user.ID = m.UserID
		user.AvatarURL = avatar
		m.User = &user
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *SocialRepository) UpsertJoinRequest(ctx context.Context, groupID, userID uuid.UUID, message string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO social_group_join_requests (group_id, user_id, message, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (group_id, user_id) DO UPDATE
		SET message = EXCLUDED.message,
		    status = 'pending',
		    created_at = NOW(),
		    reviewed_at = NULL,
		    reviewed_by = NULL
	`, groupID, userID, message)
	return err
}

func (r *SocialRepository) ListJoinRequests(ctx context.Context, groupID uuid.UUID) ([]domain.SocialGroupJoinRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.group_id, r.user_id, r.message, r.status, r.created_at,
		       u.first_name, u.last_name, u.avatar_path
		FROM social_group_join_requests r
		JOIN users u ON u.id = r.user_id
		WHERE r.group_id = $1 AND r.status = 'pending'
		ORDER BY r.created_at ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SocialGroupJoinRequest, 0)
	for rows.Next() {
		var req domain.SocialGroupJoinRequest
		var user domain.SocialAuthor
		var avatar *string
		if err := rows.Scan(
			&req.ID, &req.GroupID, &req.UserID, &req.Message, &req.Status, &req.CreatedAt,
			&user.FirstName, &user.LastName, &avatar,
		); err != nil {
			return nil, err
		}
		user.ID = req.UserID
		user.AvatarURL = avatar
		req.User = &user
		out = append(out, req)
	}
	return out, rows.Err()
}

func (r *SocialRepository) ReviewJoinRequest(ctx context.Context, groupID, userID, reviewerID uuid.UUID, approve bool) error {
	status := "rejected"
	if approve {
		status = "approved"
	}
	ct, err := r.pool.Exec(ctx, `
		UPDATE social_group_join_requests
		SET status = $4, reviewed_at = NOW(), reviewed_by = $3
		WHERE group_id = $1 AND user_id = $2 AND status = 'pending'
	`, groupID, userID, reviewerID, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	if approve {
		return r.AddGroupMember(ctx, groupID, userID, domain.SocialGroupRoleMember)
	}
	return nil
}

func (r *SocialRepository) TouchGroup(ctx context.Context, groupID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE social_groups SET updated_at = NOW() WHERE id = $1`, groupID)
	return err
}

func (r *SocialRepository) CreateGroupMessage(ctx context.Context, msg *domain.SocialGroupMessage) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO social_group_messages (group_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, msg.GroupID, msg.SenderID, msg.Body).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return err
	}
	return r.TouchGroup(ctx, msg.GroupID)
}

func (r *SocialRepository) ListGroupMessages(ctx context.Context, groupID uuid.UUID, limit int, before *time.Time) ([]domain.SocialGroupMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	args := []any{groupID, limit}
	query := `
		SELECT m.id, m.group_id, m.sender_id, m.body, m.created_at,
		       u.first_name, u.last_name, u.avatar_path
		FROM social_group_messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.group_id = $1`
	if before != nil {
		query += ` AND m.created_at < $3`
		args = append(args, *before)
	}
	query += ` ORDER BY m.created_at DESC LIMIT $2`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SocialGroupMessage, 0)
	for rows.Next() {
		var m domain.SocialGroupMessage
		var sender domain.SocialAuthor
		var avatar *string
		if err := rows.Scan(
			&m.ID, &m.GroupID, &m.SenderID, &m.Body, &m.CreatedAt,
			&sender.FirstName, &sender.LastName, &avatar,
		); err != nil {
			return nil, err
		}
		sender.ID = m.SenderID
		sender.AvatarURL = avatar
		m.Sender = &sender
		out = append(out, m)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}

func (r *SocialRepository) CountGroupAdmins(ctx context.Context, groupID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM social_group_members WHERE group_id = $1 AND role = 'admin'
	`, groupID).Scan(&n)
	return n, err
}
