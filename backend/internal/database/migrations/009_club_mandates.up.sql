-- +goose Up
CREATE TYPE club_mandate_role AS ENUM (
    'president',
    'vice_president',
    'secretary',
    'treasurer',
    'commission_president',
    'commission_member',
    'member'
);

CREATE TABLE club_mandates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    started_at DATE NOT NULL,
    ended_at DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE club_mandate_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mandate_id UUID NOT NULL REFERENCES club_mandates(id) ON DELETE CASCADE,
    role club_mandate_role NOT NULL,
    commission_id UUID REFERENCES commissions(id) ON DELETE SET NULL,
    commission_name TEXT,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    photo_path TEXT,
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_club_mandates_club ON club_mandates (club_id);
CREATE INDEX idx_club_mandates_current ON club_mandates (club_id, is_current);
CREATE INDEX idx_mandate_assignments_mandate ON club_mandate_assignments (mandate_id);
