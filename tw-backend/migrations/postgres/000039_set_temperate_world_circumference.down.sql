-- Revert Temperate world circumference to NULL

UPDATE worlds 
SET circumference = NULL
WHERE id = '00000000-0000-0000-0000-000000000002';
