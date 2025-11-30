-- Migration: Cleanup Business Domain views from architecture_views
-- Spec: 060_ViewLayouts_Context_done.md
-- Description: Soft-deletes Business Domain views from architecture_views table
-- since they are now managed by the ViewLayouts bounded context.

-- ============================================================================
-- Soft-delete Business Domain views
-- These views follow the pattern: "{domainId} Domain Layout"
-- ============================================================================

UPDATE architecture_views
SET is_deleted = true,
    updated_at = NOW()
WHERE name LIKE '% Domain Layout'
  AND is_deleted = false;

-- ============================================================================
-- Migration complete
-- Business Domain views are now hidden from the Architecture Canvas view selector
-- ============================================================================
