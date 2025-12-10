-- Revert lobby bounds back to virtual shape
UPDATE worlds
SET shape = 'virtual',
    bounds_min_x = NULL,
    bounds_min_y = NULL,
    bounds_min_z = NULL,
    bounds_max_x = NULL,
    bounds_max_y = NULL,
    bounds_max_z = NULL
WHERE id = '00000000-0000-0000-0000-000000000000';
