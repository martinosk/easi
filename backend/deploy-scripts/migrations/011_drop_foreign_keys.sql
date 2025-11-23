-- Migration: Drop Foreign Key Constraints
-- Description: Removes all foreign key constraints from read models
-- Rationale: In event-sourced systems, referential integrity is maintained by the domain model,
--            not by database constraints. Foreign keys cause coupling between bounded contexts
--            and complicate migrations.

-- Drop FK from view_component_positions to architecture_views
ALTER TABLE view_component_positions
DROP CONSTRAINT IF EXISTS view_component_positions_view_id_fkey;

-- Drop FK from view_preferences to architecture_views
ALTER TABLE view_preferences
DROP CONSTRAINT IF EXISTS view_preferences_view_id_fkey;

-- ============================================================================
-- Migration complete
-- ============================================================================
