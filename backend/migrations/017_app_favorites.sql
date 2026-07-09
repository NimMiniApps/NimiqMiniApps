CREATE TABLE IF NOT EXISTS app_favorites (
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    wallet_address TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (app_id, wallet_address)
);

CREATE INDEX IF NOT EXISTS idx_app_favorites_wallet ON app_favorites(wallet_address);
