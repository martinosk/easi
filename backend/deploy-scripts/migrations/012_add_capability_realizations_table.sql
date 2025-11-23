-- Migration: Add Capability Realizations Table
-- Spec: 026_CapabilitySystemRealization_ongoing.md
-- Description: Creates the capability_realizations table for linking capabilities to application components
-- This migration enables tracking which systems technically realize each business capability

-- ============================================================================
-- Add Multi-Tenancy Support to Application Components
-- ============================================================================
-- Note: Foreign keys were dropped in migration 011

-- Add soft delete columns to application_components
ALTER TABLE application_components
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Add soft delete columns to component_relations
ALTER TABLE component_relations
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Update application_components primary key to composite (tenant_id, id)
ALTER TABLE application_components DROP CONSTRAINT IF EXISTS application_components_pkey;
ALTER TABLE application_components ADD PRIMARY KEY (tenant_id, id);

-- Update component_relations primary key to composite (tenant_id, id)
ALTER TABLE component_relations DROP CONSTRAINT IF EXISTS component_relations_pkey;
ALTER TABLE component_relations ADD PRIMARY KEY (tenant_id, id);

-- Update architecture_views primary key to composite (tenant_id, id)
ALTER TABLE architecture_views DROP CONSTRAINT IF EXISTS architecture_views_pkey;
ALTER TABLE architecture_views ADD PRIMARY KEY (tenant_id, id);

-- Update indexes for application_components
DROP INDEX IF EXISTS idx_application_components_name;
DROP INDEX IF EXISTS idx_application_components_created_at;
CREATE INDEX idx_application_components_tenant_id ON application_components(tenant_id);
CREATE INDEX idx_application_components_name ON application_components(tenant_id, name);
CREATE INDEX idx_application_components_created_at ON application_components(tenant_id, created_at);

-- Update indexes for component_relations
DROP INDEX IF EXISTS idx_component_relations_source;
DROP INDEX IF EXISTS idx_component_relations_target;
DROP INDEX IF EXISTS idx_component_relations_type;
DROP INDEX IF EXISTS idx_component_relations_created_at;
CREATE INDEX idx_component_relations_tenant_id ON component_relations(tenant_id);
CREATE INDEX idx_component_relations_source ON component_relations(tenant_id, source_component_id);
CREATE INDEX idx_component_relations_target ON component_relations(tenant_id, target_component_id);
CREATE INDEX idx_component_relations_type ON component_relations(tenant_id, relation_type);
CREATE INDEX idx_component_relations_created_at ON component_relations(tenant_id, created_at);

-- Update indexes for architecture_views
DROP INDEX IF EXISTS idx_architecture_views_name;
DROP INDEX IF EXISTS idx_architecture_views_created_at;
DROP INDEX IF EXISTS idx_architecture_views_is_default;
CREATE INDEX idx_architecture_views_tenant_id ON architecture_views(tenant_id);
CREATE INDEX idx_architecture_views_name ON architecture_views(tenant_id, name);
CREATE INDEX idx_architecture_views_created_at ON architecture_views(tenant_id, created_at);
CREATE INDEX idx_architecture_views_is_default ON architecture_views(tenant_id, is_default);

-- Update indexes for view_component_positions
DROP INDEX IF EXISTS idx_view_component_positions_view_id;
CREATE INDEX idx_view_component_positions_view_id ON view_component_positions(tenant_id, view_id);

-- ============================================================================
-- Capability Realizations Read Model
-- ============================================================================

-- Capability realizations read model
CREATE TABLE IF NOT EXISTS capability_realizations (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    realization_level VARCHAR(50) NOT NULL CHECK (realization_level IN ('Full', 'Partial', 'Planned')),
    notes TEXT,
    linked_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

-- Indexes for efficient realization retrieval
CREATE INDEX IF NOT EXISTS idx_capability_realizations_tenant_id ON capability_realizations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_capability_realizations_capability_id ON capability_realizations(tenant_id, capability_id);
CREATE INDEX IF NOT EXISTS idx_capability_realizations_component_id ON capability_realizations(tenant_id, component_id);
CREATE INDEX IF NOT EXISTS idx_capability_realizations_level ON capability_realizations(tenant_id, realization_level);
CREATE INDEX IF NOT EXISTS idx_capability_realizations_linked_at ON capability_realizations(tenant_id, linked_at);

-- Unique constraint to prevent duplicate links
ALTER TABLE capability_realizations
ADD CONSTRAINT uq_capability_realizations_unique
UNIQUE (tenant_id, capability_id, component_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
