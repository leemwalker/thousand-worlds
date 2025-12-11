-- Set circumference for the Temperate world (half moon size)
-- Moon circumference: ~10,921 km, half = 5,460 km = 5,460,000 meters

UPDATE worlds 
SET circumference = 5460000
WHERE id = '00000000-0000-0000-0000-000000000002';
