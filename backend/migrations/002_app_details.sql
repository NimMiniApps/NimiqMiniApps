ALTER TABLE apps ADD COLUMN IF NOT EXISTS release_stage TEXT NOT NULL DEFAULT 'released'
    CHECK (release_stage IN ('concept', 'alpha', 'beta', 'released'));

ALTER TABLE apps ADD COLUMN IF NOT EXISTS long_description TEXT NOT NULL DEFAULT '';

ALTER TABLE apps ADD COLUMN IF NOT EXISTS media JSONB NOT NULL DEFAULT '[]';

UPDATE apps
SET media = (
    SELECT COALESCE(jsonb_agg(jsonb_build_object('type', 'image', 'url', url)), '[]'::jsonb)
    FROM unnest(screenshots) AS url
)
WHERE array_length(screenshots, 1) > 0;

ALTER TABLE apps DROP COLUMN IF EXISTS screenshots;

UPDATE apps SET release_stage = 'beta' WHERE slug = 'nimbomber';
UPDATE apps SET release_stage = 'released' WHERE slug = 'playnimiq';
UPDATE apps SET release_stage = 'alpha' WHERE slug = 'nimdoom';
