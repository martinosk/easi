CREATE TABLE IF NOT EXISTS import_sessions (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    source_format VARCHAR(100) NOT NULL,
    business_domain_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    preview JSONB NOT NULL DEFAULT '{}',
    progress JSONB,
    parsed_data JSONB NOT NULL DEFAULT '{}',
    result JSONB,
    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_import_sessions_tenant_id ON import_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_import_sessions_status ON import_sessions(status);
CREATE INDEX IF NOT EXISTS idx_import_sessions_created_at ON import_sessions(created_at);

ALTER TABLE import_sessions ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON import_sessions;
CREATE POLICY tenant_isolation_policy ON import_sessions
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
