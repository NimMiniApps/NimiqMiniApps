CREATE TABLE IF NOT EXISTS app_credentials (
    id           BIGSERIAL PRIMARY KEY,
    app_id       UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    key_prefix   TEXT NOT NULL UNIQUE,
    secret_hash  BYTEA NOT NULL,
    label        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by   TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS app_credentials_app_idx
    ON app_credentials (app_id) WHERE revoked_at IS NULL;

ALTER TABLE app_client_changes DROP CONSTRAINT IF EXISTS app_client_changes_reason_check;
ALTER TABLE app_client_changes ADD CONSTRAINT app_client_changes_reason_check
    CHECK (reason IN (
        'owner_transfer',
        'scopes_changed',
        'status_changed',
        'metadata_changed',
        'credential_revoked'
    ));
