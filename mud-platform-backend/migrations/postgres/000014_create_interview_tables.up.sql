CREATE TABLE IF NOT EXISTS world_interviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
    status interview_status NOT NULL DEFAULT 'not_started',
    current_question_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Create enum type if it doesn't exist
DO $$ BEGIN
    CREATE TYPE interview_status AS ENUM ('not_started', 'in_progress', 'complete', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE INDEX IF NOT EXISTS idx_interviews_user ON world_interviews(user_id);

CREATE TABLE IF NOT EXISTS world_configurations (
    id UUID PRIMARY KEY,
    interview_id UUID NOT NULL REFERENCES world_interviews(id) ON DELETE CASCADE,
    world_id UUID REFERENCES worlds(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    
    -- Theme
    theme TEXT,
    tone TEXT,
    inspirations JSONB,
    unique_aspect TEXT,
    major_conflicts JSONB,
    
    -- Tech Level
    tech_level VARCHAR(50),
    magic_level VARCHAR(50),
    advanced_tech TEXT,
    magic_impact TEXT,
    
    -- Geography
    planet_size VARCHAR(50),
    climate_range VARCHAR(50),
    land_water_ratio VARCHAR(50),
    unique_features JSONB,
    extreme_environments JSONB,
    
    -- Culture
    sentient_species JSONB,
    political_structure TEXT,
    cultural_values JSONB,
    economic_system TEXT,
    religions JSONB,
    taboos JSONB,
    
    -- Generation Parameters
    biome_weights JSONB,
    resource_distribution JSONB,
    species_start_attributes JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_world_configs_interview ON world_configurations(interview_id);
CREATE INDEX idx_world_configs_created_by ON world_configurations(created_by);
