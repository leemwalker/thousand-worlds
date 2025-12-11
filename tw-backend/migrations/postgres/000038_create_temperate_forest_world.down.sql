-- Revert the Temperate Forest world creation

-- Restore West Portal to original state
UPDATE world_entities 
SET description = 'A mysterious portal shimmers on the west wall, its depths dark and unknowable.',
    details = 'This portal feels different from the others - older, perhaps, or leading to more distant realms. Seasoned travelers speak of rare discoveries beyond its threshold.',
    metadata = metadata - 'destination_world_id'
WHERE name = 'West Portal' 
AND world_id = '00000000-0000-0000-0000-000000000000';

-- Delete the forest world
DELETE FROM worlds WHERE id = '00000000-0000-0000-0000-000000000002';
