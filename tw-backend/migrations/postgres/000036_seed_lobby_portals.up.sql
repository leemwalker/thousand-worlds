-- Seed the Lobby with portal frame objects at the walls
-- These portals show entrances to player-created worlds

-- North Portal Frame (top center wall)
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
    'structure',
    'North Portal',
    'An ornate archway frames a shimmering portal on the north wall. Through it, you glimpse distant worlds awaiting exploration.',
    'The portal frame is carved from dark obsidian, etched with silver runes that pulse with otherworldly energy. A plaque beside it reads: "Destinations may vary. Enter at your own risk."',
    5.0, 9.0, 0.0,
    true,
    true,
    true,
    '{"type": "portal_frame", "color": "purple", "glyph": "🚪", "collision_radius": 0.5}'
)
ON CONFLICT DO NOTHING;

-- East Portal Frame (right center wall)
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
    'structure',
    'East Portal',
    'An elegant portal arch stands against the east wall, its surface rippling like liquid starlight.',
    'Ancient symbols are carved into the frame, glowing faintly with arcane power. The destination beyond shifts and changes, showing glimpses of countless worlds.',
    9.0, 5.0, 0.0,
    true,
    true,
    true,
    '{"type": "portal_frame", "color": "blue", "glyph": "🚪", "collision_radius": 0.5}'
)
ON CONFLICT DO NOTHING;

-- South Portal Frame (bottom center wall, near spawn)
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
    'structure',
    'South Portal',
    'A warm, inviting portal glows on the south wall, welcoming travelers home.',
    'This portal radiates a gentle warmth. Unlike the others, it seems to remember where you came from, offering a path back to familiar places.',
    5.0, 1.0, 0.0,
    true,
    true,
    true,
    '{"type": "portal_frame", "color": "gold", "glyph": "🚪", "collision_radius": 0.5}'
)
ON CONFLICT DO NOTHING;

-- West Portal Frame (left center wall)
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
    'structure',
    'West Portal',
    'A mysterious portal shimmers on the west wall, its depths dark and unknowable.',
    'This portal feels different from the others - older, perhaps, or leading to more distant realms. Seasoned travelers speak of rare discoveries beyond its threshold.',
    1.0, 5.0, 0.0,
    true,
    true,
    true,
    '{"type": "portal_frame", "color": "green", "glyph": "🚪", "collision_radius": 0.5}'
)
ON CONFLICT DO NOTHING;
