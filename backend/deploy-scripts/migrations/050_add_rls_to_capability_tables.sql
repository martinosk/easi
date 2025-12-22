-- Migration: Add Row-Level Security to Capability Tables
-- Description: Adds RLS policies to tenant-scoped tables that were missing them
-- These tables have tenant_id columns but no database-level isolation enforcement

-- ============================================================================
-- Phase 1: Enable RLS on all missing tables
-- ============================================================================

-- Capability domain tables
ALTER TABLE capabilities ENABLE ROW LEVEL SECURITY;
ALTER TABLE capability_realizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE capability_dependencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE capability_experts ENABLE ROW LEVEL SECURITY;
ALTER TABLE capability_tags ENABLE ROW LEVEL SECURITY;

-- Business domain tables
ALTER TABLE business_domains ENABLE ROW LEVEL SECURITY;
ALTER TABLE domain_capability_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE domain_composition_view ENABLE ROW LEVEL SECURITY;

-- View preferences table
ALTER TABLE view_preferences ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- Phase 2: Create RLS policies for tenant isolation
-- ============================================================================

-- Capabilities table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON capabilities;
CREATE POLICY tenant_isolation_policy ON capabilities
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Capability realizations table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON capability_realizations;
CREATE POLICY tenant_isolation_policy ON capability_realizations
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Business domains table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON business_domains;
CREATE POLICY tenant_isolation_policy ON business_domains
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Domain capability assignments table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON domain_capability_assignments;
CREATE POLICY tenant_isolation_policy ON domain_capability_assignments
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Domain composition view table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON domain_composition_view;
CREATE POLICY tenant_isolation_policy ON domain_composition_view
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Capability dependencies table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON capability_dependencies;
CREATE POLICY tenant_isolation_policy ON capability_dependencies
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Capability experts table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON capability_experts;
CREATE POLICY tenant_isolation_policy ON capability_experts
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Capability tags table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON capability_tags;
CREATE POLICY tenant_isolation_policy ON capability_tags
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- View preferences table policies
DROP POLICY IF EXISTS tenant_isolation_policy ON view_preferences;
CREATE POLICY tenant_isolation_policy ON view_preferences
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- ============================================================================
-- Migration complete
-- ============================================================================
