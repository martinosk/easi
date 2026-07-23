-- Migration: Reconcile canvas element positions into ArchitectureViews
-- Spec: 192_CanvasPositions_SingleSource_done.md
-- Description: Folds ViewLayouts positions for architecture-canvas containers into
-- architectureviews.view_element_positions, making ArchitectureViews the single owner
-- of canvas element positions. Selection is newest-wins on each store's per-element
-- updated_at. Origin entity rows are never touched. Safe to re-run: a re-run changes
-- no rows, and positions written after the first run are never overwritten because
-- their updated_at is later than any ViewLayouts timestamp.

-- ============================================================================
-- Phase 1: Newest-wins update for elements present in both stores
-- ============================================================================

UPDATE architectureviews.view_element_positions vep
SET x = ep.x,
    y = ep.y,
    updated_at = ep.updated_at
FROM viewlayouts.element_positions ep
JOIN viewlayouts.layout_containers lc
    ON lc.id = ep.container_id
    AND lc.tenant_id = ep.tenant_id
WHERE lc.context_type = 'architecture-canvas'
  AND vep.tenant_id = ep.tenant_id
  AND vep.view_id = lc.context_ref
  AND vep.element_id = ep.element_id
  AND vep.element_type IN ('component', 'capability')
  AND ep.updated_at > COALESCE(vep.updated_at, vep.created_at);

-- ============================================================================
-- Phase 2: Carry over elements positioned only in ViewLayouts
-- element_type is derived from the owning read model; rows matching neither a
-- component nor a capability (e.g. stale ids) are skipped.
-- ============================================================================

INSERT INTO architectureviews.view_element_positions
    (view_id, tenant_id, element_id, element_type, x, y, custom_color, created_at, updated_at)
SELECT
    av.id,
    ep.tenant_id,
    ep.element_id,
    CASE WHEN ac.id IS NOT NULL THEN 'component' ELSE 'capability' END,
    ep.x,
    ep.y,
    ep.custom_color,
    ep.updated_at,
    ep.updated_at
FROM viewlayouts.element_positions ep
JOIN viewlayouts.layout_containers lc
    ON lc.id = ep.container_id
    AND lc.tenant_id = ep.tenant_id
JOIN architectureviews.architecture_views av
    ON av.id = lc.context_ref
    AND av.tenant_id = lc.tenant_id
LEFT JOIN architecturemodeling.application_components ac
    ON ac.id = ep.element_id
    AND ac.tenant_id = ep.tenant_id
LEFT JOIN capabilitymapping.capabilities c
    ON c.id = ep.element_id
    AND c.tenant_id = ep.tenant_id
WHERE lc.context_type = 'architecture-canvas'
  AND av.is_deleted = false
  AND (ac.id IS NOT NULL OR c.id IS NOT NULL)
ON CONFLICT (tenant_id, view_id, element_id, element_type) DO NOTHING;

-- ============================================================================
-- Migration complete
-- ============================================================================
