ALTER TABLE apps ADD COLUMN IF NOT EXISTS developer_wallet_address TEXT REFERENCES users(wallet_address);

CREATE INDEX IF NOT EXISTS apps_developer_wallet_address_idx
    ON apps (developer_wallet_address)
    WHERE developer_wallet_address IS NOT NULL;
