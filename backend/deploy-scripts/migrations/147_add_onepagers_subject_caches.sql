-- Migration: OnePagers subject caches (spec 209)
-- Description: OnePagers renders fact sheets over subjects owned by Architecture Modeling,
--              Capability Mapping and Enterprise Architecture, and needs a maturity scale owned
--              by MetaModel. These caches let OnePagers implement SubjectExistenceChecker,
--              BuiltInFieldSource and MaturityScaleSource over its own tables, fed by the
--              suppliers' published events instead of query-time reads of their read models.
--
--   one_pager_subject_index.built_in_fields  built-in attribute values per subject (name comes
--                                            from the existing name column)
--   subject_relation_cache                   one row per relation entry reference; related_name
--                                            is only authoritative for related things that are
--                                            not one-pager subjects (business domains) — for the
--                                            six subject types the label is joined live from
--                                            one_pager_subject_index so renames are reflected
--   business_domain_name_cache               business-domain names, the only relation target that
--                                            is not itself a one-pager subject
--   maturity_scale_cache                     MetaModel's maturity scale sections per tenant

ALTER TABLE onepagers.one_pager_subject_index
    ADD COLUMN IF NOT EXISTS built_in_fields JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE TABLE IF NOT EXISTS onepagers.subject_relation_cache (
    tenant_id VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    entry_id VARCHAR(50) NOT NULL,
    related_type VARCHAR(50) NOT NULL DEFAULT '',
    related_id VARCHAR(255) NOT NULL,
    related_name TEXT NOT NULL DEFAULT '',
    edge_id VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, subject_type, subject_id, entry_id, related_id)
);

CREATE INDEX IF NOT EXISTS idx_onepagers_subject_relation_cache_edge
    ON onepagers.subject_relation_cache (tenant_id, edge_id);

CREATE INDEX IF NOT EXISTS idx_onepagers_subject_relation_cache_entry_related
    ON onepagers.subject_relation_cache (tenant_id, entry_id, related_id);

CREATE INDEX IF NOT EXISTS idx_onepagers_subject_relation_cache_related
    ON onepagers.subject_relation_cache (tenant_id, related_type, related_id);

ALTER TABLE onepagers.subject_relation_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.subject_relation_cache;
CREATE POLICY tenant_isolation_policy ON onepagers.subject_relation_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS onepagers.business_domain_name_cache (
    tenant_id VARCHAR(50) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, business_domain_id)
);

ALTER TABLE onepagers.business_domain_name_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.business_domain_name_cache;
CREATE POLICY tenant_isolation_policy ON onepagers.business_domain_name_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS onepagers.maturity_scale_cache (
    tenant_id VARCHAR(50) NOT NULL,
    sections JSONB NOT NULL DEFAULT '[]'::jsonb,
    PRIMARY KEY (tenant_id)
);

ALTER TABLE onepagers.maturity_scale_cache ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_policy ON onepagers.maturity_scale_cache;
CREATE POLICY tenant_isolation_policy ON onepagers.maturity_scale_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
DECLARE
    target_table TEXT;
BEGIN
    FOREACH target_table IN ARRAY ARRAY[
        'onepagers.subject_relation_cache',
        'onepagers.business_domain_name_cache',
        'onepagers.maturity_scale_cache'
    ] LOOP
        IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
            EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON %s TO easi_app', target_table);
        END IF;
        IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
            EXECUTE format('GRANT ALL PRIVILEGES ON %s TO easi_admin', target_table);
        END IF;
    END LOOP;
END $$;
