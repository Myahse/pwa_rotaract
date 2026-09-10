-- +goose Up
ALTER TABLE featured_postulant
    ADD COLUMN IF NOT EXISTS period_year INTEGER,
    ADD COLUMN IF NOT EXISTS period_month INTEGER;

UPDATE featured_postulant
SET
    period_year = EXTRACT(YEAR FROM (NOW() - INTERVAL '1 month'))::INTEGER,
    period_month = EXTRACT(MONTH FROM (NOW() - INTERVAL '1 month'))::INTEGER
WHERE period_year IS NULL OR period_month IS NULL;

ALTER TABLE featured_postulant
    ALTER COLUMN period_year SET NOT NULL,
    ALTER COLUMN period_month SET NOT NULL;

ALTER TABLE featured_postulant
    ADD CONSTRAINT featured_postulant_period_month_check
    CHECK (period_month BETWEEN 1 AND 12);

CREATE UNIQUE INDEX IF NOT EXISTS idx_featured_postulant_period
    ON featured_postulant (period_year, period_month);
