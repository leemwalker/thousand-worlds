-- Add position columns to characters table
ALTER TABLE characters
ADD COLUMN IF NOT EXISTS position_x DOUBLE PRECISION DEFAULT 5.0,
ADD COLUMN IF NOT EXISTS position_y DOUBLE PRECISION DEFAULT 500.0,
ADD COLUMN IF NOT EXISTS position_z DOUBLE PRECISION DEFAULT 0.0;

-- Create index for spatial queries
CREATE INDEX IF NOT EXISTS idx_character_position
ON characters (world_id, position_x, position_y);

-- Add dimensions to worlds table
ALTER TABLE worlds
ADD COLUMN IF NOT EXISTS circumference DOUBLE PRECISION DEFAULT 10000.0;
-- Radius already exists in the worlds table from initial migration

-- Create portals table
CREATE TABLE IF NOT EXISTS portals (
    portal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID REFERENCES worlds(id),
    location_x DOUBLE PRECISION NOT NULL,
    location_y DOUBLE PRECISION NOT NULL,
    side VARCHAR(10) CHECK (side IN ('east', 'west')),
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_portal_location ON portals (location_x, location_y);
