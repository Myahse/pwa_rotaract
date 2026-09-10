-- +goose Up
ALTER TABLE access_requests
    ADD COLUMN existing_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_access_requests_existing_user ON access_requests (existing_user_id);

CREATE TABLE donations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    amount_xof INTEGER NOT NULL CHECK (amount_xof > 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'received')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_donations_email ON donations (LOWER(email));
CREATE INDEX idx_donations_user ON donations (user_id);
CREATE INDEX idx_donations_created ON donations (created_at DESC);
