CREATE TABLE IF NOT EXISTS value_stream_stages (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    value_stream_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    position INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS idx_value_stream_stages_tenant ON value_stream_stages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_value_stream_stages_vs ON value_stream_stages(tenant_id, value_stream_id);
CREATE INDEX IF NOT EXISTS idx_value_stream_stages_position ON value_stream_stages(tenant_id, value_stream_id, position);

ALTER TABLE value_stream_stages ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON value_stream_stages;
CREATE POLICY tenant_isolation_policy ON value_stream_stages
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE IF NOT EXISTS value_stream_stage_capabilities (
    tenant_id VARCHAR(50) NOT NULL,
    stage_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, stage_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_vs_stage_caps_tenant ON value_stream_stage_capabilities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vs_stage_caps_stage ON value_stream_stage_capabilities(tenant_id, stage_id);
CREATE INDEX IF NOT EXISTS idx_vs_stage_caps_capability ON value_stream_stage_capabilities(tenant_id, capability_id);

ALTER TABLE value_stream_stage_capabilities ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON value_stream_stage_capabilities;
CREATE POLICY tenant_isolation_policy ON value_stream_stage_capabilities
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
