-- Migration: Add custom_color to view_element_positions
-- Description: Adds custom_color column to support per-element custom colors in views
-- Part of Spec 045: Color Scheme Backend Support (Slice 3)

ALTER TABLE view_element_positions
ADD COLUMN IF NOT EXISTS custom_color VARCHAR(7);
