-- +goose Down
ALTER TABLE clubs
    DROP COLUMN IF EXISTS logo_path,
    DROP COLUMN IF EXISTS founded_at;
