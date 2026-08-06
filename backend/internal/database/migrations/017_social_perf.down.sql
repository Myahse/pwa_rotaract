-- +goose Down
DROP INDEX IF EXISTS idx_users_active_created;
DROP INDEX IF EXISTS idx_club_memberships_user;
DROP INDEX IF EXISTS idx_social_reactions_post_user;
DROP INDEX IF EXISTS idx_social_comments_post_id;
DROP INDEX IF EXISTS idx_social_follows_follower_created;
