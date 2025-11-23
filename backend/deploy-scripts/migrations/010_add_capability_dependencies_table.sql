-- Migration: Add Capability Dependencies Table
-- Spec: 025_CapabilityDependencies_ongoing.md
-- Description: Creates the capability_dependencies table for modeling dependencies between capabilities
-- This migration enables tracking of how capabilities depend on each other across business domains

-- ============================================================================
-- Capability Dependencies Read Model
-- ============================================================================

-- Capability dependencies read model
CREATE TABLE IF NOT EXISTS capability_dependencies (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    source_capability_id VARCHAR(255) NOT NULL,
    target_capability_id VARCHAR(255) NOT NULL,
    dependency_type VARCHAR(50) NOT NULL CHECK (dependency_type IN ('Requires', 'Enables', 'Supports')),
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

-- Indexes for efficient dependency retrieval
CREATE INDEX IF NOT EXISTS idx_capability_dependencies_tenant_id ON capability_dependencies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_capability_dependencies_source ON capability_dependencies(tenant_id, source_capability_id);
CREATE INDEX IF NOT EXISTS idx_capability_dependencies_target ON capability_dependencies(tenant_id, target_capability_id);
CREATE INDEX IF NOT EXISTS idx_capability_dependencies_type ON capability_dependencies(tenant_id, dependency_type);
CREATE INDEX IF NOT EXISTS idx_capability_dependencies_created_at ON capability_dependencies(tenant_id, created_at);

-- Unique constraint to prevent duplicate dependencies
ALTER TABLE capability_dependencies
ADD CONSTRAINT uq_capability_dependencies_unique
UNIQUE (tenant_id, source_capability_id, target_capability_id, dependency_type);

-- ============================================================================
-- Migration complete
-- ============================================================================
