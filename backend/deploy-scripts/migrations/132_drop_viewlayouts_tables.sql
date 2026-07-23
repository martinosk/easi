-- Migration: Drop obsolete ViewLayouts tables
-- Spec: 192_CanvasPositions_SingleSource_done.md
-- Description: Removes the ViewLayouts storage after migration 131 reconciled canvas
-- positions into architectureviews.view_element_positions. The bounded context has no
-- remaining consumer: the Architecture Canvas reads positions from the view payload,
-- and the Business Domain grid stopped persisting positions with spec 179. The Go
-- context (API, event handlers, repositories) is removed in the same release, so
-- nothing references these tables.

DROP TABLE IF EXISTS viewlayouts.element_positions;

DROP TABLE IF EXISTS viewlayouts.layout_containers;

DROP SCHEMA IF EXISTS viewlayouts;
