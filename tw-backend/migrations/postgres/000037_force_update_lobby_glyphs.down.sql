-- Revert glyphs and positions (approximate revert to defaults if needed)
-- Note: We generally don't want to revert "fixes", but for completeness:

UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"T"'), -- Default marker
    x = 5.0, y = 5.0
WHERE name = 'statue' AND world_id = '00000000-0000-0000-0000-000000000000';

UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"P"'), -- Default portal
    x = 5.0, y = 9.0
WHERE name = 'North Portal' AND world_id = '00000000-0000-0000-0000-000000000000';

UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"P"'),
    x = 9.0, y = 5.0
WHERE name = 'East Portal' AND world_id = '00000000-0000-0000-0000-000000000000';

UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"P"'),
    x = 5.0, y = 1.0
WHERE name = 'South Portal' AND world_id = '00000000-0000-0000-0000-000000000000';

UPDATE world_entities 
SET metadata = jsonb_set(metadata, '{glyph}', '"P"'),
    x = 1.0, y = 5.0
WHERE name = 'West Portal' AND world_id = '00000000-0000-0000-0000-000000000000';
