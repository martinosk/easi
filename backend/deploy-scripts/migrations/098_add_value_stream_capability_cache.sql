CREATE TABLE IF NOT EXISTS value_stream_capability_cache (
    tenant_id VARCHAR(255) NOT NULL,
    id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX idx_vs_capability_cache_tenant ON value_stream_capability_cache(tenant_id);

ALTER TABLE value_stream_capability_cache ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON value_stream_capability_cache;
CREATE POLICY tenant_isolation_policy ON value_stream_capability_cache
    FOR ALL
    TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

INSERT INTO value_stream_capability_cache (tenant_id, id, name)
SELECT tenant_id, id, name
FROM capabilities
ON CONFLICT (tenant_id, id) DO NOTHING;
