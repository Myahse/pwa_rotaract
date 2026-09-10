-- +goose Down
DROP TABLE IF EXISTS donations;
DROP INDEX IF EXISTS idx_access_requests_existing_user;
ALTER TABLE access_requests DROP COLUMN IF EXISTS existing_user_id;
