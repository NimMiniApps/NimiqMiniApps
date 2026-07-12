CREATE TABLE IF NOT EXISTS submit_rate_limits (
    ip TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS submit_rate_limits_ip_created_idx
    ON submit_rate_limits (ip, created_at DESC);
