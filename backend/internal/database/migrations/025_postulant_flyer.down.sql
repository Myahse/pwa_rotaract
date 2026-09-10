-- +goose Down
ALTER TABLE featured_postulant ADD COLUMN IF NOT EXISTS photo_path TEXT;

UPDATE featured_postulant
SET photo_path = flyer_path
WHERE (photo_path IS NULL OR photo_path = '') AND flyer_path IS NOT NULL AND flyer_path <> '';

ALTER TABLE featured_postulant DROP COLUMN IF EXISTS flyer_path;
