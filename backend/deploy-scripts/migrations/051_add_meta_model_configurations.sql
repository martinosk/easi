-- Migration: MetaModel Configurations Table
-- Spec: 090_MetaModel_BoundedContext
-- Description: Creates the read model table for the MetaModel bounded context

CREATE TABLE IF NOT EXISTS meta_model_configurations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    sections JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    is_default BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_by VARCHAR(255) NOT NULL,

    CONSTRAINT uq_metamodel_tenant UNIQUE (tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_metamodel_configurations_tenant ON meta_model_configurations(tenant_id);

-- Row-Level Security for tenant isolation
ALTER TABLE meta_model_configurations ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON meta_model_configurations;
CREATE POLICY tenant_isolation_policy ON meta_model_configurations
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
