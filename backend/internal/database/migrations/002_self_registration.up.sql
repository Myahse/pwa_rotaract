-- +goose Up
ALTER TABLE users
    ADD COLUMN birth_date DATE,
    ADD COLUMN profession TEXT,
    ADD COLUMN member_since DATE;

ALTER TABLE clubs
    ADD COLUMN invite_code TEXT;

UPDATE clubs SET invite_code = UPPER(REPLACE(slug, '-', '')) || '-' || SUBSTRING(REPLACE(gen_random_uuid()::text, '-', ''), 1, 6)
WHERE invite_code IS NULL;

ALTER TABLE clubs
    ALTER COLUMN invite_code SET NOT NULL;

CREATE UNIQUE INDEX idx_clubs_invite_code ON clubs (invite_code);

CREATE TYPE access_request_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE access_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID REFERENCES clubs(id) ON DELETE SET NULL,
    club_name TEXT NOT NULL,
    requested_role_id UUID REFERENCES club_roles(id) ON DELETE SET NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    phone TEXT,
    birth_date DATE,
    profession TEXT,
    member_since DATE,
    status access_request_status NOT NULL DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id),
    review_note TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_access_requests_club ON access_requests (club_id);
CREATE INDEX idx_access_requests_status ON access_requests (status);

INSERT INTO permissions (key, description) VALUES
    ('club.requests.review', 'Review and approve membership access requests')
ON CONFLICT (key) DO NOTHING;
