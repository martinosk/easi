-- Migration: Add Business Domains Tables
-- Spec: 053_BusinessDomain_Aggregate_pending.md, 054_BusinessDomain_Assignment_Aggregate_pending.md, 055_BusinessDomain_ReadModels_pending.md
-- Description: Creates tables for business domain management and capability assignments
-- Supports strategic grouping of L1 capabilities into business domains

-- ============================================================================
-- Business Domain Read Models
-- ============================================================================

-- Business domains read model
CREATE TABLE IF NOT EXISTS business_domains (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    capability_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, name)
);

-- Indexes for business domains
CREATE INDEX IF NOT EXISTS idx_business_domains_tenant ON business_domains(tenant_id);
CREATE INDEX IF NOT EXISTS idx_business_domains_name ON business_domains(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_business_domains_created_at ON business_domains(tenant_id, created_at);

-- ============================================================================
-- Business Domain Assignment Read Model
-- ============================================================================

-- Domain-capability assignments read model (many-to-many)
CREATE TABLE IF NOT EXISTS domain_capability_assignments (
    assignment_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    business_domain_name VARCHAR(100) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    capability_code VARCHAR(50) NOT NULL,
    capability_name VARCHAR(200) NOT NULL,
    capability_level VARCHAR(2) NOT NULL CHECK (capability_level = 'L1'),
    assigned_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, assignment_id),
    UNIQUE (tenant_id, business_domain_id, capability_id)
);

-- Indexes for assignment queries
CREATE INDEX IF NOT EXISTS idx_dca_tenant ON domain_capability_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dca_domain ON domain_capability_assignments(tenant_id, business_domain_id);
CREATE INDEX IF NOT EXISTS idx_dca_capability ON domain_capability_assignments(tenant_id, capability_id);
CREATE INDEX IF NOT EXISTS idx_dca_assigned_at ON domain_capability_assignments(tenant_id, assigned_at);

-- ============================================================================
-- Domain Composition View (for visualization)
-- ============================================================================

-- Domain composition with full capability hierarchy
CREATE TABLE IF NOT EXISTS domain_composition_view (
    tenant_id VARCHAR(50) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    l1_capability_id VARCHAR(255) NOT NULL,
    l1_capability_code VARCHAR(50) NOT NULL,
    l1_capability_name VARCHAR(200) NOT NULL,
    child_capabilities JSONB,
    realizing_systems JSONB,
    assigned_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, business_domain_id, l1_capability_id)
);

-- Indexes for composition queries
CREATE INDEX IF NOT EXISTS idx_dc_tenant ON domain_composition_view(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dc_domain ON domain_composition_view(tenant_id, business_domain_id);
CREATE INDEX IF NOT EXISTS idx_dc_capability ON domain_composition_view(tenant_id, l1_capability_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
