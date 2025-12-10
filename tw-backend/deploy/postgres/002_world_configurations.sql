-- World Configurations Table
-- Stores extracted structured world parameters from completed interviews

CREATE TABLE IF NOT EXISTS world_configurations (
    id UUID PRIMARY KEY,
    interview_id UUID NOT NULL REFERENCES world_interviews(id) ON DELETE CASCADE,
    world_id UUID REFERENCES worlds(id) ON DELETE SET NULL,
    created_by UUID NOT NULL,
    
    -- Theme fields
    theme VARCHAR(255) NOT NULL,
    tone VARCHAR(255),
    inspirations JSONB DEFAULT '[]',
    unique_aspect TEXT,
    major_conflicts JSONB DEFAULT '[]',
    
    -- Tech Level fields
    tech_level VARCHAR(100) NOT NULL,
    magic_level VARCHAR(100),
    advanced_tech TEXT,
    magic_impact TEXT,
    
    -- Geography fields
    planet_size VARCHAR(100) NOT NULL,
    climate_range VARCHAR(255),
    land_water_ratio VARCHAR(100),
    unique_features JSONB DEFAULT '[]',
    extreme_environments JSONB DEFAULT '[]',
    
    -- Culture fields
    sentient_species JSONB NOT NULL DEFAULT '[]',
    political_structure VARCHAR(255),
    cultural_values JSONB DEFAULT '[]',
    economic_system VARCHAR(255),
    religions JSONB DEFAULT '[]',
    taboos JSONB DEFAULT '[]',
    
    -- Generation parameters (derived from above)
    biome_weights JSONB DEFAULT '{}',
    resource_distribution JSONB DEFAULT '{}',
    species_start_attributes JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure at least one sentient species is defined
    CONSTRAINT at_least_one_species CHECK (jsonb_array_length(sentient_species) > 0)
);

-- Index for finding configuration by interview
CREATE INDEX IF NOT EXISTS idx_world_configurations_interview_id ON world_configurations(interview_id);

-- Index for finding configurations by world
CREATE INDEX IF NOT EXISTS idx_world_configurations_world_id ON world_configurations(world_id);

-- Index for finding configurations by creator
CREATE INDEX IF NOT EXISTS idx_world_configurations_created_by ON world_configurations(created_by);

-- Index for searching by theme
CREATE INDEX IF NOT EXISTS idx_world_configurations_theme ON world_configurations(theme);
