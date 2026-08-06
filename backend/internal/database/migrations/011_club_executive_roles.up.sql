-- +goose Up
ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'president_elect';
ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'immediate_past_president';
ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'assistant_secretary';
ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'assistant_treasurer';
ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'protocol';
ALTER TYPE club_mandate_role ADD VALUE IF NOT EXISTS 'assistant_protocol';
