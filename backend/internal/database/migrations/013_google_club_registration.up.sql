-- +goose Up
ALTER TABLE users
    ADD COLUMN google_sub TEXT;

CREATE UNIQUE INDEX idx_users_google_sub ON users (google_sub) WHERE google_sub IS NOT NULL;

ALTER TABLE club_registration_requests
    ADD COLUMN access_token TEXT,
    ADD COLUMN access_token_expires_at TIMESTAMPTZ,
    ADD COLUMN access_token_used_at TIMESTAMPTZ,
    ADD COLUMN approved_slug TEXT;

CREATE UNIQUE INDEX idx_club_registration_access_token
    ON club_registration_requests (access_token)
    WHERE access_token IS NOT NULL;
