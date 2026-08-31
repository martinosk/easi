-- Migration: 141_add_accessdelegation_artifact_name_cache
-- Purpose: Access-Delegation-owned cache of grant artifact display names (spec 209) so edit
-- grants can be rendered without reading Capability Mapping, Architecture Modeling or
-- Architecture Views read models at request time. Maintained by a projector on those contexts'
-- published Created/Updated/Renamed/Deleted events; backfilled by migration 142.
-- artifact_type holds the same values as accessdelegation.edit_grants.artifact_type:
-- capability, component, view, domain, vendor, acquired_entity, internal_team.

CREATE TABLE IF NOT EXISTS accessdelegation.artifact_name_cache (
    tenant_id VARCHAR(50) NOT NULL,
    artifact_type VARCHAR(50) NOT NULL,
    artifact_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (tenant_id, artifact_type, artifact_id)
);

ALTER TABLE accessdelegation.artifact_name_cache ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON accessdelegation.artifact_name_cache;
CREATE POLICY tenant_isolation_policy ON accessdelegation.artifact_name_cache
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_app') THEN
        EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON accessdelegation.artifact_name_cache TO easi_app';
    END IF;
    IF EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'easi_admin') THEN
        EXECUTE 'GRANT ALL PRIVILEGES ON accessdelegation.artifact_name_cache TO easi_admin';
    END IF;
END $$;
