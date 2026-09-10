-- +goose Down
DROP INDEX IF EXISTS idx_social_posts_reposted;
ALTER TABLE social_posts DROP COLUMN IF EXISTS reposted_post_id;