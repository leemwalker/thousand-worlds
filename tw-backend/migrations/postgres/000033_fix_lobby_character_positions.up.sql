-- Move all existing lobby characters to center of lobby (5,5)
-- This fixes characters that were created at 0,0 (boundary edge)
UPDATE characters
SET position_x = 5.0, position_y = 5.0
WHERE world_id = '00000000-0000-0000-0000-000000000000'
  AND (position_x = 0 OR position_y = 0 OR position_x IS NULL OR position_y IS NULL);
