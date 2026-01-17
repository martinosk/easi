-- Migration: Add Origin Relationships Tables
-- Spec: 117_Portfolio_Metadata_Foundation
-- Description: Add tables for origin relationships (acquired_via, purchased_from, built_by)

-- Acquired Via Relationships table
CREATE TABLE IF NOT EXISTS acquired_via_relationships (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    acquired_entity_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    notes VARCHAR(500),
    created_at TIMESTAMP NOT NULL,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, acquired_entity_id, component_id)
);

CREATE INDEX IF NOT EXISTS idx_acquired_via_relationships_entity ON acquired_via_relationships(tenant_id, acquired_entity_id);
CREATE INDEX IF NOT EXISTS idx_acquired_via_relationships_component ON acquired_via_relationships(tenant_id, component_id);

ALTER TABLE acquired_via_relationships ENABLE ROW LEVEL SECURITY;

CREATE POLICY acquired_via_relationships_tenant_isolation ON acquired_via_relationships
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Purchased From Relationships table
CREATE TABLE IF NOT EXISTS purchased_from_relationships (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    vendor_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    notes VARCHAR(500),
    created_at TIMESTAMP NOT NULL,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, vendor_id, component_id)
);

CREATE INDEX IF NOT EXISTS idx_purchased_from_relationships_vendor ON purchased_from_relationships(tenant_id, vendor_id);
CREATE INDEX IF NOT EXISTS idx_purchased_from_relationships_component ON purchased_from_relationships(tenant_id, component_id);

ALTER TABLE purchased_from_relationships ENABLE ROW LEVEL SECURITY;

CREATE POLICY purchased_from_relationships_tenant_isolation ON purchased_from_relationships
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Built By Relationships table
CREATE TABLE IF NOT EXISTS built_by_relationships (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    internal_team_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    notes VARCHAR(500),
    created_at TIMESTAMP NOT NULL,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, internal_team_id, component_id)
);

CREATE INDEX IF NOT EXISTS idx_built_by_relationships_team ON built_by_relationships(tenant_id, internal_team_id);
CREATE INDEX IF NOT EXISTS idx_built_by_relationships_component ON built_by_relationships(tenant_id, component_id);

ALTER TABLE built_by_relationships ENABLE ROW LEVEL SECURITY;

CREATE POLICY built_by_relationships_tenant_isolation ON built_by_relationships
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
