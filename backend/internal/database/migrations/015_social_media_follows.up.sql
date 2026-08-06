-- +goose Up
CREATE TABLE social_post_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES social_posts(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('image', 'video')),
    path TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_social_post_media_post ON social_post_media (post_id, sort_order);

CREATE TABLE social_follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id <> following_id)
);

CREATE INDEX idx_social_follows_following ON social_follows (following_id);

-- Allow text-only OR media posts (body may be empty when media exists)
ALTER TABLE social_posts DROP CONSTRAINT IF EXISTS social_posts_body_check;
ALTER TABLE social_posts ADD CONSTRAINT social_posts_body_check
    CHECK (char_length(body) <= 5000);

-- +goose Down
ALTER TABLE social_posts DROP CONSTRAINT IF EXISTS social_posts_body_check;
ALTER TABLE social_posts ADD CONSTRAINT social_posts_body_check
    CHECK (char_length(trim(body)) > 0 AND char_length(body) <= 5000);

DROP TABLE IF EXISTS social_follows;
DROP TABLE IF EXISTS social_post_media;
