ALTER TABLE apps ADD COLUMN IF NOT EXISTS featured_order INTEGER NOT NULL DEFAULT 0
    CHECK (featured_order >= 0);

UPDATE apps SET featured_order = 10 WHERE slug = 'nimbomber' AND featured = true;
UPDATE apps SET featured_order = 20 WHERE slug = 'playnimiq' AND featured = true;
