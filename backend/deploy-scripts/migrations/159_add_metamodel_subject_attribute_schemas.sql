-- Migration: MetaModel subject attribute schemas (spec 217)
-- Description: Read model for the per-subject-type custom attribute schema owned by MetaModel.
--              One row per tenant and subject type; attributes is the projected attribute set
--              (id, name, type, helpText, active, options, min, max).

CREATE TABLE IF NOT EXISTS metamodel.subject_attribute_schemas (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    attributes JSONB NOT NULL DEFAULT '[]'::jsonb,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    modified_by VARCHAR(500) NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_metamodel_subject_attribute_schemas_subject_type
    ON metamodel.subject_attribute_schemas (tenant_id, subject_type);

ALTER TABLE metamodel.subject_attribute_schemas ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON metamodel.subject_attribute_schemas;
CREATE POLICY tenant_isolation_policy ON metamodel.subject_attribute_schemas
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON metamodel.subject_attribute_schemas TO easi_app;
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        GRANT ALL PRIVILEGES ON metamodel.subject_attribute_schemas TO easi_admin;
    END IF;
END $$;
