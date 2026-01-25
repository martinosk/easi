-- Migration: Remove element_type constraint from view_element_positions
-- Description: Drops the CHECK constraint on element_type. The domain model
-- is the source of truth for valid element types, not the database.

ALTER TABLE view_element_positions
    DROP CONSTRAINT IF EXISTS view_element_positions_element_type_check;
