CREATE TABLE IF NOT EXISTS app_competition_results (
    id         BIGSERIAL PRIMARY KEY,
    app_id     UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    cycle      INTEGER NOT NULL CHECK (cycle > 0),
    score      INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    place      INTEGER CHECK (place IS NULL OR place > 0),
    UNIQUE (app_id, cycle)
);

CREATE INDEX IF NOT EXISTS app_competition_results_cycle_idx
    ON app_competition_results (cycle);

-- Winner is cataloged but was missing competition membership.
UPDATE apps
SET competition_cycle = 1
WHERE slug = 'nimiq-space'
  AND competition_cycle IS NULL;

-- Placeholder scores until official council totals are published.
INSERT INTO app_competition_results (app_id, cycle, score, place)
SELECT id, 1, 0, NULL
FROM apps
WHERE competition_cycle = 1
ON CONFLICT (app_id, cycle) DO NOTHING;

UPDATE app_competition_results AS r
SET place = v.place
FROM (VALUES
    ('nimiq-space', 1),
    ('nimjump', 2),
    ('nimquest', 3)
) AS v(slug, place)
JOIN apps AS a ON a.slug = v.slug
WHERE r.app_id = a.id
  AND r.cycle = 1;
