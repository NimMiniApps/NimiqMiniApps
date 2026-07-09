CREATE TABLE app_stats_daily (
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    day DATE NOT NULL,
    opens INT NOT NULL DEFAULT 0,
    views INT NOT NULL DEFAULT 0,
    PRIMARY KEY (app_id, day)
);
