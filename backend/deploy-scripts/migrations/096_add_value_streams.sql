CREATE TABLE IF NOT EXISTS value_streams (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    stage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id),
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_value_streams_tenant ON value_streams(tenant_id);
CREATE INDEX IF NOT EXISTS idx_value_streams_name ON value_streams(tenant_id, name);

ALTER TABLE value_streams ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON value_streams;
CREATE POLICY tenant_isolation_policy ON value_streams
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
