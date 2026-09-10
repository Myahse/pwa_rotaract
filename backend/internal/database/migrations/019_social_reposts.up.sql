-- +goose Up
ALTER TABLE social_posts
    ADD COLUMN reposted_post_id UUID REFERENCES social_posts(id) ON DELETE CASCADE;

CREATE INDEX idx_social_posts_reposted ON social_posts (reposted_post_id)
    WHERE reposted_post_id IS NOT NULL;
