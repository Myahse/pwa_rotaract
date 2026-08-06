-- +goose Up
-- Speed up social feed / follows / suggestions on Neon

CREATE INDEX IF NOT EXISTS idx_social_follows_follower_created
    ON social_follows (follower_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_social_comments_post_id
    ON social_comments (post_id);

CREATE INDEX IF NOT EXISTS idx_social_reactions_post_user
    ON social_reactions (post_id, user_id);

CREATE INDEX IF NOT EXISTS idx_club_memberships_user
    ON club_memberships (user_id);

CREATE INDEX IF NOT EXISTS idx_users_active_created
    ON users (created_at DESC)
    WHERE is_active = TRUE;
