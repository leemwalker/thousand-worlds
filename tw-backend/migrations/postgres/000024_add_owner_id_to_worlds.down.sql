-- Rollback: Remove owner_id column from worlds table
-- This migration reverses the changes made in 000024_add_owner_id_to_worlds.up.sql

-- Step 1: Restore owner_id to metadata for any worlds that still exist
UPDATE worlds 
SET metadata = jsonb_set(
    COALESCE(metadata, '{}'::jsonb), 
    '{owner_id}', 
    to_jsonb(owner_id::text)
)
WHERE owner_id IS NOT NULL;

-- Step 2: Drop index
DROP INDEX IF EXISTS idx_worlds_owner_id;

-- Step 3: Drop foreign key constraint
ALTER TABLE worlds DROP CONSTRAINT IF EXISTS fk_worlds_owner;

-- Step 4: Drop owner_id column
ALTER TABLE worlds DROP COLUMN IF EXISTS owner_id;
