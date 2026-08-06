-- +goose Up
CREATE TYPE club_registration_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE club_registration_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_name TEXT NOT NULL,
    contact_email TEXT NOT NULL,
    contact_first_name TEXT NOT NULL,
    contact_last_name TEXT NOT NULL,
    phone TEXT,
    description TEXT,
    country TEXT NOT NULL,
    city TEXT NOT NULL,
    commune TEXT NOT NULL,
    founded_at DATE,
    message TEXT,
    status club_registration_status NOT NULL DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id),
    review_note TEXT,
    reviewed_at TIMESTAMPTZ,
    created_club_id UUID REFERENCES clubs(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_club_registration_requests_status ON club_registration_requests (status);
CREATE INDEX idx_club_registration_requests_email ON club_registration_requests (LOWER(contact_email));
CREATE UNIQUE INDEX idx_club_registration_requests_pending_email
    ON club_registration_requests (LOWER(contact_email))
    WHERE status = 'pending';
