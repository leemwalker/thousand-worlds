-- Add bounds to the Lobby world for bounded movement (10x10 grid centered at 0,0)
-- This enables proper spatial boundaries in the lobby instead of spherical wrapping
UPDATE worlds
SET shape = 'cube',
    bounds_min_x = 0,
    bounds_min_y = 0,
    bounds_min_z = 0,
    bounds_max_x = 10,
    bounds_max_y = 10,
    bounds_max_z = 0
WHERE id = '00000000-0000-0000-0000-000000000000';
