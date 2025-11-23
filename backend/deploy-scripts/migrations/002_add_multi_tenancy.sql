-- Migration: Add Multi-Tenancy Support
-- Spec: 013_MultiTenancy_pending.md
-- Description: Adds tenant_id columns to all tables for tenant isolation
-- This migration modifies existing tables to add multi-tenancy support

-- ============================================================================
-- Phase 1: Add tenant_id columns to event store tables
-- ============================================================================

-- Add tenant_id to events table (if not exists for idempotency)
ALTER TABLE events
ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50);

-- Add tenant_id to snapshots table (if not exists for idempotency)
ALTER TABLE snapshots
ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50);

-- ============================================================================
-- Phase 2: Add tenant_id columns to read model tables
-- ============================================================================

-- Add tenant_id to application_components
ALTER TABLE application_components
ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50);

-- Add tenant_id to component_relations
ALTER TABLE component_relations
ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50);

-- Add tenant_id to architecture_views
ALTER TABLE architecture_views
ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50);

-- Add tenant_id to view_component_positions
ALTER TABLE view_component_positions
ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50);

-- ============================================================================
-- Phase 3: Backfill with default tenant
-- ============================================================================

-- Backfill events with default tenant
UPDATE events
SET tenant_id = 'default'
WHERE tenant_id IS NULL;

-- Backfill snapshots with default tenant
UPDATE snapshots
SET tenant_id = 'default'
WHERE tenant_id IS NULL;

-- Backfill application_components with default tenant
UPDATE application_components
SET tenant_id = 'default'
WHERE tenant_id IS NULL;

-- Backfill component_relations with default tenant
UPDATE component_relations
SET tenant_id = 'default'
WHERE tenant_id IS NULL;

-- Backfill architecture_views with default tenant
UPDATE architecture_views
SET tenant_id = 'default'
WHERE tenant_id IS NULL;

-- Backfill view_component_positions with default tenant
UPDATE view_component_positions
SET tenant_id = 'default'
WHERE tenant_id IS NULL;

-- ============================================================================
-- Phase 4: Add NOT NULL constraints
-- ============================================================================

ALTER TABLE events
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE snapshots
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE application_components
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE component_relations
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE architecture_views
ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE view_component_positions
ALTER COLUMN tenant_id SET NOT NULL;

-- ============================================================================
-- Phase 5: Create indexes for tenant isolation
-- ============================================================================

-- Event store indexes
CREATE INDEX IF NOT EXISTS idx_events_tenant_id ON events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_events_tenant_aggregate ON events(tenant_id, aggregate_id);
CREATE INDEX IF NOT EXISTS idx_events_tenant_type ON events(tenant_id, event_type);
CREATE INDEX IF NOT EXISTS idx_events_tenant_occurred ON events(tenant_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_snapshots_tenant_id ON snapshots(tenant_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_tenant_aggregate ON snapshots(tenant_id, aggregate_id);

-- Read model indexes
CREATE INDEX IF NOT EXISTS idx_application_components_tenant ON application_components(tenant_id);
CREATE INDEX IF NOT EXISTS idx_component_relations_tenant ON component_relations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_architecture_views_tenant ON architecture_views(tenant_id);
CREATE INDEX IF NOT EXISTS idx_view_component_positions_tenant ON view_component_positions(tenant_id);

-- ============================================================================
-- Phase 6: Update unique constraints to include tenant_id
-- ============================================================================

-- Drop old unique constraint on events if it exists
ALTER TABLE events
DROP CONSTRAINT IF EXISTS events_aggregate_id_version_key;

-- Add new unique constraint including tenant_id
ALTER TABLE events
ADD CONSTRAINT events_tenant_aggregate_version_key
UNIQUE (tenant_id, aggregate_id, version);

-- Drop old unique constraint on snapshots if it exists
ALTER TABLE snapshots
DROP CONSTRAINT IF EXISTS snapshots_aggregate_id_version_key;

-- Add new unique constraint including tenant_id
ALTER TABLE snapshots
ADD CONSTRAINT snapshots_tenant_aggregate_version_key
UNIQUE (tenant_id, aggregate_id, version);

-- Drop old primary key on view_component_positions if needed and recreate with tenant
-- Note: This is complex, so we'll just add a unique constraint for tenant isolation
ALTER TABLE view_component_positions
DROP CONSTRAINT IF EXISTS view_component_positions_pkey;

ALTER TABLE view_component_positions
ADD CONSTRAINT view_component_positions_pkey
PRIMARY KEY (tenant_id, view_id, component_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
