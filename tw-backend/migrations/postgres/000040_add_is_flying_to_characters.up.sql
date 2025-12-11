-- Add is_flying column to characters table for flight mode

ALTER TABLE characters ADD COLUMN IF NOT EXISTS is_flying BOOLEAN DEFAULT FALSE;
