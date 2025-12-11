-- Remove is_flying column from characters table

ALTER TABLE characters DROP COLUMN IF EXISTS is_flying;
