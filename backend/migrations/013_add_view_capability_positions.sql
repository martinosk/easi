-- Migration: Unify View Element Positions
-- Description: Creates unified view_element_positions table for both components and capabilities
-- This replaces view_component_positions with a more extensible design

-- ============================================================================
-- Phase 1: Create unified view_element_positions table
-- ============================================================================

CREATE TABLE IF NOT EXISTS view_element_positions (
    view_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    element_id VARCHAR(255) NOT NULL,
    element_type VARCHAR(20) NOT NULL CHECK (element_type IN ('component', 'capability')),
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, view_id, element_id, element_type)
);

-- ============================================================================
-- Phase 2: Create indexes for performance
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_view_element_positions_view_id ON view_element_positions(view_id);
CREATE INDEX IF NOT EXISTS idx_view_element_positions_element_id ON view_element_positions(element_id);
CREATE INDEX IF NOT EXISTS idx_view_element_positions_tenant_id ON view_element_positions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_view_element_positions_type ON view_element_positions(element_type);

-- ============================================================================
-- Phase 3: Migrate existing component positions
-- ============================================================================

INSERT INTO view_element_positions (view_id, tenant_id, element_id, element_type, x, y, created_at, updated_at)
SELECT view_id, tenant_id, component_id, 'component', x, y, created_at, updated_at
FROM view_component_positions
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Phase 4: Enable RLS and create policies
-- ============================================================================

ALTER TABLE view_element_positions ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON view_element_positions;
CREATE POLICY tenant_isolation_policy ON view_element_positions
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- ============================================================================
-- Phase 5: Drop old table
-- ============================================================================

DROP TABLE IF EXISTS view_component_positions;

-- ============================================================================
-- Migration complete
-- ============================================================================
