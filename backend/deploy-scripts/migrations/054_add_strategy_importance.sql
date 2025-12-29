-- Migration: Add Strategy Importance Tables
-- Spec: 099_DomainCapability_StrategyAlignment_pending.md
-- Description: Creates table for domain capability strategy importance ratings

-- ============================================================================
-- Strategy Importance Read Model
-- ============================================================================

CREATE TABLE IF NOT EXISTS strategy_importance (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    business_domain_name VARCHAR(100) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    capability_name VARCHAR(200) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    pillar_name VARCHAR(100) NOT NULL,
    importance INTEGER NOT NULL CHECK (importance >= 1 AND importance <= 5),
    importance_label VARCHAR(50) NOT NULL,
    rationale VARCHAR(500),
    set_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, business_domain_id, capability_id, pillar_id)
);

-- Indexes for strategy importance queries
CREATE INDEX IF NOT EXISTS idx_si_tenant ON strategy_importance(tenant_id);
CREATE INDEX IF NOT EXISTS idx_si_domain ON strategy_importance(tenant_id, business_domain_id);
CREATE INDEX IF NOT EXISTS idx_si_capability ON strategy_importance(tenant_id, capability_id);
CREATE INDEX IF NOT EXISTS idx_si_pillar ON strategy_importance(tenant_id, pillar_id);
CREATE INDEX IF NOT EXISTS idx_si_importance ON strategy_importance(tenant_id, importance);
CREATE INDEX IF NOT EXISTS idx_si_domain_capability ON strategy_importance(tenant_id, business_domain_id, capability_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
