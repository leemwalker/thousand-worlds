-- Insert the Lobby "world" record
-- The Lobby is a special virtual space where players gather before entering actual worlds
INSERT INTO worlds (id, name, shape, metadata) VALUES
('00000000-0000-0000-0000-000000000000', 'Lobby', 'virtual', '{"description": "The Grand Lobby - a virtual gathering space between worlds"}')
ON CONFLICT (id) DO NOTHING;
