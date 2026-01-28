-- Migration: Add strategy pillar cache tables for bounded context autonomy
-- Description: Creates local cache tables for strategy pillars in CapabilityMapping and EnterpriseArchitecture contexts
--              to maintain event-driven architecture and avoid cross-context database queries

-- CapabilityMapping context cache
CREATE TABLE IF NOT EXISTS cm_strategy_pillar_cache (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    fit_scoring_enabled BOOLEAN NOT NULL DEFAULT false,
    fit_criteria TEXT,
    fit_type VARCHAR(50),
    PRIMARY KEY (id, tenant_id)
);

-- Row Level Security for CapabilityMapping cache
ALTER TABLE cm_strategy_pillar_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY cm_strategy_pillar_cache_tenant_isolation ON cm_strategy_pillar_cache
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE POLICY cm_strategy_pillar_cache_tenant_insert ON cm_strategy_pillar_cache
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

-- EnterpriseArchitecture context cache
CREATE TABLE IF NOT EXISTS ea_strategy_pillar_cache (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    fit_scoring_enabled BOOLEAN NOT NULL DEFAULT false,
    fit_criteria TEXT,
    fit_type VARCHAR(50),
    PRIMARY KEY (id, tenant_id)
);

-- Row Level Security for EnterpriseArchitecture cache
ALTER TABLE ea_strategy_pillar_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_strategy_pillar_cache_tenant_isolation ON ea_strategy_pillar_cache
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE POLICY ea_strategy_pillar_cache_tenant_insert ON ea_strategy_pillar_cache
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

-- Indexes for performance
CREATE INDEX idx_cm_strategy_pillar_cache_tenant ON cm_strategy_pillar_cache(tenant_id);
CREATE INDEX idx_cm_strategy_pillar_cache_active ON cm_strategy_pillar_cache(tenant_id, active) WHERE active = true;

CREATE INDEX idx_ea_strategy_pillar_cache_tenant ON ea_strategy_pillar_cache(tenant_id);
CREATE INDEX idx_ea_strategy_pillar_cache_active ON ea_strategy_pillar_cache(tenant_id, active) WHERE active = true;
