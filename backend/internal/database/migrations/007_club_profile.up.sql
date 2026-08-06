-- +goose Up
ALTER TABLE clubs
    ADD COLUMN logo_path TEXT,
    ADD COLUMN founded_at DATE;
