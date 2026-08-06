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

type SocialRepository struct {
	pool *pgxpool.Pool
}

func NewSocialRepository(pool *pgxpool.Pool) *SocialRepository {
	return &SocialRepository{pool: pool}
}

type SocialPostRow struct {
	Post               domain.SocialPost
	AuthorAvatar       *string
	ClubName           *string
	CommentCount       int
	ReactionCount      int
	ReactedByMe        bool
	AuthorFollowedByMe bool
}

func (r *SocialRepository) CreatePost(ctx context.Context, post *domain.SocialPost) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO social_posts (author_id, club_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at, is_hidden
	`, post.AuthorID, post.ClubID, post.Body).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt, &post.IsHidden)
}

func (r *SocialRepository) AddMedia(ctx context.Context, media *domain.SocialPostMedia) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO social_post_media (post_id, kind, path, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, media.PostID, media.Kind, media.Path, media.SortOrder).Scan(&media.ID, &media.CreatedAt)
}

func (r *SocialRepository) ListMedia(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]domain.SocialPostMedia, error) {
	out := make(map[uuid.UUID][]domain.SocialPostMedia)
	if len(postIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, post_id, kind, path, sort_order, created_at
		FROM social_post_media
		WHERE post_id = ANY($1)
		ORDER BY post_id, sort_order ASC, created_at ASC
	`, postIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m domain.SocialPostMedia
		if err := rows.Scan(&m.ID, &m.PostID, &m.Kind, &m.Path, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, err
		}
		out[m.PostID] = append(out[m.PostID], m)
	}
	return out, rows.Err()
}

func (r *SocialRepository) GetPost(ctx context.Context, postID, viewerID uuid.UUID) (*SocialPostRow, error) {
	row := r.pool.QueryRow(ctx, socialPostSelectFast+`
		WHERE p.id = ANY($2) AND (p.is_hidden = FALSE OR p.author_id = $1)
	`, viewerID, []uuid.UUID{postID})
	return scanSocialPostRow(row)
}

func (r *SocialRepository) ListFeed(ctx context.Context, viewerID uuid.UUID, limit int, before *time.Time, followingOnly bool) ([]SocialPostRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// 1) Cheap ID page first (uses feed index), then hydrate only those rows.
	idArgs := []any{limit}
	idQuery := `
		SELECT p.id
		FROM social_posts p
		WHERE p.is_hidden = FALSE`
	if followingOnly {
		idQuery += `
		  AND EXISTS (
			SELECT 1 FROM social_follows f
			WHERE f.follower_id = $2 AND f.following_id = p.author_id
		  )`
		idArgs = append(idArgs, viewerID)
	}
	if before != nil {
		argN := len(idArgs) + 1
		idQuery += fmt.Sprintf(` AND p.created_at < $%d`, argN)
		idArgs = append(idArgs, *before)
	}
	idQuery += ` ORDER BY p.created_at DESC LIMIT $1`

	idRows, err := r.pool.Query(ctx, idQuery, idArgs...)
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

func (r *SocialRepository) DeletePost(ctx context.Context, postID, requesterID uuid.UUID, isAdmin bool) error {
	ct, err := r.pool.Exec(ctx, `
		DELETE FROM social_posts
		WHERE id = $1 AND ($2 OR author_id = $3)
	`, postID, isAdmin, requesterID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SocialRepository) CreateComment(ctx context.Context, comment *domain.SocialComment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO social_comments (post_id, author_id, parent_id, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, comment.PostID, comment.AuthorID, comment.ParentID, comment.Body).Scan(&comment.ID, &comment.CreatedAt)
}

func (r *SocialRepository) GetComment(ctx context.Context, commentID uuid.UUID) (*domain.SocialComment, error) {
	var c domain.SocialComment
	var author domain.SocialAuthor
	var avatar *string
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.post_id, c.author_id, c.parent_id, c.body, c.created_at,
		       u.first_name, u.last_name, u.avatar_path
		FROM social_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.id = $1
	`, commentID).Scan(
		&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Body, &c.CreatedAt,
		&author.FirstName, &author.LastName, &avatar,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	author.ID = c.AuthorID
	author.AvatarURL = avatar
	c.Author = &author
	return &c, nil
}

func (r *SocialRepository) ListComments(ctx context.Context, postID uuid.UUID) ([]domain.SocialComment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.post_id, c.author_id, c.parent_id, c.body, c.created_at,
		       u.first_name, u.last_name, u.avatar_path
		FROM social_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.post_id = $1
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flat := make([]domain.SocialComment, 0)
	for rows.Next() {
		var c domain.SocialComment
		var author domain.SocialAuthor
		var avatar *string
		if err := rows.Scan(
			&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Body, &c.CreatedAt,
			&author.FirstName, &author.LastName, &avatar,
		); err != nil {
			return nil, err
		}
		author.ID = c.AuthorID
		author.AvatarURL = avatar
		c.Author = &author
		flat = append(flat, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	byParent := make(map[uuid.UUID][]domain.SocialComment)
	roots := make([]domain.SocialComment, 0)
	for _, c := range flat {
		if c.ParentID == nil {
			roots = append(roots, c)
			continue
		}
		byParent[*c.ParentID] = append(byParent[*c.ParentID], c)
	}
	for i := range roots {
		roots[i].Replies = byParent[roots[i].ID]
		if roots[i].Replies == nil {
			roots[i].Replies = []domain.SocialComment{}
		}
	}
	return roots, nil
}

func (r *SocialRepository) AddReaction(ctx context.Context, postID, userID uuid.UUID, kind string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO social_reactions (post_id, user_id, kind)
		VALUES ($1, $2, $3)
		ON CONFLICT (post_id, user_id) DO UPDATE SET kind = EXCLUDED.kind
	`, postID, userID, kind)
	return err
}

func (r *SocialRepository) RemoveReaction(ctx context.Context, postID, userID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `
		DELETE FROM social_reactions WHERE post_id = $1 AND user_id = $2
	`, postID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SocialRepository) PostExistsVisible(ctx context.Context, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM social_posts WHERE id = $1 AND is_hidden = FALSE)
	`, postID).Scan(&exists)
	return exists, err
}

func (r *SocialRepository) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO social_follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, followerID, followingID)
	return err
}

func (r *SocialRepository) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM social_follows WHERE follower_id = $1 AND following_id = $2
	`, followerID, followingID)
	return err
}

func (r *SocialRepository) ListFollowing(ctx context.Context, userID uuid.UUID) ([]domain.SocialFollowProfile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.avatar_path,
		       TRUE AS followed_by_me,
		       COALESCE(fc.cnt, 0) AS followers_count,
		       COALESCE(fr.status, 'none') AS friendship_status
		FROM social_follows f
		JOIN users u ON u.id = f.following_id
		LEFT JOIN (
			SELECT following_id, COUNT(*)::int AS cnt
			FROM social_follows
			GROUP BY following_id
		) fc ON fc.following_id = u.id
		LEFT JOIN social_friendships fr
		  ON LEAST(fr.requester_id, fr.addressee_id) = LEAST($1::uuid, u.id)
		 AND GREATEST(fr.requester_id, fr.addressee_id) = GREATEST($1::uuid, u.id)
		WHERE f.follower_id = $1 AND u.is_active = TRUE
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFollowProfiles(rows, true)
}

func (r *SocialRepository) ListSuggestions(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.SocialFollowProfile, error) {
	if limit <= 0 || limit > 30 {
		limit = 12
	}
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.avatar_path,
		       FALSE AS followed_by_me,
		       COALESCE(fc.cnt, 0) AS followers_count,
		       'none'::text AS friendship_status
		FROM users u
		JOIN club_memberships cm ON cm.user_id = u.id
		LEFT JOIN (
			SELECT following_id, COUNT(*)::int AS cnt
			FROM social_follows
			GROUP BY following_id
		) fc ON fc.following_id = u.id
		WHERE u.is_active = TRUE
		  AND u.id <> $1
		  AND NOT EXISTS (
			SELECT 1 FROM social_follows f WHERE f.follower_id = $1 AND f.following_id = u.id
		  )
		GROUP BY u.id, u.first_name, u.last_name, u.avatar_path, u.created_at, fc.cnt
		ORDER BY COALESCE(fc.cnt, 0) DESC, u.created_at DESC
		LIMIT $2
	`, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFollowProfiles(rows, true)
}

func (r *SocialRepository) PublicSuggestions(ctx context.Context, limit int) ([]domain.SocialFollowProfile, error) {
	if limit <= 0 || limit > 30 {
		limit = 12
	}
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.avatar_path,
		       FALSE AS followed_by_me,
		       COALESCE(fc.cnt, 0) AS followers_count,
		       'none'::text AS friendship_status
		FROM users u
		JOIN club_memberships cm ON cm.user_id = u.id
		LEFT JOIN (
			SELECT following_id, COUNT(*)::int AS cnt
			FROM social_follows
			GROUP BY following_id
		) fc ON fc.following_id = u.id
		WHERE u.is_active = TRUE
		GROUP BY u.id, u.first_name, u.last_name, u.avatar_path, u.created_at, fc.cnt
		ORDER BY COALESCE(fc.cnt, 0) DESC, u.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFollowProfiles(rows, true)
}

func scanFollowProfiles(rows pgx.Rows, withFriendship bool) ([]domain.SocialFollowProfile, error) {
	out := make([]domain.SocialFollowProfile, 0)
	for rows.Next() {
		var p domain.SocialFollowProfile
		var avatar *string
		var status string
		var err error
		if withFriendship {
			err = rows.Scan(&p.ID, &p.FirstName, &p.LastName, &avatar, &p.FollowedByMe, &p.FollowersCount, &status)
		} else {
			err = rows.Scan(&p.ID, &p.FirstName, &p.LastName, &avatar, &p.FollowedByMe, &p.FollowersCount)
			status = string(domain.FriendshipNone)
		}
		if err != nil {
			return nil, err
		}
		p.AvatarURL = avatar
		if status == "" {
			status = string(domain.FriendshipNone)
		}
		p.FriendshipStatus = domain.FriendshipStatus(status)
		out = append(out, p)
	}
	return out, rows.Err()
}

// socialPostSelectFast aggregates counts only for the selected posts ($2 = uuid[]).
// $1 is always the viewer id for reacted/followed flags.
const socialPostSelectFast = `
	SELECT p.id, p.author_id, p.club_id, p.body, p.is_hidden, p.created_at, p.updated_at,
	       u.first_name, u.last_name, u.avatar_path,
	       c.name AS club_name,
	       COALESCE(cc.cnt, 0) AS comment_count,
	       COALESCE(rc.cnt, 0) AS reaction_count,
	       EXISTS(SELECT 1 FROM social_reactions sr2 WHERE sr2.post_id = p.id AND sr2.user_id = $1) AS reacted_by_me,
	       EXISTS(SELECT 1 FROM social_follows ff WHERE ff.follower_id = $1 AND ff.following_id = p.author_id) AS author_followed_by_me
	FROM social_posts p
	JOIN users u ON u.id = p.author_id
	LEFT JOIN clubs c ON c.id = p.club_id
	LEFT JOIN (
		SELECT post_id, COUNT(*)::int AS cnt
		FROM social_comments
		WHERE post_id = ANY($2)
		GROUP BY post_id
	) cc ON cc.post_id = p.id
	LEFT JOIN (
		SELECT post_id, COUNT(*)::int AS cnt
		FROM social_reactions
		WHERE post_id = ANY($2)
		GROUP BY post_id
	) rc ON rc.post_id = p.id
`

func scanSocialPostRow(row pgx.Row) (*SocialPostRow, error) {
	var item SocialPostRow
	var first, last string
	var avatar *string
	var clubName *string
	err := row.Scan(
		&item.Post.ID, &item.Post.AuthorID, &item.Post.ClubID, &item.Post.Body, &item.Post.IsHidden,
		&item.Post.CreatedAt, &item.Post.UpdatedAt,
		&first, &last, &avatar, &clubName,
		&item.CommentCount, &item.ReactionCount, &item.ReactedByMe, &item.AuthorFollowedByMe,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	fillPostRow(&item, first, last, avatar, clubName)
	return &item, nil
}

func scanSocialPostRows(rows pgx.Rows) (*SocialPostRow, error) {
	var item SocialPostRow
	var first, last string
	var avatar *string
	var clubName *string
	err := rows.Scan(
		&item.Post.ID, &item.Post.AuthorID, &item.Post.ClubID, &item.Post.Body, &item.Post.IsHidden,
		&item.Post.CreatedAt, &item.Post.UpdatedAt,
		&first, &last, &avatar, &clubName,
		&item.CommentCount, &item.ReactionCount, &item.ReactedByMe, &item.AuthorFollowedByMe,
	)
	if err != nil {
		return nil, err
	}
	fillPostRow(&item, first, last, avatar, clubName)
	return &item, nil
}

func fillPostRow(item *SocialPostRow, first, last string, avatar, clubName *string) {
	item.AuthorAvatar = avatar
	item.ClubName = clubName
	item.Post.Author = &domain.SocialAuthor{ID: item.Post.AuthorID, FirstName: first, LastName: last}
	if item.Post.ClubID != nil && clubName != nil {
		item.Post.Club = &domain.SocialClubTag{ID: *item.Post.ClubID, Name: *clubName}
	}
	item.Post.CommentCount = item.CommentCount
	item.Post.ReactionCount = item.ReactionCount
	item.Post.ReactedByMe = item.ReactedByMe
	item.Post.AuthorFollowedByMe = item.AuthorFollowedByMe
}
