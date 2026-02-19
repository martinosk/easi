CREATE TABLE archassistant.conversations (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(100) NOT NULL DEFAULT 'New conversation',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_user
    ON archassistant.conversations (tenant_id, user_id, last_message_at DESC);

ALTER TABLE archassistant.conversations ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON archassistant.conversations;
CREATE POLICY tenant_isolation_policy ON archassistant.conversations
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE archassistant.messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    tool_calls JSONB,
    tool_call_id VARCHAR(100),
    tool_name VARCHAR(100),
    tokens_used INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation
    ON archassistant.messages (conversation_id, created_at);

ALTER TABLE archassistant.messages ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON archassistant.messages;
CREATE POLICY tenant_isolation_policy ON archassistant.messages
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));

CREATE TABLE archassistant.usage_tracking (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID NOT NULL,
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    tool_calls_count INTEGER NOT NULL DEFAULT 0,
    model_used VARCHAR(100) NOT NULL,
    latency_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_tracking_tenant
    ON archassistant.usage_tracking (tenant_id, created_at DESC);

ALTER TABLE archassistant.usage_tracking ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_policy ON archassistant.usage_tracking;
CREATE POLICY tenant_isolation_policy ON archassistant.usage_tracking
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
