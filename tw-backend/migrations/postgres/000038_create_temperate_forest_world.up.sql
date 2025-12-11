-- Create the Temperate world for ecosystem testing
-- This world is linked to the West Portal in the lobby
-- UUID: 00000000-0000-0000-0000-000000000002 (fixed for reproducibility)

-- Create the world
INSERT INTO worlds (id, name, shape, metadata, owner_id, is_system_world) VALUES
('00000000-0000-0000-0000-000000000002', 
 'Temperate', 
 'spherical', 
 '{"description": "A temperate world in its primordial state. The air is thick with possibility as life begins to stir.", "biome": "temperate", "age": "precambrian", "climate": "temperate"}',
 '00000000-0000-0000-0000-000000000001',
 true)
ON CONFLICT (id) DO NOTHING;

-- Update the West Portal to link to this world
UPDATE world_entities 
SET metadata = jsonb_set(
    metadata, 
    '{destination_world_id}', 
    '"00000000-0000-0000-0000-000000000002"'
)
WHERE name = 'West Portal' 
AND world_id = '00000000-0000-0000-0000-000000000000';

-- Also update description to mention the forest
UPDATE world_entities 
SET description = 'An ancient portal shimmers on the west wall, framed by carved oak leaves. Through it, you glimpse a verdant forest awaiting exploration.',
    details = 'This portal leads to the Temperate Forest - a rich ecosystem where flora and fauna thrive in balance. Step through to experience nature in its full glory.'
WHERE name = 'West Portal' 
AND world_id = '00000000-0000-0000-0000-000000000000';
