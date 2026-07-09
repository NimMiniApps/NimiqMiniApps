CREATE TABLE IF NOT EXISTS app_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_slug TEXT NOT NULL REFERENCES apps(slug) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    category TEXT NOT NULL,
    developer_slug TEXT NOT NULL,
    developer_name TEXT NOT NULL,
    tagline TEXT NOT NULL,
    description TEXT NOT NULL,
    long_description TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    assets TEXT[] NOT NULL DEFAULT '{}',
    release_stage TEXT NOT NULL DEFAULT 'released',
    website_url TEXT,
    github_url TEXT,
    icon_url TEXT,
    banner_url TEXT,
    media JSONB NOT NULL DEFAULT '[]',
    socials JSONB NOT NULL DEFAULT '[]',
    author_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS app_revisions_one_pending_per_app
    ON app_revisions (app_slug) WHERE status = 'pending';
