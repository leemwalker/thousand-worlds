-- Remove role column
ALTER TABLE characters DROP COLUMN IF EXISTS role;

-- Remove appearance column
ALTER TABLE characters DROP COLUMN IF EXISTS appearance;
