-- +goose Down
DROP INDEX IF EXISTS idx_featured_postulant_period;

ALTER TABLE featured_postulant
    DROP CONSTRAINT IF EXISTS featured_postulant_period_month_check;

ALTER TABLE featured_postulant
    DROP COLUMN IF EXISTS period_year,
    DROP COLUMN IF EXISTS period_month;
