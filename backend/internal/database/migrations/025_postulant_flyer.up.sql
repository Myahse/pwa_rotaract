-- +goose Up
ALTER TABLE featured_postulant ADD COLUMN IF NOT EXISTS flyer_path TEXT;

-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'featured_postulant'
      AND column_name = 'photo_path'
  ) THEN
    UPDATE featured_postulant
    SET flyer_path = photo_path
    WHERE (flyer_path IS NULL OR flyer_path = '') AND photo_path IS NOT NULL AND photo_path <> '';
    ALTER TABLE featured_postulant DROP COLUMN photo_path;
  END IF;
END $$;
-- +goose StatementEnd
