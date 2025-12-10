-- Seed the Lobby with the central statue
-- This statue is the world creation interview NPC that players TELL to create worlds
INSERT INTO world_entities (
    world_id, 
    entity_type, 
    name, 
    description, 
    details, 
    x, y, z,
    collision, 
    locked, 
    interactable,
    metadata
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    'static',
    'statue',
    'A towering white marble statue stands in the center of the Grand Lobby, depicting an ancient figure with arms outstretched toward infinite worlds. A plaque at its base reads: "Speak to me, traveler, and I shall forge your world." Try: TELL STATUE CREATE WORLD',
    'Upon closer inspection, you notice intricate runes carved into the base, glowing faintly with ethereal light. Each rune represents a world that has been created, countless stories waiting to be told.',
    5.0, 5.0, 0.0,
    true,
    true,
    true,
    '{"type": "landmark", "color": "white", "size": "large", "glyph": "🗿", "collision_radius": 0.8}'
)
ON CONFLICT DO NOTHING;
