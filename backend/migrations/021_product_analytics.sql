CREATE TABLE analytics_events (
    id BIGSERIAL PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    event_type TEXT NOT NULL CHECK (event_type IN (
        'catalog_visit', 'wallet_login', 'app_view', 'app_open', 'app_link_copy'
    )),
    visitor_hash TEXT NOT NULL,
    session_hash TEXT NOT NULL,
    wallet_hash TEXT,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE
);

CREATE INDEX analytics_events_occurred_at_idx ON analytics_events (occurred_at);
CREATE INDEX analytics_events_app_time_idx ON analytics_events (app_id, occurred_at)
    WHERE app_id IS NOT NULL;
CREATE INDEX analytics_events_type_time_idx ON analytics_events (event_type, occurred_at);

CREATE UNIQUE INDEX analytics_events_catalog_session_unique
    ON analytics_events (event_type, session_hash)
    WHERE event_type = 'catalog_visit';
CREATE UNIQUE INDEX analytics_events_app_view_session_unique
    ON analytics_events (event_type, app_id, session_hash)
    WHERE event_type = 'app_view';

CREATE TABLE analytics_daily (
    day DATE NOT NULL,
    event_type TEXT NOT NULL,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    total_events BIGINT NOT NULL DEFAULT 0,
    UNIQUE NULLS NOT DISTINCT (day, event_type, app_id)
);

CREATE TABLE analytics_uniques (
    subject_kind TEXT NOT NULL CHECK (subject_kind IN ('visitor', 'wallet')),
    event_type TEXT NOT NULL,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    subject_hash TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    UNIQUE NULLS NOT DISTINCT (subject_kind, event_type, app_id, subject_hash)
);

CREATE INDEX analytics_uniques_range_idx
    ON analytics_uniques (event_type, app_id, first_seen_at, last_seen_at);
