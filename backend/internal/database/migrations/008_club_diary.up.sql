-- +goose Up
INSERT INTO permissions (key, description) VALUES
    ('club.diary.read', 'View club genealogy diary'),
    ('club.diary.manage', 'Manage club genealogy diary')
ON CONFLICT (key) DO NOTHING;

CREATE TYPE club_diary_entry_type AS ENUM ('parrain', 'president', 'member');

CREATE TABLE club_diary_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES club_diary_entries(id) ON DELETE SET NULL,
    entry_type club_diary_entry_type NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    organization TEXT,
    photo_path TEXT,
    started_at DATE,
    ended_at DATE,
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_club_diary_club ON club_diary_entries (club_id);
CREATE INDEX idx_club_diary_parent ON club_diary_entries (parent_id);
CREATE INDEX idx_club_diary_type ON club_diary_entries (club_id, entry_type);

INSERT INTO club_role_permissions (club_role_id, permission_id)
SELECT cr.id, p.id
FROM club_roles cr
CROSS JOIN permissions p
WHERE cr.name = 'Responsable de club' AND cr.is_system = TRUE
  AND p.key IN ('club.diary.read', 'club.diary.manage')
ON CONFLICT DO NOTHING;

INSERT INTO club_role_permissions (club_role_id, permission_id)
SELECT cr.id, p.id
FROM club_roles cr
CROSS JOIN permissions p
WHERE cr.name = 'Membre' AND cr.is_system = TRUE
  AND p.key = 'club.diary.read'
ON CONFLICT DO NOTHING;
