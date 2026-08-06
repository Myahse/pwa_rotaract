-- +goose Down
DROP TABLE IF EXISTS club_mandate_assignments;
DROP TABLE IF EXISTS club_mandates;
DROP TYPE IF EXISTS club_mandate_role;
