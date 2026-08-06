-- +goose Up
CREATE TABLE email_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role_id UUID REFERENCES club_roles(id) ON DELETE SET NULL,
    token TEXT NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_invites_club ON email_invites (club_id);
CREATE INDEX idx_email_invites_email ON email_invites (LOWER(email));
CREATE UNIQUE INDEX idx_email_invites_pending_email_club
    ON email_invites (club_id, LOWER(email))
    WHERE used_at IS NULL;

INSERT INTO permissions (key, description) VALUES
    ('club.invites.send', 'Send registration invite emails to members')
ON CONFLICT (key) DO NOTHING;

-- Grant to existing head roles
INSERT INTO club_role_permissions (club_role_id, permission_id)
SELECT cr.id, p.id
FROM club_roles cr
JOIN permissions p ON p.key = 'club.invites.send'
WHERE cr.name = 'Responsable de club'
ON CONFLICT DO NOTHING;
