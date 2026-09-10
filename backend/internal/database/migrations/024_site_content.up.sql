-- +goose Up
CREATE TABLE site_gallery_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    caption_fr TEXT NOT NULL DEFAULT '',
    caption_en TEXT NOT NULL DEFAULT '',
    path TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_site_gallery_published ON site_gallery_images (published, sort_order, created_at DESC);

CREATE TABLE featured_postulant (
    id UUID PRIMARY KEY,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    home_club TEXT NOT NULL DEFAULT '',
    quote_fr TEXT NOT NULL DEFAULT '',
    quote_en TEXT NOT NULL DEFAULT '',
    visit_count INTEGER NOT NULL DEFAULT 0 CHECK (visit_count >= 0),
    clubs_visited TEXT[] NOT NULL DEFAULT '{}',
    flyer_path TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO featured_postulant (id) VALUES ('00000000-0000-0000-0000-000000000024');
