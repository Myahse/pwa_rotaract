-- +goose Down
ALTER TABLE social_posts DROP CONSTRAINT IF EXISTS social_posts_body_check;
ALTER TABLE social_posts ADD CONSTRAINT social_posts_body_check
    CHECK (char_length(trim(body)) > 0 AND char_length(body) <= 5000);

DROP TABLE IF EXISTS social_follows;
DROP TABLE IF EXISTS social_post_media;
