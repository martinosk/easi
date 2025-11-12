-- Migration: Enable Row-Level Security (RLS) for Tenant Isolation
-- Spec: 017_PostgreSQLRLSImplementation_pending.md
-- Description: Implements PostgreSQL RLS for database-level tenant isolation
-- This provides defense-in-depth security at the database layer

-- ============================================================================
-- Phase 1: Create database users
-- ============================================================================

-- Create application user (used by the application at runtime)
-- This user does NOT have BYPASSRLS privilege
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        CREATE USER easi_app WITH PASSWORD 'change_me_in_production';
    END IF;
END
$$;

-- Create admin user (used for migrations and administrative tasks)
-- This user HAS BYPASSRLS privilege
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        CREATE USER easi_admin WITH PASSWORD 'change_me_in_production' BYPASSRLS;
    END IF;
END
$$;

-- Grant necessary permissions to application user
GRANT CONNECT ON DATABASE easi TO easi_app;
GRANT USAGE ON SCHEMA public TO easi_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO easi_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO easi_app;

-- Grant all permissions to admin user
GRANT ALL PRIVILEGES ON DATABASE easi TO easi_admin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO easi_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO easi_admin;

-- Ensure future tables also have proper permissions
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO easi_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT ON SEQUENCES TO easi_app;

-- ============================================================================
-- Phase 2: Enable RLS on all tenant-scoped tables
-- ============================================================================

ALTER TABLE events ENABLE ROW LEVEL SECURITY;
ALTER TABLE snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE application_components ENABLE ROW LEVEL SECURITY;
ALTER TABLE component_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE architecture_views ENABLE ROW LEVEL SECURITY;
ALTER TABLE view_component_positions ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- Phase 3: Create RLS policies for tenant isolation
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
-- Phase 4: Create helper functions for tenant context management
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

-- Grant execute permissions to application user
GRANT EXECUTE ON FUNCTION set_tenant_context(VARCHAR) TO easi_app;
GRANT EXECUTE ON FUNCTION get_current_tenant() TO easi_app;

-- ============================================================================
-- Migration complete
-- ============================================================================

-- Note: After this migration, the application should use the easi_app user
-- and set the tenant context via: SELECT set_tenant_context('tenant-id');
-- or directly via: SET app.current_tenant = 'tenant-id';
