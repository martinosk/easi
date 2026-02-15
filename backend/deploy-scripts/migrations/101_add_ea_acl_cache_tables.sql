ALTER TABLE domain_capability_metadata ADD COLUMN IF NOT EXISTS maturity_value INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS ea_realization_cache (
    tenant_id VARCHAR(255) NOT NULL,
    realization_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    component_name VARCHAR(500) NOT NULL,
    origin VARCHAR(50) NOT NULL,
    PRIMARY KEY (tenant_id, realization_id)
);

ALTER TABLE ea_realization_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_realization_cache_tenant_isolation ON ea_realization_cache
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE POLICY ea_realization_cache_tenant_insert ON ea_realization_cache
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE INDEX idx_ea_realization_cache_tenant ON ea_realization_cache(tenant_id);
CREATE INDEX idx_ea_realization_cache_capability ON ea_realization_cache(tenant_id, capability_id);
CREATE INDEX idx_ea_realization_cache_component ON ea_realization_cache(tenant_id, component_id);

CREATE TABLE IF NOT EXISTS ea_importance_cache (
    tenant_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    effective_importance INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, capability_id, business_domain_id, pillar_id)
);

ALTER TABLE ea_importance_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_importance_cache_tenant_isolation ON ea_importance_cache
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE POLICY ea_importance_cache_tenant_insert ON ea_importance_cache
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE INDEX idx_ea_importance_cache_tenant ON ea_importance_cache(tenant_id);

CREATE TABLE IF NOT EXISTS ea_fit_score_cache (
    tenant_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,
    rationale TEXT,
    PRIMARY KEY (tenant_id, component_id, pillar_id)
);

ALTER TABLE ea_fit_score_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_fit_score_cache_tenant_isolation ON ea_fit_score_cache
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE POLICY ea_fit_score_cache_tenant_insert ON ea_fit_score_cache
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::TEXT);

CREATE INDEX idx_ea_fit_score_cache_tenant ON ea_fit_score_cache(tenant_id);

-- Backfill ea_realization_cache from capability_realizations
INSERT INTO ea_realization_cache (tenant_id, realization_id, capability_id, component_id, component_name, origin)
SELECT cr.tenant_id, cr.id, cr.capability_id, cr.component_id, cr.component_name, cr.origin
FROM capability_realizations cr
ON CONFLICT (tenant_id, realization_id) DO NOTHING;

-- Backfill ea_importance_cache from effective_capability_importance
INSERT INTO ea_importance_cache (tenant_id, capability_id, business_domain_id, pillar_id, effective_importance)
SELECT eci.tenant_id, eci.capability_id, eci.business_domain_id, eci.pillar_id, eci.effective_importance
FROM effective_capability_importance eci
ON CONFLICT (tenant_id, capability_id, business_domain_id, pillar_id) DO NOTHING;

-- Backfill ea_fit_score_cache from application_fit_scores
INSERT INTO ea_fit_score_cache (tenant_id, component_id, pillar_id, score, rationale)
SELECT afs.tenant_id, afs.component_id, afs.pillar_id, afs.score, afs.rationale
FROM application_fit_scores afs
ON CONFLICT (tenant_id, component_id, pillar_id) DO NOTHING;

-- Backfill maturity_value into domain_capability_metadata from capabilities
UPDATE domain_capability_metadata dcm
SET maturity_value = c.maturity_value
FROM capabilities c
WHERE dcm.capability_id = c.id AND dcm.tenant_id = c.tenant_id;
