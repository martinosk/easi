-- Migration: Add partial unique indexes for origin relationship upsert support
-- Description: The Upsert SQL uses ON CONFLICT (tenant_id, component_id) WHERE is_deleted = FALSE
--              but no matching partial unique index existed, causing silent failures.
--              Each component can only have one origin link per type, so (tenant_id, component_id)
--              uniqueness is the correct business constraint.

CREATE UNIQUE INDEX IF NOT EXISTS idx_acquired_via_relationships_upsert
    ON acquired_via_relationships(tenant_id, component_id) WHERE is_deleted = FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_purchased_from_relationships_upsert
    ON purchased_from_relationships(tenant_id, component_id) WHERE is_deleted = FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_built_by_relationships_upsert
    ON built_by_relationships(tenant_id, component_id) WHERE is_deleted = FALSE;
