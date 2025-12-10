CREATE TABLE IF NOT EXISTS character_skills (
    character_id UUID NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    skill_name VARCHAR(50) NOT NULL,
    xp DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (character_id, skill_name)
);

CREATE INDEX IF NOT EXISTS idx_character_skills_character_id ON character_skills(character_id);
