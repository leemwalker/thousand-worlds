-- World cleanup migration: removes all worlds except Lobby
-- Step 1: Delete characters in non-Lobby worlds first (to avoid FK constraint violations)
DELETE FROM characters 
WHERE world_id != '00000000-0000-0000-0000-000000000000';

-- Step 2: Delete non-Lobby worlds
DELETE FROM worlds 
WHERE id != '00000000-0000-0000-0000-000000000000';

-- Step 3: Add unique index on lower case name to prevent duplicate world names
CREATE UNIQUE INDEX IF NOT EXISTS unique_world_name_lower ON worlds (LOWER(name));
