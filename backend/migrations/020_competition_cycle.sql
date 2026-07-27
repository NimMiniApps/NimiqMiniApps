ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS competition_cycle INTEGER
    CHECK (competition_cycle IS NULL OR competition_cycle > 0);

ALTER TABLE app_revisions
    ADD COLUMN IF NOT EXISTS competition_cycle INTEGER
    CHECK (competition_cycle IS NULL OR competition_cycle > 0);

UPDATE apps
SET competition_cycle = 1
WHERE slug IN (
    'xcrowhub',
    'nimagent',
    'nimhunt',
    'nimiq-bazar',
    'roundtrip',
    'nimjump',
    'unlockmedia',
    'nimiq-invoice-pay',
    'nimiqstake',
    'nimconnect',
    'tipwall',
    'nimga',
    'nimiq-radio',
    'verilock'
);
