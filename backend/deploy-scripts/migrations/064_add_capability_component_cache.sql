-- Migration: 064_add_capability_component_cache
-- Purpose: Add local component cache for CapabilityMapping bounded context
-- This eliminates cross-context read model coupling by maintaining a local cache
-- populated via event subscription to ApplicationComponent lifecycle events

CREATE TABLE IF NOT EXISTS capability_component_cache (
    tenant_id VARCHAR(255) NOT NULL,
    id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX idx_capability_component_cache_tenant ON capability_component_cache(tenant_id);

-- Backfill from existing application_components table
INSERT INTO capability_component_cache (tenant_id, id, name)
SELECT tenant_id, id, name
FROM application_components
WHERE is_deleted = FALSE
ON CONFLICT (tenant_id, id) DO NOTHING;
