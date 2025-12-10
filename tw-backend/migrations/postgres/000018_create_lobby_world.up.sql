-- Insert the Lobby "world" record
-- The Lobby is a special virtual space where players gather before entering actual worlds
-- First ensure the system user exists, then create the lobby
INSERT INTO users (user_id, email, password_hash)
VALUES ('00000000-0000-0000-0000-000000000001', 'system@mud.com', 'system_hash_placeholder')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO worlds (id, name, shape, metadata) VALUES
('00000000-0000-0000-0000-000000000000', 'Lobby', 'virtual', '{"description": "The Grand Lobby - a virtual gathering space between worlds", "owner_id": "00000000-0000-0000-0000-000000000001"}')
ON CONFLICT (id) DO NOTHING;
