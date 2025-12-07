-- Clean up potential duplicates before adding constraint (Keep Lobby)
DELETE FROM worlds WHERE id != '00000000-0000-0000-0000-000000000000';

-- Add unique index on lower case name to prevent duplicates
CREATE UNIQUE INDEX IF NOT EXISTS unique_world_name_lower ON worlds (LOWER(name));
