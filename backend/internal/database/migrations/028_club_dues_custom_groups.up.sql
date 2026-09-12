-- +goose NO TRANSACTION
-- +goose Up
-- PostgreSQL: new enum values cannot be used in CHECK constraints in the same transaction.

ALTER TYPE chat_group_type ADD VALUE IF NOT EXISTS 'custom';

ALTER TABLE chat_groups DROP CONSTRAINT IF EXISTS chat_groups_check;
ALTER TABLE chat_groups ADD CONSTRAINT chat_groups_check CHECK (
    (group_type = 'club' AND commission_id IS NULL) OR
    (group_type = 'commission' AND commission_id IS NOT NULL) OR
    (group_type = 'custom' AND commission_id IS NULL)
);

CREATE TABLE IF NOT EXISTS club_due_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    due_month DATE NOT NULL,
    amount_xof INTEGER NOT NULL CHECK (amount_xof > 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'received')),
    receipt_path TEXT NOT NULL DEFAULT '',
    receipt_mime TEXT NOT NULL DEFAULT '',
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (club_id, user_id, due_month)
);

CREATE INDEX IF NOT EXISTS idx_club_due_payments_club ON club_due_payments (club_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_club_due_payments_user ON club_due_payments (user_id, due_month DESC);
