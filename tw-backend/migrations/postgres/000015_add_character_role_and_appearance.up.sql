-- Add role column to characters table
ALTER TABLE characters ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'player';

-- Add appearance column to characters table
ALTER TABLE characters ADD COLUMN IF NOT EXISTS appearance JSONB;

-- Create index for role
CREATE INDEX IF NOT EXISTS idx_characters_role ON characters(role);
