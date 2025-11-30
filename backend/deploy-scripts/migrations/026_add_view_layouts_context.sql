-- Migration: Add ViewLayouts Bounded Context
-- Spec: 060_ViewLayouts_Context_pending.md
-- Description: Creates layout_containers and element_positions tables for the new ViewLayouts
-- bounded context. Migrates existing Business Domain grid layouts from architecture_views.

-- ============================================================================
-- Phase 1: Create layout_containers table
-- ============================================================================

CREATE TABLE IF NOT EXISTS layout_containers (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    context_type VARCHAR(50) NOT NULL,
    context_ref VARCHAR(255) NOT NULL,
    preferences JSONB DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(tenant_id, context_type, context_ref)
);

CREATE INDEX IF NOT EXISTS idx_layout_containers_tenant_id ON layout_containers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_layout_containers_context_type ON layout_containers(context_type);
CREATE INDEX IF NOT EXISTS idx_layout_containers_context_ref ON layout_containers(context_ref);

-- ============================================================================
-- Phase 2: Create element_positions table
-- ============================================================================

CREATE TABLE IF NOT EXISTS element_positions (
    container_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    element_id VARCHAR(255) NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    width DOUBLE PRECISION,
    height DOUBLE PRECISION,
    custom_color VARCHAR(50),
    sort_order INTEGER,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, container_id, element_id)
);

CREATE INDEX IF NOT EXISTS idx_element_positions_container_id ON element_positions(container_id);
CREATE INDEX IF NOT EXISTS idx_element_positions_element_id ON element_positions(element_id);
CREATE INDEX IF NOT EXISTS idx_element_positions_tenant_id ON element_positions(tenant_id);

-- ============================================================================
-- Phase 3: Enable RLS on new tables
-- ============================================================================

ALTER TABLE layout_containers ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON layout_containers;
CREATE POLICY tenant_isolation_policy ON layout_containers
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

ALTER TABLE element_positions ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON element_positions;
CREATE POLICY tenant_isolation_policy ON element_positions
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- ============================================================================
-- Phase 4: Migrate Business Domain grid layouts
-- Business Domain views follow pattern: "{domainId} Domain Layout"
-- ============================================================================

INSERT INTO layout_containers (id, tenant_id, context_type, context_ref, preferences, version, created_at, updated_at)
SELECT
    gen_random_uuid()::text,
    av.tenant_id,
    'business-domain-grid',
    REPLACE(av.name, ' Domain Layout', ''),
    COALESCE(
        jsonb_build_object(
            'colorScheme', vp.color_scheme,
            'layoutDirection', vp.layout_direction,
            'edgeType', vp.edge_type
        ) - ARRAY(SELECT key FROM jsonb_each(jsonb_build_object(
            'colorScheme', vp.color_scheme,
            'layoutDirection', vp.layout_direction,
            'edgeType', vp.edge_type
        )) WHERE value = 'null'),
        '{}'::jsonb
    ),
    1,
    av.created_at,
    COALESCE(vp.updated_at, av.updated_at)
FROM architecture_views av
LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
WHERE av.name LIKE '% Domain Layout'
  AND av.is_deleted = false
ON CONFLICT (tenant_id, context_type, context_ref) DO NOTHING;

-- ============================================================================
-- Phase 5: Migrate element positions for Business Domain grids
-- ============================================================================

INSERT INTO element_positions (container_id, tenant_id, element_id, x, y, width, height, custom_color, sort_order, updated_at)
SELECT
    lc.id,
    vep.tenant_id,
    vep.element_id,
    vep.x,
    vep.y,
    NULL,
    NULL,
    vep.custom_color,
    NULL,
    vep.updated_at
FROM view_element_positions vep
JOIN architecture_views av ON vep.view_id = av.id AND vep.tenant_id = av.tenant_id
JOIN layout_containers lc ON lc.context_ref = REPLACE(av.name, ' Domain Layout', '')
    AND lc.tenant_id = av.tenant_id
    AND lc.context_type = 'business-domain-grid'
WHERE av.name LIKE '% Domain Layout'
  AND av.is_deleted = false
ON CONFLICT (tenant_id, container_id, element_id) DO NOTHING;

-- ============================================================================
-- Phase 6: Migrate Architecture Canvas layouts
-- All non-Business Domain views become architecture-canvas layouts
-- ============================================================================

INSERT INTO layout_containers (id, tenant_id, context_type, context_ref, preferences, version, created_at, updated_at)
SELECT
    gen_random_uuid()::text,
    av.tenant_id,
    'architecture-canvas',
    av.id,
    COALESCE(
        jsonb_build_object(
            'colorScheme', vp.color_scheme,
            'layoutDirection', vp.layout_direction,
            'edgeType', vp.edge_type
        ) - ARRAY(SELECT key FROM jsonb_each(jsonb_build_object(
            'colorScheme', vp.color_scheme,
            'layoutDirection', vp.layout_direction,
            'edgeType', vp.edge_type
        )) WHERE value = 'null'),
        '{}'::jsonb
    ),
    1,
    av.created_at,
    COALESCE(vp.updated_at, av.updated_at)
FROM architecture_views av
LEFT JOIN view_preferences vp ON av.id = vp.view_id AND av.tenant_id = vp.tenant_id
WHERE av.name NOT LIKE '% Domain Layout'
  AND av.is_deleted = false
ON CONFLICT (tenant_id, context_type, context_ref) DO NOTHING;

-- ============================================================================
-- Phase 7: Migrate element positions for Architecture Canvas
-- ============================================================================

INSERT INTO element_positions (container_id, tenant_id, element_id, x, y, width, height, custom_color, sort_order, updated_at)
SELECT
    lc.id,
    vep.tenant_id,
    vep.element_id,
    vep.x,
    vep.y,
    NULL,
    NULL,
    vep.custom_color,
    NULL,
    vep.updated_at
FROM view_element_positions vep
JOIN architecture_views av ON vep.view_id = av.id AND vep.tenant_id = av.tenant_id
JOIN layout_containers lc ON lc.context_ref = av.id
    AND lc.tenant_id = av.tenant_id
    AND lc.context_type = 'architecture-canvas'
WHERE av.name NOT LIKE '% Domain Layout'
  AND av.is_deleted = false
ON CONFLICT (tenant_id, container_id, element_id) DO NOTHING;

-- ============================================================================
-- Migration complete
-- ============================================================================
