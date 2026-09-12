CREATE TABLE member_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_number VARCHAR(32) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ,
    sent_by UUID REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (club_id, user_id),
    UNIQUE (club_id, card_number)
);

CREATE INDEX idx_member_cards_club ON member_cards (club_id);
CREATE INDEX idx_member_cards_unsent ON member_cards (club_id) WHERE sent_at IS NULL;
