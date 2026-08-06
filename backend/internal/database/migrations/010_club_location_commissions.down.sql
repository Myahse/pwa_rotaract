-- +goose Down
DROP INDEX IF EXISTS idx_commissions_club_code;
ALTER TABLE commissions DROP COLUMN IF EXISTS code, DROP COLUMN IF EXISTS is_system;
ALTER TABLE clubs DROP COLUMN IF EXISTS country, DROP COLUMN IF EXISTS commune;
