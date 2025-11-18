-- Migration: Add Capabilities Table
-- Spec: 023_CapabilityModel_pending.md
-- Description: Creates the capabilities table for enterprise capability mapping
-- This migration adds support for hierarchical business capability modeling (L1-L4)

-- ============================================================================
-- Capability Mapping Read Model
-- ============================================================================

-- Capabilities read model
CREATE TABLE IF NOT EXISTS capabilities (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    parent_id VARCHAR(255),
    level VARCHAR(2) NOT NULL CHECK (level IN ('L1', 'L2', 'L3', 'L4')),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

-- Indexes for efficient capability retrieval
CREATE INDEX IF NOT EXISTS idx_capabilities_tenant_id ON capabilities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_capabilities_name ON capabilities(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_capabilities_level ON capabilities(tenant_id, level);
CREATE INDEX IF NOT EXISTS idx_capabilities_parent_id ON capabilities(tenant_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_capabilities_created_at ON capabilities(tenant_id, created_at);

-- Composite index for hierarchy queries
CREATE INDEX IF NOT EXISTS idx_capabilities_hierarchy ON capabilities(tenant_id, level, parent_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
