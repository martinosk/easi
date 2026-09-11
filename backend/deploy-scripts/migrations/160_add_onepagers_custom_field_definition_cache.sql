-- Migration: OnePagers custom-field definition cache (spec 217)
-- Description: OnePagers no longer owns custom-field definitions; MetaModel publishes them as
--              SubjectAttribute* events and OnePagers keeps this local cache, one row per field.
--              definition holds the projected definition (id, name, type, helpText, active,
--              options, min, max). pending_transfer marks rows seeded by the backfill from the
--              legacy OnePagers configuration that still have to be handed to MetaModel through
--              its ImportSubjectAttribute command; the cache projector clears the flag once
--              MetaModel has published the attribute.

CREATE TABLE IF NOT EXISTS onepagers.custom_field_definition_cache (
    tenant_id VARCHAR(50) NOT NULL,
    field_id VARCHAR(255) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    definition JSONB NOT NULL,
    pending_transfer BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (tenant_id, field_id)
);

CREATE INDEX IF NOT EXISTS idx_onepagers_custom_field_definition_cache_subject_type
    ON onepagers.custom_field_definition_cache (tenant_id, subject_type);

ALTER TABLE onepagers.custom_field_definition_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.custom_field_definition_cache;
CREATE POLICY tenant_isolation_policy ON onepagers.custom_field_definition_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON onepagers.custom_field_definition_cache TO easi_app;
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        GRANT ALL PRIVILEGES ON onepagers.custom_field_definition_cache TO easi_admin;
    END IF;
END $$;
