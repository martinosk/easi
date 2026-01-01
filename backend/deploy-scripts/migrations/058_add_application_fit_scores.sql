-- Migration: Add Application Fit Scores Tables
-- Spec: 103_Strategic_Fit_Analysis_pending.md
-- Description: Creates table for application component strategic fit scores

-- ============================================================================
-- Application Fit Scores Read Model
-- ============================================================================

CREATE TABLE IF NOT EXISTS application_fit_scores (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    component_name VARCHAR(200) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    pillar_name VARCHAR(100) NOT NULL,
    score INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
    score_label VARCHAR(50) NOT NULL,
    rationale VARCHAR(500),
    scored_at TIMESTAMP NOT NULL,
    scored_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, component_id, pillar_id)
);

-- Indexes for application fit scores queries
CREATE INDEX IF NOT EXISTS idx_afs_tenant ON application_fit_scores(tenant_id);
CREATE INDEX IF NOT EXISTS idx_afs_component ON application_fit_scores(tenant_id, component_id);
CREATE INDEX IF NOT EXISTS idx_afs_pillar ON application_fit_scores(tenant_id, pillar_id);
CREATE INDEX IF NOT EXISTS idx_afs_score ON application_fit_scores(tenant_id, score);
CREATE INDEX IF NOT EXISTS idx_afs_component_pillar ON application_fit_scores(tenant_id, component_id, pillar_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
