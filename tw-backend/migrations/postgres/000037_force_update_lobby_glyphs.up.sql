-- Force update the glyphs and positions for Lobby entities
-- to ensure they match the latest configuration even if previously seeded.

-- 1. Update Statue
UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"🗿"'),
    x = 5.0, y = 5.0
WHERE name = 'statue' AND world_id = '00000000-0000-0000-0000-000000000000';

-- 2. Update North Portal
UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"🌀"'),
    x = 5.0, y = 10.0
WHERE name = 'North Portal' AND world_id = '00000000-0000-0000-0000-000000000000';

-- 3. Update East Portal
UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"🌀"'),
    x = 10.0, y = 5.0
WHERE name = 'East Portal' AND world_id = '00000000-0000-0000-0000-000000000000';

-- 4. Update South Portal
UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"🌀"'),
    x = 5.0, y = 0.0
WHERE name = 'South Portal' AND world_id = '00000000-0000-0000-0000-000000000000';

-- 5. Update West Portal
UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"🌀"'),
    x = 0.0, y = 5.0
WHERE name = 'West Portal' AND world_id = '00000000-0000-0000-0000-000000000000';
