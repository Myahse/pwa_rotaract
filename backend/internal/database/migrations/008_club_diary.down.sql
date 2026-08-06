-- +goose Down
DROP TABLE IF EXISTS club_diary_entries;
DROP TYPE IF EXISTS club_diary_entry_type;

DELETE FROM club_role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE key IN ('club.diary.read', 'club.diary.manage')
);

DELETE FROM permissions WHERE key IN ('club.diary.read', 'club.diary.manage');
