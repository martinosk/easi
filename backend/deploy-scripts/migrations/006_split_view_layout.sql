-- Migration: Split View Layout from Domain
-- Description: Separates component positions (presentation concern) from view membership (domain concern)
-- Strategy: Keep existing view_component_positions table structure but change its usage pattern
-- The table now serves as direct persistence for layout data, not as an event-sourced projection

-- ============================================================================
-- No structural changes needed
-- ============================================================================

-- The view_component_positions table already has the correct structure:
-- - view_id: which view the positions belong to
-- - component_id: which component is positioned
-- - x, y: position coordinates (presentation data)
-- - created_at, updated_at: tracking timestamps

-- Usage pattern changes (handled in application layer):
-- BEFORE: Positions populated by ComponentAddedToView and ComponentPositionUpdated events
-- AFTER: Positions updated directly via ViewLayoutRepository (no events for positions)

-- ============================================================================
-- Add index for better performance on direct writes
-- ============================================================================

-- Index for component lookups (useful when component is deleted)
CREATE INDEX IF NOT EXISTS idx_view_component_positions_component_id ON view_component_positions(component_id);

-- ============================================================================
-- Data migration: No changes needed
-- ============================================================================

-- Existing position data remains valid - it's already in the correct table
-- The difference is in how new data will be written (direct writes instead of event projections)

-- ============================================================================
-- Migration complete
-- ============================================================================
