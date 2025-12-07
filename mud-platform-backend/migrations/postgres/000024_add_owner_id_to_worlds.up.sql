-- Add owner_id column to worlds table
-- This migration adds a dedicated owner_id column to replace the JSON metadata approach

-- Step 1: Add owner_id column (nullable initially for migration)
ALTER TABLE worlds ADD COLUMN owner_id UUID;

-- Step 2: Migrate existing owner_id values from metadata to new column
-- This handles worlds that were created with metadata['owner_id']
UPDATE worlds 
SET owner_id = (metadata->>'owner_id')::UUID 
WHERE metadata->>'owner_id' IS NOT NULL;

-- Step 3: Add NOT NULL constraint after migration
-- If any worlds still have NULL owner_id, this will fail
ALTER TABLE worlds ALTER COLUMN owner_id SET NOT NULL;

-- Step 4: Add foreign key constraint to ensure referential integrity
ALTER TABLE worlds 
ADD CONSTRAINT fk_worlds_owner 
FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- Step 5: Add index for performance on owner-based queries
CREATE INDEX idx_worlds_owner_id ON worlds(owner_id);

-- Step 6: Clean up metadata by removing owner_id (optional but recommended)
UPDATE worlds SET metadata = metadata - 'owner_id';
