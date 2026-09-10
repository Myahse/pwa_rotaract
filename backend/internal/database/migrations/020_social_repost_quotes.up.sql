-- +goose Up
ALTER TABLE social_posts
    ADD COLUMN quote_body TEXT NOT NULL DEFAULT ''
    CHECK (char_length(quote_body) <= 5000);

DELETE FROM social_post_media
WHERE post_id IN (SELECT id FROM social_posts WHERE reposted_post_id IS NOT NULL);

UPDATE social_posts SET body = '' WHERE reposted_post_id IS NOT NULL;
