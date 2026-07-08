CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    category TEXT NOT NULL,
    developer_slug TEXT NOT NULL,
    developer_name TEXT NOT NULL,
    tagline TEXT NOT NULL,
    description TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    assets TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'submitted'
        CHECK (status IN ('submitted', 'approved', 'verified', 'experimental', 'rejected')),
    featured BOOLEAN NOT NULL DEFAULT false,
    website_url TEXT,
    github_url TEXT,
    icon_url TEXT,
    banner_url TEXT,
    screenshots TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, tags, assets, status, featured) VALUES
(
    'nimbomber', 'NimBomber', 'nimbomber.maestroi.cc', 'Games', 'maestro', 'Maestro',
    'A mobile-first bomber game where your wallet becomes your character.',
    'Play quick matches, private rooms, or AI matches with Nimiq Pay wallet identity.',
    '{games,multiplayer,"wallet identity",beta}', '{NIM}', 'experimental', true
),
(
    'playnimiq', 'PlayNimiq', 'playnimiq.com', 'Games', 'maestro', 'Maestro',
    'Skill-based games with Nimiq wallet identity and rewards.',
    'Skill-based games with Nimiq wallet identity and rewards.',
    '{games,rewards,leaderboard,skill}', '{NIM}', 'approved', true
),
(
    'nimdoom', 'NimDoom', 'nimminiapps.github.io/NimDoom', 'Games', 'maestro', 'Maestro',
    'A fun Doom experiment for Nimiq Pay.',
    'A fun Doom experiment for Nimiq Pay.',
    '{game,doom,experiment}', '{NIM}', 'experimental', false
),
(
    'nimlens', 'NimLens', 'replace-with-domain-later.example', 'Utilities', 'maestro', 'Maestro',
    'Scan prices and convert them into crypto values.',
    'Scan prices and convert them into crypto values.',
    '{scanner,OCR,prices,utility}', '{NIM,USDT}', 'experimental', false
)
ON CONFLICT (slug) DO NOTHING;
