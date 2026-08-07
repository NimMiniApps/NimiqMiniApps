ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS declared_scopes TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE app_revisions
    ADD COLUMN IF NOT EXISTS declared_scopes TEXT[] NOT NULL DEFAULT '{}';

CREATE TABLE IF NOT EXISTS app_client_changes (
    id BIGSERIAL PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    reason TEXT NOT NULL
        CHECK (reason IN ('owner_transfer', 'scopes_changed', 'status_changed', 'metadata_changed')),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS app_client_changes_since_idx
    ON app_client_changes (occurred_at, app_id);
