-- +goose Up
ALTER TABLE clubs
    ADD COLUMN country TEXT,
    ADD COLUMN commune TEXT;

ALTER TABLE commissions
    ADD COLUMN code TEXT,
    ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX idx_commissions_club_code ON commissions (club_id, code) WHERE code IS NOT NULL;

ALTER TYPE commission_member_role ADD VALUE IF NOT EXISTS 'secretary';

ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'commission_secretary';
