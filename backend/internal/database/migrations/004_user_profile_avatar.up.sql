-- +goose Up
ALTER TABLE users
    ADD COLUMN avatar_path TEXT;
