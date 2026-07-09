CREATE TABLE IF NOT EXISTS users (
    wallet_address TEXT PRIMARY KEY,
    display_name TEXT CHECK (display_name IS NULL OR char_length(display_name) <= 50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
