-- +goose Up
ALTER TABLE donations ADD COLUMN receipt_path TEXT NOT NULL DEFAULT '';
ALTER TABLE donations ADD COLUMN receipt_mime TEXT NOT NULL DEFAULT '';
