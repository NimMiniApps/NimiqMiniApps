CREATE TABLE IF NOT EXISTS app_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    wallet_address TEXT NOT NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body TEXT NOT NULL DEFAULT '' CHECK (char_length(body) <= 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (app_id, wallet_address)
);

CREATE INDEX IF NOT EXISTS app_reviews_app_id_idx ON app_reviews (app_id);
