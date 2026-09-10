-- +goose Down
ALTER TABLE social_posts DROP COLUMN IF EXISTS quote_body;