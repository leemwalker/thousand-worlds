DROP TABLE IF EXISTS portals;
ALTER TABLE worlds DROP COLUMN IF EXISTS radius;
ALTER TABLE worlds DROP COLUMN IF EXISTS circumference;
DROP INDEX IF EXISTS idx_character_position;
ALTER TABLE characters DROP COLUMN IF EXISTS position_z;
ALTER TABLE characters DROP COLUMN IF EXISTS position_y;
ALTER TABLE characters DROP COLUMN IF EXISTS position_x;
