CREATE SCHEMA IF NOT EXISTS archassistant;

CREATE TABLE archassistant.ai_configurations (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    provider VARCHAR(20) NOT NULL DEFAULT '',
    endpoint VARCHAR(500) NOT NULL DEFAULT '',
    api_key_encrypted TEXT NOT NULL DEFAULT '',
    model VARCHAR(100) NOT NULL DEFAULT '',
    max_tokens INTEGER NOT NULL DEFAULT 4096,
    temperature NUMERIC(3,1) NOT NULL DEFAULT 0.3,
    system_prompt_override TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'not_configured',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ai_configurations_tenant
    ON archassistant.ai_configurations (tenant_id);

ALTER TABLE archassistant.ai_configurations ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON archassistant.ai_configurations;
CREATE POLICY tenant_isolation_policy ON archassistant.ai_configurations
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
