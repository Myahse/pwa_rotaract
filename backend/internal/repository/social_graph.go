package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rotaract-civ/backend/internal/domain"
)

func orderedPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func (r *SocialRepository) AreFriends(ctx context.Context, a, b uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM social_friendships
			WHERE status = 'accepted'
			  AND LEAST(requester_id, addressee_id) = LEAST($1::uuid, $2::uuid)
			  AND GREATEST(requester_id, addressee_id) = GREATEST($1::uuid, $2::uuid)
		)
	`, a, b).Scan(&ok)
	return ok, err
}

func (r *SocialRepository) GetFriendship(ctx context.Context, a, b uuid.UUID) (status domain.FriendshipStatus, requesterID uuid.UUID, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT status, requester_id
		FROM social_friendships
		WHERE LEAST(requester_id, addressee_id) = LEAST($1::uuid, $2::uuid)
		  AND GREATEST(requester_id, addressee_id) = GREATEST($1::uuid, $2::uuid)
	`, a, b).Scan(&status, &requesterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FriendshipNone, uuid.Nil, nil
	}
	return status, requesterID, err
}

func (r *SocialRepository) RequestFriendship(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	status, _, err := r.GetFriendship(ctx, requesterID, addresseeID)
	if err != nil {
		return err
	}
	switch status {
	case domain.FriendshipAccepted, domain.FriendshipPending:
		return nil
	case domain.FriendshipDeclined:
		_, err = r.pool.Exec(ctx, `
			UPDATE social_friendships
			SET requester_id = $1,
			    addressee_id = $2,
			    status = 'pending',
			    updated_at = NOW()
			WHERE LEAST(requester_id, addressee_id) = LEAST($1::uuid, $2::uuid)
			  AND GREATEST(requester_id, addressee_id) = GREATEST($1::uuid, $2::uuid)
		`, requesterID, addresseeID)
		return err
	default:
		_, err = r.pool.Exec(ctx, `
			INSERT INTO social_friendships (requester_id, addressee_id, status)
			VALUES ($1, $2, 'pending')
		`, requesterID, addresseeID)
		return err
	}
}

func (r *SocialRepository) RespondFriendship(ctx context.Context, addresseeID, requesterID uuid.UUID, accept bool) error {
	status := "declined"
	if accept {
		status = "accepted"
	}
	ct, err := r.pool.Exec(ctx, `
		UPDATE social_friendships
		SET status = $3, updated_at = NOW()
		WHERE requester_id = $1 AND addressee_id = $2 AND status = 'pending'
	`, requesterID, addresseeID, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SocialRepository) RemoveFriendship(ctx context.Context, a, b uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `
		DELETE FROM social_friendships
		WHERE LEAST(requester_id, addressee_id) = LEAST($1::uuid, $2::uuid)
		  AND GREATEST(requester_id, addressee_id) = GREATEST($1::uuid, $2::uuid)
	`, a, b)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SocialRepository) ListFriends(ctx context.Context, userID uuid.UUID) ([]domain.SocialFriendProfile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.avatar_path, f.updated_at
		FROM social_friendships f
		JOIN users u ON u.id = CASE
			WHEN f.requester_id = $1 THEN f.addressee_id
			ELSE f.requester_id
		END
		WHERE f.status = 'accepted'
		  AND (f.requester_id = $1 OR f.addressee_id = $1)
		ORDER BY u.first_name ASC, u.last_name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.SocialFriendProfile, 0)
	for rows.Next() {
		var p domain.SocialFriendProfile
		var since time.Time
		if err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.AvatarURL, &since); err != nil {
			return nil, err
		}
		p.FriendshipStatus = domain.FriendshipAccepted
		p.FriendsSince = &since
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *SocialRepository) ListFriendRequests(ctx context.Context, userID uuid.UUID) ([]domain.SocialFriendProfile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.avatar_path
		FROM social_friendships f
		JOIN users u ON u.id = f.requester_id
		WHERE f.addressee_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.SocialFriendProfile, 0)
	for rows.Next() {
		var p domain.SocialFriendProfile
		if err := rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.AvatarURL); err != nil {
			return nil, err
		}
		p.FriendshipStatus = domain.FriendshipPending
		p.IncomingRequest = true
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *SocialRepository) GetUserProfile(ctx context.Context, userID, viewerID uuid.UUID) (*domain.SocialUserProfile, error) {
	var p domain.SocialUserProfile
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.avatar_path,
		       (SELECT COUNT(*)::int FROM social_follows f WHERE f.following_id = u.id),
		       (SELECT COUNT(*)::int FROM social_follows f WHERE f.follower_id = u.id),
		       (SELECT COUNT(*)::int FROM social_friendships f
		         WHERE f.status = 'accepted'
		           AND (f.requester_id = u.id OR f.addressee_id = u.id)),
		       EXISTS(SELECT 1 FROM social_follows f WHERE f.follower_id = $2 AND f.following_id = u.id)
		FROM users u
		WHERE u.id = $1 AND u.is_active = TRUE
	`, userID, viewerID).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.AvatarURL,
		&p.FollowersCount, &p.FollowingCount, &p.FriendsCount, &p.FollowedByMe,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.IsMe = viewerID != uuid.Nil && viewerID == userID
	p.FriendshipStatus = domain.FriendshipNone
	if viewerID != uuid.Nil && viewerID != userID {
		status, requesterID, err := r.GetFriendship(ctx, userID, viewerID)
		if err != nil {
			return nil, err
		}
		p.FriendshipStatus = status
		p.IncomingRequest = status == domain.FriendshipPending && requesterID == userID
	}
	return &p, nil
}

func (r *SocialRepository) ListUserClubs(ctx context.Context, userID uuid.UUID) ([]domain.SocialClubTag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name
		FROM club_memberships m
		JOIN clubs c ON c.id = m.club_id
		WHERE m.user_id = $1 AND c.is_active = TRUE
		ORDER BY c.name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SocialClubTag, 0)
	for rows.Next() {
		var c domain.SocialClubTag
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *SocialRepository) ListUserPosts(ctx context.Context, userID, viewerID uuid.UUID, limit int) ([]SocialPostRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	idRows, err := r.pool.Query(ctx, `
		SELECT p.id
		FROM social_posts p
		WHERE p.author_id = $1 AND (p.is_hidden = FALSE OR p.author_id = $2)
		ORDER BY p.created_at DESC
		LIMIT $3
	`, userID, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer idRows.Close()
	ids := make([]uuid.UUID, 0, limit)
	for idRows.Next() {
		var id uuid.UUID
		if err := idRows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := idRows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []SocialPostRow{}, nil
	}

	rows, err := r.pool.Query(ctx, socialPostSelectFast+`
		WHERE p.id = ANY($2)
		ORDER BY p.created_at DESC
	`, viewerID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SocialPostRow, 0, len(ids))
	for rows.Next() {
		item, err := scanSocialPostRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *SocialRepository) GetOrCreateDM(ctx context.Context, a, b uuid.UUID) (uuid.UUID, error) {
	low, high := orderedPair(a, b)
	var conversationID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT conversation_id FROM social_dm_pairs WHERE user_low = $1 AND user_high = $2
	`, low, high).Scan(&conversationID)
	if err == nil {
		return conversationID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `
		INSERT INTO social_conversations DEFAULT VALUES RETURNING id
	`).Scan(&conversationID); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO social_conversation_members (conversation_id, user_id) VALUES ($1, $2), ($1, $3)
	`, conversationID, a, b); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO social_dm_pairs (conversation_id, user_low, user_high) VALUES ($1, $2, $3)
	`, conversationID, low, high); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return conversationID, nil
}

func (r *SocialRepository) IsConversationMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM social_conversation_members
			WHERE conversation_id = $1 AND user_id = $2
		)
	`, conversationID, userID).Scan(&ok)
	return ok, err
}

func (r *SocialRepository) ListConversations(ctx context.Context, userID uuid.UUID) ([]domain.SocialConversation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.updated_at,
		       peer.id, peer.first_name, peer.last_name, peer.avatar_path,
		       lm.id, lm.sender_id, lm.body, lm.shared_comment_id, lm.shared_post_id, lm.created_at
		FROM social_conversations c
		JOIN social_conversation_members m ON m.conversation_id = c.id AND m.user_id = $1
		JOIN social_conversation_members m2 ON m2.conversation_id = c.id AND m2.user_id <> $1
		JOIN users peer ON peer.id = m2.user_id
		LEFT JOIN LATERAL (
			SELECT id, sender_id, body, shared_comment_id, shared_post_id, created_at
			FROM social_messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) lm ON TRUE
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.SocialConversation, 0)
	for rows.Next() {
		var conv domain.SocialConversation
		var peer domain.SocialAuthor
		var avatar *string
		var msgID *uuid.UUID
		var senderID *uuid.UUID
		var body *string
		var sharedCommentID *uuid.UUID
		var sharedPostID *uuid.UUID
		var createdAt *time.Time
		if err := rows.Scan(
			&conv.ID, &conv.UpdatedAt,
			&peer.ID, &peer.FirstName, &peer.LastName, &avatar,
			&msgID, &senderID, &body, &sharedCommentID, &sharedPostID, &createdAt,
		); err != nil {
			return nil, err
		}
		peer.AvatarURL = avatar
		conv.Peer = &peer
		if msgID != nil {
			conv.LastMessage = &domain.SocialMessage{
				ID:              *msgID,
				ConversationID:  conv.ID,
				SenderID:        *senderID,
				Body:            derefStr(body),
				SharedCommentID: sharedCommentID,
				SharedPostID:    sharedPostID,
				CreatedAt:       *createdAt,
			}
		}
		out = append(out, conv)
	}
	return out, rows.Err()
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *SocialRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, limit int, before *time.Time) ([]domain.SocialMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	args := []any{conversationID, limit}
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.body, m.shared_comment_id, m.shared_post_id, m.created_at,
		       u.first_name, u.last_name, u.avatar_path,
		       sc.id, sc.post_id, sc.author_id, sc.parent_id, sc.body, sc.created_at,
		       scu.first_name, scu.last_name
		FROM social_messages m
		JOIN users u ON u.id = m.sender_id
		LEFT JOIN social_comments sc ON sc.id = m.shared_comment_id
		LEFT JOIN users scu ON scu.id = sc.author_id
		WHERE m.conversation_id = $1`
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

	out := make([]domain.SocialMessage, 0)
	for rows.Next() {
		var m domain.SocialMessage
		var sender domain.SocialAuthor
		var avatar *string
		var scID *uuid.UUID
		var scPostID *uuid.UUID
		var scAuthorID *uuid.UUID
		var scParentID *uuid.UUID
		var scBody *string
		var scCreated *time.Time
		var scFirst, scLast *string
		if err := rows.Scan(
			&m.ID, &m.ConversationID, &m.SenderID, &m.Body, &m.SharedCommentID, &m.SharedPostID, &m.CreatedAt,
			&sender.FirstName, &sender.LastName, &avatar,
			&scID, &scPostID, &scAuthorID, &scParentID, &scBody, &scCreated,
			&scFirst, &scLast,
		); err != nil {
			return nil, err
		}
		sender.ID = m.SenderID
		sender.AvatarURL = avatar
		m.Sender = &sender
		if scID != nil {
			m.SharedComment = &domain.SocialComment{
				ID: *scID, PostID: *scPostID, AuthorID: *scAuthorID, ParentID: scParentID,
				Body: derefStr(scBody), CreatedAt: *scCreated,
				Author: &domain.SocialAuthor{ID: *scAuthorID, FirstName: derefStr(scFirst), LastName: derefStr(scLast)},
			}
		}
		out = append(out, m)
	}
	// reverse to chronological
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}

func (r *SocialRepository) CreateMessage(ctx context.Context, msg *domain.SocialMessage) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO social_messages (conversation_id, sender_id, body, shared_comment_id, shared_post_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, msg.ConversationID, msg.SenderID, msg.Body, msg.SharedCommentID, msg.SharedPostID).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE social_conversations SET updated_at = NOW() WHERE id = $1
	`, msg.ConversationID)
	return err
}

func (r *SocialRepository) UserExistsActive(ctx context.Context, userID uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = TRUE)`, userID).Scan(&ok)
	return ok, err
}
