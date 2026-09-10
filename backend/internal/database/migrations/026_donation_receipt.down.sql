-- +goose Down
ALTER TABLE donations DROP COLUMN IF EXISTS receipt_mime;
ALTER TABLE donations DROP COLUMN IF EXISTS receipt_path;
