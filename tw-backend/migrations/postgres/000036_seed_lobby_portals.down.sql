-- Remove the lobby portal entities
DELETE FROM world_entities 
WHERE world_id = '00000000-0000-0000-0000-000000000000'
  AND entity_type = 'structure'
  AND name LIKE '%Portal';
