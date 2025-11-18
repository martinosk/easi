-- Migration: Add Canvas Layout Fields to Views
-- Spec: 021_CanvasLayoutImprovements_ongoing.md
-- Description: Adds edge_type and layout_direction columns to architecture_views table
-- to support configurable edge routing and automatic layout capabilities

-- ============================================================================
-- Add columns to architecture_views table
-- ============================================================================

ALTER TABLE architecture_views
ADD COLUMN edge_type VARCHAR(20),
ADD COLUMN layout_direction VARCHAR(2);

-- ============================================================================
-- Migration complete
-- ============================================================================
