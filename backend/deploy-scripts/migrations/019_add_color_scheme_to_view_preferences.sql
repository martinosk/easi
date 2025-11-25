-- Migration: Add color_scheme to view_preferences
-- Description: Adds color_scheme column to support per-view color scheme selection
-- Part of Spec 045: Color Scheme Backend Support

ALTER TABLE view_preferences
ADD COLUMN IF NOT EXISTS color_scheme VARCHAR(20);
