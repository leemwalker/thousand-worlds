-- Seed example worlds for testing

-- 1. Earth-sized planet
INSERT INTO worlds (id, name, shape, radius, metadata) VALUES
('00000000-0000-0000-0000-000000000001', 'Earth', 'sphere', 6371000, '{"description": "Earth-sized planet for testing"}');

-- 2. Test continent (1km x 1km)
INSERT INTO worlds (id, name, shape, bounds_min_x, bounds_max_x, bounds_min_y, bounds_max_y, bounds_min_z, bounds_max_z, metadata) VALUES
('00000000-0000-0000-0000-000000000002', 'Test Continent', 'cube', -500, 500, -500, 500, -100, 1000, '{"description": "1km x 1km test continent"}');

-- 3. Mansion interior
INSERT INTO worlds (id, name, shape, bounds_min_x, bounds_max_x, bounds_min_y, bounds_max_y, bounds_min_z, bounds_max_z, metadata) VALUES
('00000000-0000-0000-0000-000000000003', 'Mansion', 'cube', 0, 100, 0, 80, 0, 20, '{"description": "Large building interior"}');
