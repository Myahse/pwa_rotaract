-- +goose Down
DROP INDEX IF EXISTS idx_club_registration_access_token;
ALTER TABLE club_registration_requests
    DROP COLUMN IF EXISTS access_token,
    DROP COLUMN IF EXISTS access_token_expires_at,
    DROP COLUMN IF EXISTS access_token_used_at,
    DROP COLUMN IF EXISTS approved_slug;

DROP INDEX IF EXISTS idx_users_google_sub;
ALTER TABLE users DROP COLUMN IF EXISTS google_sub;
