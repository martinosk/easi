-- Migration: Add Origin Entities Tables
-- Spec: 117_Portfolio_Metadata_Foundation
-- Description: Add tables for acquired entities, vendors, and internal teams

-- Acquired Entities table
CREATE TABLE IF NOT EXISTS acquired_entities (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    acquisition_date DATE,
    integration_status VARCHAR(50),
    notes VARCHAR(500),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS idx_acquired_entities_tenant ON acquired_entities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_acquired_entities_name ON acquired_entities(tenant_id, name);

ALTER TABLE acquired_entities ENABLE ROW LEVEL SECURITY;

CREATE POLICY acquired_entities_tenant_isolation ON acquired_entities
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Vendors table
CREATE TABLE IF NOT EXISTS vendors (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    implementation_partner VARCHAR(100),
    notes VARCHAR(500),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS idx_vendors_tenant ON vendors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vendors_name ON vendors(tenant_id, name);

ALTER TABLE vendors ENABLE ROW LEVEL SECURITY;

CREATE POLICY vendors_tenant_isolation ON vendors
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

-- Internal Teams table
CREATE TABLE IF NOT EXISTS internal_teams (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    department VARCHAR(100),
    contact_person VARCHAR(100),
    notes VARCHAR(500),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS idx_internal_teams_tenant ON internal_teams(tenant_id);
CREATE INDEX IF NOT EXISTS idx_internal_teams_name ON internal_teams(tenant_id, name);

ALTER TABLE internal_teams ENABLE ROW LEVEL SECURITY;

CREATE POLICY internal_teams_tenant_isolation ON internal_teams
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
