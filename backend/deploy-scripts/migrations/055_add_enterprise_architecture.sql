-- Migration: Add Enterprise Architecture Tables
-- Spec: 100_EnterpriseCapability_Groupings_pending.md
-- Description: Creates tables for enterprise capabilities, links, and strategic importance

-- ============================================================================
-- Enterprise Capabilities Read Model
-- ============================================================================

CREATE TABLE IF NOT EXISTS enterprise_capabilities (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description VARCHAR(1000),
    category VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT true,
    link_count INTEGER NOT NULL DEFAULT 0,
    domain_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

-- Indexes for enterprise capabilities queries
CREATE INDEX IF NOT EXISTS idx_ec_tenant ON enterprise_capabilities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ec_name ON enterprise_capabilities(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_ec_category ON enterprise_capabilities(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_ec_active ON enterprise_capabilities(tenant_id, active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ec_unique_name ON enterprise_capabilities(tenant_id, LOWER(name)) WHERE active = true;

-- ============================================================================
-- Enterprise Capability Links Read Model
-- ============================================================================

CREATE TABLE IF NOT EXISTS enterprise_capability_links (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    enterprise_capability_id VARCHAR(255) NOT NULL,
    domain_capability_id VARCHAR(255) NOT NULL,
    linked_by VARCHAR(255) NOT NULL,
    linked_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

-- Indexes for enterprise capability links queries
CREATE INDEX IF NOT EXISTS idx_ecl_tenant ON enterprise_capability_links(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ecl_enterprise_capability ON enterprise_capability_links(tenant_id, enterprise_capability_id);
CREATE INDEX IF NOT EXISTS idx_ecl_domain_capability ON enterprise_capability_links(tenant_id, domain_capability_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ecl_unique_domain_capability ON enterprise_capability_links(tenant_id, domain_capability_id);

-- ============================================================================
-- Enterprise Strategic Importance Read Model
-- ============================================================================

CREATE TABLE IF NOT EXISTS enterprise_strategic_importance (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    enterprise_capability_id VARCHAR(255) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    pillar_name VARCHAR(100) NOT NULL,
    importance INTEGER NOT NULL CHECK (importance >= 1 AND importance <= 5),
    rationale VARCHAR(500),
    set_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, enterprise_capability_id, pillar_id)
);

-- Indexes for enterprise strategic importance queries
CREATE INDEX IF NOT EXISTS idx_esi_tenant ON enterprise_strategic_importance(tenant_id);
CREATE INDEX IF NOT EXISTS idx_esi_enterprise_capability ON enterprise_strategic_importance(tenant_id, enterprise_capability_id);
CREATE INDEX IF NOT EXISTS idx_esi_pillar ON enterprise_strategic_importance(tenant_id, pillar_id);
CREATE INDEX IF NOT EXISTS idx_esi_importance ON enterprise_strategic_importance(tenant_id, importance);

-- ============================================================================
-- Capability Link Blocking Read Model (Pre-computed hierarchy blocking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS capability_link_blocking (
    tenant_id VARCHAR(50) NOT NULL,
    domain_capability_id VARCHAR(255) NOT NULL,
    blocked_by_capability_id VARCHAR(255) NOT NULL,
    blocked_by_enterprise_id VARCHAR(255) NOT NULL,
    blocked_by_capability_name VARCHAR(500) NOT NULL,
    blocked_by_enterprise_name VARCHAR(500) NOT NULL,
    is_ancestor BOOLEAN NOT NULL,
    PRIMARY KEY (tenant_id, domain_capability_id, blocked_by_capability_id)
);

CREATE INDEX IF NOT EXISTS idx_clb_domain_capability ON capability_link_blocking(tenant_id, domain_capability_id);
CREATE INDEX IF NOT EXISTS idx_clb_blocked_by ON capability_link_blocking(tenant_id, blocked_by_capability_id);

-- ============================================================================
-- Domain Capability Metadata Read Model (Anti-Corruption Layer)
-- Pre-computed L1 ancestors and business domain mappings for Enterprise Architecture
-- ============================================================================

CREATE TABLE IF NOT EXISTS domain_capability_metadata (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    capability_name VARCHAR(500) NOT NULL,
    capability_level VARCHAR(2) NOT NULL,
    parent_id VARCHAR(255),
    l1_capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255),
    business_domain_name VARCHAR(100),
    PRIMARY KEY (tenant_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_dcm_l1 ON domain_capability_metadata(tenant_id, l1_capability_id);
CREATE INDEX IF NOT EXISTS idx_dcm_business_domain ON domain_capability_metadata(tenant_id, business_domain_id);
CREATE INDEX IF NOT EXISTS idx_dcm_parent ON domain_capability_metadata(tenant_id, parent_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
