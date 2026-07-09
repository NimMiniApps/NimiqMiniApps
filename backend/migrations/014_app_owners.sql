CREATE TABLE IF NOT EXISTS app_owners (
    app_slug TEXT NOT NULL REFERENCES apps(slug) ON DELETE CASCADE,
    wallet_address TEXT NOT NULL REFERENCES users(wallet_address),
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (app_slug, wallet_address)
);

INSERT INTO app_owners (app_slug, wallet_address)
SELECT slug, developer_wallet_address FROM apps
WHERE developer_wallet_address IS NOT NULL
ON CONFLICT DO NOTHING;

DROP INDEX IF EXISTS apps_developer_wallet_address_idx;
ALTER TABLE apps DROP COLUMN IF EXISTS developer_wallet_address;
