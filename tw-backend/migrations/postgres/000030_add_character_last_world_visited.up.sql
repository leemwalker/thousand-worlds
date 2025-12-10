CREATE TABLE IF NOT EXISTS characters_new (
    character_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    world_id UUID NOT NULL,
    name TEXT NOT NULL,
    role TEXT DEFAULT 'player',
    appearance JSONB,
    description TEXT,
    occupation TEXT,
    position GEOMETRY(POINT, 4326),
    position_x FLOAT DEFAULT 0,
    position_y FLOAT DEFAULT 0,
    position_z FLOAT DEFAULT 0,
    orientation_x FLOAT DEFAULT 0,
    orientation_y FLOAT DEFAULT 1,
    orientation_z FLOAT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_played TIMESTAMP WITH TIME ZONE,
    last_world_visited UUID
);

-- Note: We are altering the existing table, not creating a new one in practice, but for the migration file content:
ALTER TABLE characters ADD COLUMN IF NOT EXISTS last_world_visited UUID;
