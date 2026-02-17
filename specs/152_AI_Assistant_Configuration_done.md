# AI Assistant Configuration

**Status**: done

**Series**: Chat with your Architecture (1 of 3)
- **Spec 152**: AI Assistant Configuration (this spec)
- Spec 153: Architecture Assistant Backend
- Spec 154: Architecture Assistant Chat UI

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_done.md), [090_MetaModel_BoundedContext](090_MetaModel_BoundedContext_done.md)

## User Value

> "As a tenant admin, I want to configure which LLM provider and model my organization uses for the architecture assistant, so I can control costs, comply with data residency requirements, and choose the best model for our needs."

## Description

Foundation spec that establishes the `archassistant` bounded context, per-tenant LLM configuration, a new `assistant:use` permission, an agent token mechanism for secure API delegation, and a settings tab for managing the configuration.

## Design Decisions

- **Separate bounded context** (`archassistant`): The assistant has its own lifecycle, configuration, and conversation state that doesn't belong in any existing context.
- **CRUD persistence, not event sourcing**: Infrastructure configuration with no event consumers or temporal queries. Follows the View Layouts pattern.
- **Native provider choice: OpenAI or Anthropic**: Tenant explicitly chooses provider. Backend uses provider-specific API contract (`/v1/chat/completions` for OpenAI, `/v1/messages` for Anthropic) without requiring compatibility proxies.
- **Per-tenant configuration**: Each tenant stores its own endpoint, API key, model, and parameters. No global default — tenants must configure before use.
- **Encrypted API key storage**: AES-256-GCM with server-managed key. Tenant ID as AAD ensures cross-tenant isolation. Key versioned (`v1:` prefix) for future rotation. Never returned via API.
- **New permission `assistant:use`**: Granted to admin and architect roles. Configuration management reuses `metamodel:write`.
- **Agent token for API delegation**: The chat backend (spec 153) calls EASI APIs using a short-lived internal token — not the user's session. Carries only `userId`, `tenantId`, `source: "agent"`. Permissions resolved at request time via normal RBAC, never embedded in the token.

---

## Domain Model

### Aggregate: AIConfiguration

Singleton per tenant in `archassistant` bounded context.

```
id: AIConfigurationId (UUID)
provider: LLMProvider (openai | anthropic)
endpoint: LLMEndpoint (URL, max 500 chars, optional override; defaults to provider endpoint)
apiKey: EncryptedAPIKey (AES-256-GCM, stored as base64)
model: ModelName (string, max 100 chars)
maxTokens: MaxTokens (integer, 256..32768, default 4096)
temperature: Temperature (float, 0.0..2.0, default 0.3)
systemPromptOverride: string | null (max 2000 chars)
status: ConfigurationStatus (not_configured | configured | error)
```

**Invariants:** Provider required. Endpoint, if provided, must be valid URL (https or localhost). Model required. MaxTokens and temperature within bounds. API key required for `configured` status.

Repository uses CRUD with upsert semantics (`INSERT ... ON CONFLICT DO UPDATE`).

---

## Database

### Migration 104

```sql
CREATE SCHEMA IF NOT EXISTS archassistant;

CREATE TABLE archassistant.ai_configurations (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    endpoint VARCHAR(500),
    api_key_encrypted TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    max_tokens INTEGER NOT NULL DEFAULT 4096,
    temperature NUMERIC(3,1) NOT NULL DEFAULT 0.3,
    system_prompt_override TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'not_configured',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ai_configurations_tenant
    ON archassistant.ai_configurations (tenant_id);

ALTER TABLE archassistant.ai_configurations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_policy ON archassistant.ai_configurations
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

---

## Permission

New permission: `assistant:use` — admin and architect roles. Stakeholder: no.

Configuration management reuses `metamodel:write` (same as other settings).

---

## API Endpoints

All under `/assistant-config`. Requires authentication + `metamodel:write`.

| Method | Path | Description |
|---|---|---|
| `GET` | `/assistant-config` | Returns tenant config. `apiKeyStatus` is `configured` or `not_configured` — actual key never returned. Returns 200 with `status: not_configured` until credentials saved. |
| `PUT` | `/assistant-config` | Create/update config. `apiKey` optional on update (preserves existing). Validates provider, endpoint (if set), model, maxTokens, temperature. |
| `POST` | `/assistant-config/test` | Tests provider-specific LLM connectivity. Returns `{ success, provider, model, latencyMs }` or `{ success: false, error }`. |

### HATEOAS Links

| Link | Condition |
|---|---|
| `self` | Always |
| `update` | `metamodel:write` |
| `test` | `metamodel:write` AND status `configured` |

On `GET /auth/sessions/current`: add `x-assistant` link when actor has `assistant:use` AND config status is `configured`. This controls chat button visibility in the frontend.

---

## Cross-Context Integration

On `TenantCreated` event, create a default `AIConfiguration` in `not_configured` status. Same pattern as `MetaModelConfiguration`.

Published language exposes `AIConfigProvider` for spec 153 to get decrypted config (provider, endpoint override, key, model, parameters).

---

## Agent Token

Internal-only credential for the chat backend to make EASI API calls on behalf of the authenticated user without exposing the user's session.

**Payload:** `userId`, `tenantId`, `source: "agent"`, `exp` (expiry). Format: `base64(payload).base64(HMAC-SHA256)`.

**Signing:** HMAC-SHA256 with `AGENT_TOKEN_SECRET` env var (separate from `ENCRYPTION_KEY`).

**Lifecycle:** Minted by SSE handler after session auth. TTL 5 minutes (refreshed during long streams). Never leaves the server process. Never sent to LLM provider.

**Middleware extension:** Auth middleware recognizes `Authorization: AgentToken <token>`. Verifies signature + expiry, extracts userId/tenantId, resolves permissions via normal RBAC, sets `viaAgent = true` on actor context. Only accepted from loopback (`127.0.0.1` / `::1`).

**Audit:** When `viaAgent` is true, audit entries append `(via AI assistant)` to the log message.

---

## Frontend

### Settings Page: AI Configuration Tab

New tab on Settings page at `/settings/ai-configuration`. Gated on `metamodel:write`.

**Form fields:** LLM Endpoint, API Key (masked when configured, "Change" to reveal), Model, Advanced section (Max Tokens, Temperature, System Prompt Override).

**Provider selection:** Required dropdown with `OpenAI` and `Anthropic`. Endpoint field label changes to "Base URL override (optional)" and is pre-filled with provider default when empty at runtime.

**Actions:** "Test Connection" (calls test endpoint, shows result inline), "Save" (calls PUT).

**Data residency notice:** Info banner: *"Architecture data will be sent to the configured LLM endpoint. Ensure compliance with your organization's data handling requirements."*

**System prompt help text:** *"Provide additional organizational context. This is appended to the built-in system prompt as informational context."*

New dependency: `react-markdown` and `remark-gfm` (no `rehype-raw` — prevent XSS).

---

## Checklist

- [x] Specification approved
- [x] AES-256-GCM encryption utility + tests
- [x] Agent token mint/verify + tests
- [x] Auth middleware extended for `AgentToken` scheme
- [x] `viaAgent` flag on actor context + audit formatting
- [x] `assistant:use` permission (admin + architect)
- [x] `x-assistant` HATEOAS link on current session
- [x] AIConfiguration aggregate + value objects + tests
- [x] Migration 104
- [x] API handlers (GET, PUT, POST test) + HATEOAS links
- [x] Provider-specific connectivity tests (OpenAI + Anthropic)
- [x] TenantCreated handler (default config)
- [x] Published language: AIConfigProvider
- [x] Frontend: Settings tab with form, test connection, save
- [x] Tests and build passing
