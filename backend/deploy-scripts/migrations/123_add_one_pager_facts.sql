-- Migration: Add One-Pager Facts
-- Spec: 176_OnePagerFacts
-- Description: Add one_pager_facts read model table for recorded field values per subject

CREATE TABLE IF NOT EXISTS onepagers.one_pager_facts (
    tenant_id VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    field_id VARCHAR(255) NOT NULL,
    facts_id VARCHAR(255) NOT NULL,
    value JSONB,
    value_type VARCHAR(50),
    display_text TEXT,
    modified_at TIMESTAMP NOT NULL,
    modified_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, subject_type, subject_id, field_id)
);

CREATE INDEX IF NOT EXISTS idx_one_pager_facts_facts_id
    ON onepagers.one_pager_facts(tenant_id, facts_id);

ALTER TABLE onepagers.one_pager_facts ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.one_pager_facts;
CREATE POLICY tenant_isolation_policy ON onepagers.one_pager_facts
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON onepagers.one_pager_facts TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON onepagers.one_pager_facts TO easi_admin';
    END IF;
END $$;
