-- Migration: Enable Row-Level Security (RLS) for Tenant Isolation
-- Spec: 017_PostgreSQLRLSImplementation_pending.md
-- Description: Implements PostgreSQL RLS for database-level tenant isolation
-- This provides defense-in-depth security at the database layer
--
-- NOTE: Database users (easi_app, easi_admin) are provisioned by deploy-scripts/provision-db-users.sh
-- which runs on every deployment to allow secure credential rotation.

-- ============================================================================
-- Phase 1: Enable RLS on all tenant-scoped tables
-- ============================================================================

ALTER TABLE events ENABLE ROW LEVEL SECURITY;
ALTER TABLE snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE application_components ENABLE ROW LEVEL SECURITY;
ALTER TABLE component_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE architecture_views ENABLE ROW LEVEL SECURITY;
ALTER TABLE view_component_positions ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- Phase 2: Create RLS policies for tenant isolation
-- ============================================================================

-- Events table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON events;
CREATE POLICY tenant_isolation_policy ON events
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Snapshots table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON snapshots;
CREATE POLICY tenant_isolation_policy ON snapshots
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Application components table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON application_components;
CREATE POLICY tenant_isolation_policy ON application_components
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Component relations table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON component_relations;
CREATE POLICY tenant_isolation_policy ON component_relations
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Architecture views table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON architecture_views;
CREATE POLICY tenant_isolation_policy ON architecture_views
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- View component positions table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON view_component_positions;
CREATE POLICY tenant_isolation_policy ON view_component_positions
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- ============================================================================
-- Phase 3: Create helper functions for tenant context management
-- ============================================================================

-- Function to set tenant context
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id VARCHAR(50))
RETURNS VOID AS $$
BEGIN
    -- Validate tenant ID format (basic validation)
    IF p_tenant_id IS NULL OR LENGTH(p_tenant_id) < 3 OR LENGTH(p_tenant_id) > 50 THEN
        RAISE EXCEPTION 'Invalid tenant ID: %', p_tenant_id;
    END IF;

    -- Set the session variable
    PERFORM set_config('app.current_tenant', p_tenant_id, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get current tenant
CREATE OR REPLACE FUNCTION get_current_tenant()
RETURNS VARCHAR(50) AS $$
BEGIN
    RETURN current_setting('app.current_tenant', true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- Migration complete
-- ============================================================================

-- Note: After this migration, the application should use the easi_app user
-- and set the tenant context via: SELECT set_tenant_context('tenant-id');
-- or directly via: SET app.current_tenant = 'tenant-id';
--
-- Function grants (GRANT EXECUTE ON FUNCTION ...) are handled by
-- deploy-scripts/provision-db-users.sh to ensure they run after user creation.
