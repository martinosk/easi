CREATE TABLE IF NOT EXISTS edit_grants (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    grantor_id VARCHAR(255) NOT NULL,
    grantor_email VARCHAR(255) NOT NULL,
    grantee_email VARCHAR(255) NOT NULL,
    artifact_type VARCHAR(20) NOT NULL,
    artifact_id UUID NOT NULL,
    scope VARCHAR(20) NOT NULL DEFAULT 'write',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

CREATE INDEX idx_edit_grants_tenant ON edit_grants(tenant_id);
CREATE INDEX idx_edit_grants_grantee_active ON edit_grants(tenant_id, grantee_email, artifact_type, artifact_id)
    WHERE status = 'active';
CREATE INDEX idx_edit_grants_artifact ON edit_grants(tenant_id, artifact_type, artifact_id, status);
CREATE INDEX idx_edit_grants_grantor ON edit_grants(tenant_id, grantor_id);

ALTER TABLE edit_grants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON edit_grants
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
