-- Migration: Add Capability Metadata
-- Spec: 024_CapabilityMetadata_pending.md
-- Description: Adds metadata fields to capabilities table for strategic alignment, maturity, ownership, and enrichment
-- This migration extends the capabilities table with metadata support

-- ============================================================================
-- Add metadata columns to capabilities table
-- ============================================================================

-- Strategic alignment
ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS strategy_pillar VARCHAR(50) CHECK (strategy_pillar IN ('', 'AlwaysOn', 'Grow', 'Transform'));

ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS pillar_weight INT CHECK (pillar_weight >= 0 AND pillar_weight <= 100);

-- Maturity tracking
ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS maturity_level VARCHAR(50) NOT NULL DEFAULT 'Genesis';

-- Ownership model
ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS ownership_model VARCHAR(50) CHECK (ownership_model IN ('', 'TribeOwned', 'TeamOwned', 'Shared', 'EnterpriseService'));

ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS primary_owner TEXT;

ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS ea_owner TEXT;

-- Lifecycle status
ALTER TABLE capabilities
ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'Active' CHECK (status IN ('Active', 'Planned', 'Deprecated'));

-- ============================================================================
-- Create supporting tables for experts and tags
-- ============================================================================

-- Capability experts table
CREATE TABLE IF NOT EXISTS capability_experts (
    capability_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    expert_name VARCHAR(500) NOT NULL,
    expert_role VARCHAR(500) NOT NULL,
    contact_info VARCHAR(500) NOT NULL,
    added_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, capability_id, expert_name)
);

CREATE INDEX IF NOT EXISTS idx_capability_experts_capability ON capability_experts(tenant_id, capability_id);

-- Capability tags table
CREATE TABLE IF NOT EXISTS capability_tags (
    capability_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    tag VARCHAR(100) NOT NULL,
    added_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, capability_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_capability_tags_capability ON capability_tags(tenant_id, capability_id);
CREATE INDEX IF NOT EXISTS idx_capability_tags_tag ON capability_tags(tenant_id, tag);

-- ============================================================================
-- Create indexes for filtering by metadata
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_capabilities_strategy_pillar ON capabilities(tenant_id, strategy_pillar) WHERE strategy_pillar IS NOT NULL AND strategy_pillar != '';
CREATE INDEX IF NOT EXISTS idx_capabilities_maturity_level ON capabilities(tenant_id, maturity_level);
CREATE INDEX IF NOT EXISTS idx_capabilities_ownership_model ON capabilities(tenant_id, ownership_model) WHERE ownership_model IS NOT NULL AND ownership_model != '';
CREATE INDEX IF NOT EXISTS idx_capabilities_status ON capabilities(tenant_id, status);

-- ============================================================================
-- Migration complete
-- ============================================================================
