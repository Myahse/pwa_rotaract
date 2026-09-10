-- +goose Up
CREATE TABLE public_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    starts_at TIMESTAMPTZ NOT NULL,
    city TEXT NOT NULL DEFAULT '',
    venue_fr TEXT NOT NULL DEFAULT '',
    venue_en TEXT NOT NULL DEFAULT '',
    title_fr TEXT NOT NULL,
    title_en TEXT NOT NULL DEFAULT '',
    summary_fr TEXT NOT NULL DEFAULT '',
    summary_en TEXT NOT NULL DEFAULT '',
    body_fr TEXT NOT NULL DEFAULT '',
    body_en TEXT NOT NULL DEFAULT '',
    cta_label_fr TEXT NOT NULL DEFAULT '',
    cta_label_en TEXT NOT NULL DEFAULT '',
    cta_url TEXT NOT NULL DEFAULT '',
    flyer_path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_public_events_starts ON public_events (starts_at DESC);
CREATE INDEX idx_public_events_published ON public_events (published, starts_at DESC);

CREATE TABLE public_event_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES public_events(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_public_event_images_event ON public_event_images (event_id, sort_order);

INSERT INTO public_events (
    published, starts_at, city, venue_fr, venue_en, title_fr, title_en, summary_fr, summary_en
) VALUES
(
    TRUE, '2026-10-18T09:00:00+00:00', 'Abidjan',
    'Palais de la Culture, Treichville', 'Palace of Culture, Treichville',
    'Forum national Rotaract CIV', 'Rotaract CIV National Forum',
    'Rencontre annuelle des clubs : projets de service, leadership et amitié, avec les responsables nationaux.',
    'Annual gathering of clubs: service projects, leadership and fellowship with the national team.'
),
(
    TRUE, '2026-10-24T08:00:00+00:00', 'Abidjan',
    'Sites partenaires à Abidjan', 'Partner sites in Abidjan',
    'Journée nationale de service', 'National Day of Service',
    'Actions concertées des clubs dans les quartiers : santé, éducation et environnement.',
    'Coordinated club actions in local communities: health, education and the environment.'
),
(
    TRUE, '2026-11-14T09:00:00+00:00', 'Yamoussoukro',
    'Maison du Rotary, Yamoussoukro', 'Rotary House, Yamoussoukro',
    'Formation des présidents de club', 'Club presidents training',
    'Atelier pour les nouveaux bureaux : gouvernance, mandats, commissions et vie de club.',
    'Workshop for incoming boards: governance, mandates, commissions and club life.'
),
(
    TRUE, '2026-06-21T09:00:00+00:00', 'Abidjan',
    'Hôtel Ivoire, Cocody', 'Hôtel Ivoire, Cocody',
    'Assemblée annuelle 2025–2026', 'Annual assembly 2025–2026',
    'Bilan des mandats, passation et orientations pour l’année Rotaract suivante.',
    'Mandate reviews, handover and direction for the next Rotaract year.'
);

