-- Migration: Add Soft Delete Support
-- Spec: 020A_DeleteOperations_Backend_ongoing.md
-- Description: Adds is_deleted and deleted_at columns to support soft deletion
-- This maintains audit trail while allowing components and relations to be removed

-- ============================================================================
-- Phase 1: Add columns to application_components table
-- ============================================================================

ALTER TABLE application_components
ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_application_components_is_deleted ON application_components(tenant_id, is_deleted);

-- ============================================================================
-- Phase 2: Add columns to component_relations table
-- ============================================================================

ALTER TABLE component_relations
ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_component_relations_is_deleted ON component_relations(tenant_id, is_deleted);

-- ============================================================================
-- Migration complete
-- ============================================================================
