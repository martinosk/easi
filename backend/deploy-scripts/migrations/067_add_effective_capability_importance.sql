-- Migration: Add Effective Capability Importance Table
-- Spec: 110_Hierarchical_Strategic_Rating_Evaluation_pending.md
-- Description: Pre-computed materialized view of resolved importance for every capability
-- that has an effective rating (direct or inherited).

-- ============================================================================
-- Effective Capability Importance Read Model
-- ============================================================================

CREATE TABLE IF NOT EXISTS effective_capability_importance (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    effective_importance INTEGER NOT NULL CHECK (effective_importance >= 1 AND effective_importance <= 5),
    importance_label VARCHAR(50) NOT NULL,
    source_capability_id VARCHAR(255) NOT NULL,
    source_capability_name VARCHAR(200) NOT NULL,
    is_inherited BOOLEAN NOT NULL DEFAULT FALSE,
    computed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, capability_id, pillar_id, business_domain_id)
);

CREATE INDEX IF NOT EXISTS idx_eci_tenant ON effective_capability_importance(tenant_id);
CREATE INDEX IF NOT EXISTS idx_eci_capability ON effective_capability_importance(tenant_id, capability_id);
CREATE INDEX IF NOT EXISTS idx_eci_pillar ON effective_capability_importance(tenant_id, pillar_id);
CREATE INDEX IF NOT EXISTS idx_eci_domain ON effective_capability_importance(tenant_id, business_domain_id);
CREATE INDEX IF NOT EXISTS idx_eci_source ON effective_capability_importance(tenant_id, source_capability_id);
CREATE INDEX IF NOT EXISTS idx_eci_pillar_domain ON effective_capability_importance(tenant_id, pillar_id, business_domain_id);

-- ============================================================================
-- Row Level Security
-- ============================================================================

ALTER TABLE effective_capability_importance ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON effective_capability_importance;
CREATE POLICY tenant_isolation_policy ON effective_capability_importance
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- ============================================================================
-- Migration complete
-- ============================================================================
