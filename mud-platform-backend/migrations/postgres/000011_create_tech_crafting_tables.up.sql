-- Tech Trees
CREATE TABLE IF NOT EXISTS tech_trees (
    tree_id UUID PRIMARY KEY,
    world_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    tech_level VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tech_trees_world ON tech_trees(world_id);

CREATE TABLE IF NOT EXISTS tech_nodes (
    node_id UUID PRIMARY KEY,
    tree_id UUID REFERENCES tech_trees(tree_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tech_level VARCHAR(20) NOT NULL,
    tier INT NOT NULL,
    research_time_seconds BIGINT NOT NULL,
    icon_path VARCHAR(255),
    metadata JSONB, -- Stores prerequisites, costs, unlocks
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tech_nodes_tree ON tech_nodes(tree_id);
CREATE INDEX idx_tech_nodes_level ON tech_nodes(tech_level);

-- Recipes
CREATE TABLE IF NOT EXISTS recipes (
    recipe_id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    tech_node_id UUID, -- Optional link to tech node
    required_skill VARCHAR(50) NOT NULL,
    min_skill_level INT NOT NULL,
    crafting_time_seconds BIGINT NOT NULL,
    base_value INT NOT NULL,
    difficulty VARCHAR(20) NOT NULL,
    data JSONB, -- Stores ingredients, tools, stations, outputs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recipes_category ON recipes(category);
CREATE INDEX idx_recipes_tech_node ON recipes(tech_node_id);
CREATE INDEX idx_recipes_skill ON recipes(required_skill);

-- Knowledge Tracking
CREATE TABLE IF NOT EXISTS unlocked_tech (
    entity_id UUID NOT NULL, -- Player or NPC
    node_id UUID REFERENCES tech_nodes(node_id) ON DELETE CASCADE,
    unlocked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entity_id, node_id)
);

CREATE TABLE IF NOT EXISTS recipe_knowledge (
    entity_id UUID NOT NULL,
    recipe_id UUID REFERENCES recipes(recipe_id) ON DELETE CASCADE,
    proficiency FLOAT NOT NULL DEFAULT 0.0,
    times_used INT NOT NULL DEFAULT 0,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    source VARCHAR(50) NOT NULL,
    teacher_id UUID,
    PRIMARY KEY (entity_id, recipe_id)
);
