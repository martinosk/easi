# Bounded Context Canvas: Arch Assistant

## Name
**Arch Assistant**

## Purpose
Provide an AI-powered conversational assistant that helps users explore and modify their enterprise architecture through natural language. The assistant acts as an intelligent agent that can read and write architecture data across all bounded contexts via a loopback API, governed by user permissions and a hard permission ceiling.

**Key Stakeholders:**
- Enterprise Architects (primary users of the conversational interface)
- Solution Architects (querying and exploring architecture data)
- Tenant Administrators (configuring LLM provider settings)

**Value Proposition:**
- Natural language interface for architecture exploration and modification
- Reduces barrier to entry for non-expert users navigating complex architecture models
- Cross-context querying and modification through a single conversational interface
- Permission-controlled write operations with safety budgets per access class
- Multi-provider LLM support (OpenAI, Anthropic) configurable per tenant

## Strategic Classification

### Domain Importance
**Supporting Domain** - Enhances the usability of all core domains but does not define the architecture modeling methodology itself. The competitive advantage lies in the core domains (Capability Mapping, Enterprise Architecture, Value Streams); this context makes them more accessible.

### Business Model
**Engagement Creator** - Increases user engagement by lowering the barrier to exploring and modifying architecture data through conversational AI.

### Evolution Stage
**Custom-Built** - Tailored agent loop with domain-specific tool registry, permission ceiling, and rate limiting. Not a generic chatbot wrapper.

## Domain Roles
- **Gateway**: Single conversational entry point that fans out to all other bounded contexts via tool calls
- **Orchestrator**: Manages the agent loop (LLM → tool call → result → LLM) with iteration limits and budget enforcement
- **Configuration Holder**: Stores per-tenant LLM provider configuration (provider, endpoint, API key, model, temperature)

## Inbound Communication

### Messages Received

**Commands** (from Frontend/API):
- `UpdateAIConfiguration` - Tenant admin configures LLM provider settings
- `TestConnection` - Tenant admin tests LLM connectivity
- `CreateConversation` - User starts a new conversation
- `DeleteConversation` - User deletes a conversation
- `SendMessage` - User sends a message (triggers agent loop, returns SSE stream)

**Events** (from other contexts):
- From **Platform (Auth)**:
  - `TenantCreated` - Provisions a blank `AIConfiguration` for the new tenant

### Collaborators
- **Frontend UI**: Primary source of commands and SSE stream consumer
- **Auth Context**: Source of `TenantCreated` events for auto-provisioning
- **All other Bounded Contexts**: Provide tool specifications via `publishedlanguage.AgentToolSpec`; tools execute via loopback HTTP to their respective APIs

### Relationship Types
- **Customer-Supplier** with Auth (downstream of `TenantCreated`)
- **Open Host Service** to all other contexts (consumes their APIs via loopback HTTP, not direct domain coupling)

## Outbound Communication

### Messages Sent

**SSE Events** (streamed to frontend):
- `token` - Incremental text token from LLM
- `done` - Stream complete (includes messageId, tokensUsed)
- `error` - Error occurred (codes: `iteration_limit`, `timeout`, `validation_error`, `llm_error`)
- `thinking` - Agent processing status
- `tool_call_start` - Tool execution beginning (includes tool name and arguments)
- `tool_call_result` - Tool execution complete (includes result preview)
- `ping` - Keep-alive

**Tool Calls** (to other contexts via loopback HTTP):
- To **Architecture Modeling**: Component CRUD, vendor/acquired entity/internal team management, origin links
- To **Capability Mapping**: Capability CRUD, metadata, business domains, dependencies, strategy
- To **Enterprise Architecture**: Enterprise capability management
- To **Value Streams**: Value stream and stage management
- To **MetaModel**: Configuration queries

### Collaborators
- **Frontend UI**: Receives SSE stream
- **All Bounded Contexts**: Receive loopback HTTP requests from agent tool execution

### Integration Pattern
- **Loopback HTTP** for tool execution: Agent mints a short-lived `AgentToken`, calls back into the EASI API as if it were a user, routed through the normal request pipeline of each target context
- **SSE streaming** for real-time response delivery to the frontend
- **Event subscription** for `TenantCreated` provisioning

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **AI Configuration** | Per-tenant settings for the LLM provider (provider, endpoint, API key, model, temperature, max tokens) |
| **LLM Provider** | The AI model vendor: `openai` or `anthropic` |
| **Configuration Status** | Whether the tenant's AI is ready: `not_configured`, `configured`, or `error` |
| **Conversation** | A named thread of messages between a user and the assistant |
| **Message** | A single exchange unit with a role (`user`, `assistant`, `tool`), content, and optional token count |
| **Agent Loop** | The iterative cycle where the LLM generates a response, optionally requests tool calls, receives results, and continues until it produces a final text response |
| **Tool Registry** | The collection of all available tools the agent can call, assembled from tool specs contributed by each bounded context |
| **Agent Tool Spec** | The contract (`AgentToolSpec`) each bounded context implements to register its tools with the assistant |
| **Access Class** | Budget category for tool calls: `read` (500/message), `create` (50), `update` (100), `delete` (5) |
| **Permission Ceiling** | The hard upper bound on what the agent can do, regardless of the user's actual permissions: a fixed set of scoped permissions |
| **Loopback API** | The pattern where the agent calls back into the EASI API itself using an `AgentToken`, executing tools through the normal API pipeline |
| **Agent Token** | A short-lived authentication token minted at message send time, used by the agent for loopback HTTP calls |
| **System Prompt** | The assembled instructions given to the LLM, including tenant context, role, write mode rules, domain model summary, and optional tenant override |
| **Composite Tool** | A tool defined within this context that aggregates data from multiple other contexts (e.g., `search_architecture`, `get_portfolio_summary`) |
| **Context Window** | The LLM's token budget (default 128K); conversation history is truncated to fit |

## Business Decisions

### Core Business Rules
1. **Permission ceiling**: The agent can never exceed a fixed set of permissions (`components:read/write`, `capabilities:read/write`, `domains:read/write`, `valuestreams:read/write`, `enterprise-arch:read/write`, `views:read`, `metamodel:read`, `assistant:use`), regardless of the user's actual role
2. **Tool call budgets**: Each access class has a maximum number of calls per message (read: 500, create: 50, update: 100, delete: 5) to prevent runaway agent behavior
3. **Iteration limit**: The agent loop terminates after 50 iterations maximum
4. **Conversation ownership**: Users can only access their own conversations
5. **Conversation limit**: Maximum 100 conversations per user
6. **API key encryption**: LLM API keys are encrypted at rest using the tenant ID as encryption context
7. **System prompt injection sanitization**: Tenant-provided system prompt overrides are sanitized for prompt injection patterns (e.g., "ignore previous", "system prompt")
8. **Write mode gating**: Write operations require explicit opt-in; read-only mode strips all write tools from the registry

### Policy Decisions
- Conversations use CRUD persistence (not event sourcing) for simplicity
- No shared conversation state between users
- Rate limiting is in-memory (resets on server restart)
- Tool results are truncated to 32K characters to fit context window
- Usage tracking is append-only for future analytics (prompt tokens, completion tokens, tool calls count, model, latency)

## Assumptions

1. **Concurrent users**: Few concurrent agent sessions per tenant (rate limit: 50 messages/minute per tenant)
2. **Message frequency**: Individual users send at most 10 messages/minute
3. **Conversation length**: Conversations are short to medium (context window manages overflow via truncation)
4. **LLM availability**: External LLM providers (OpenAI, Anthropic) are generally available; connection test verifies before use
5. **Loopback latency**: Loopback HTTP calls to the EASI API complete within 5 seconds
6. **Tool count**: Total tool catalogue is manageable (currently ~50+ tools across all contexts)
7. **Single stream**: Only one concurrent streaming session per user

## Verification Metrics

### Boundary Health Indicators
- **Zero direct domain coupling**: All cross-context interaction via loopback HTTP, never direct imports
- **Published language stability**: `AgentToolSpec` contract remains backward-compatible
- **Permission ceiling enforcement**: Agent never executes a tool outside the ceiling

### Context Effectiveness Metrics
- **Tool execution success rate**: Percentage of tool calls that return non-error results
- **Agent loop completion rate**: Percentage of messages that reach a final text response (vs. iteration limit or timeout)
- **Average iterations per message**: Lower is better (indicates efficient tool use)

### Business Value Metrics
- **User engagement**: Number of conversations and messages per user
- **Write operation adoption**: Ratio of write vs. read tool calls (indicates trust in agent modifications)
- **Token efficiency**: Average tokens used per message (cost optimization)

## Open Questions

1. **Should domain events be published to the shared bus?** Currently `ConversationStarted`, `UserMessageSent`, and `AssistantMessageReceived` are defined but not published externally. Should they be for audit or analytics?

2. **Persistent rate limiting?** Current in-memory rate limiter resets on restart. Should rate limits be persisted (Redis, database)?

3. **Multi-turn tool confirmation?** Should the agent ask the user for confirmation before executing write operations, rather than relying solely on the write mode flag?

4. **Conversation sharing?** Should users be able to share conversations with team members or export them?

5. **Model selection per conversation?** Should users be able to override the tenant-wide model configuration for specific conversations?

6. **Usage quotas?** Should there be token or cost budgets per user/tenant beyond the current rate limits?

7. **Tool result caching?** Should identical read tool calls within the same agent loop return cached results to save loopback calls?

## Architecture Notes

### Implementation Location
`/backend/internal/archassistant/`

### Key Packages
- `domain/aggregates/` - AIConfiguration, Conversation (with Message entity)
- `domain/valueobjects/` - LLMProvider, LLMEndpoint, EncryptedAPIKey, ModelName, MaxTokens, Temperature, ConfigurationStatus, ConversationTitle, MessageContent, MessageRole, TokenCount
- `domain/` - Repository interfaces (AIConfigurationRepository, ConversationRepository)
- `application/orchestrator/` - Agent loop orchestrator with streaming support
- `application/systemprompt/` - System prompt builder with injection sanitization
- `application/tools/` - Tool registry, permission checker, LLM tool format conversion
- `application/handlers/` - TenantCreatedHandler (event handler for auto-provisioning)
- `infrastructure/api/` - REST routes and handlers
- `infrastructure/adapters/` - AIConfigProviderAdapter, AIConfigStatusAdapter, LLMClientFactory
- `infrastructure/llm/` - OpenAI and Anthropic streaming client implementations
- `infrastructure/agenthttp/` - Loopback HTTP client with AgentToken authentication
- `infrastructure/ratelimit/` - In-memory sliding window rate limiter
- `infrastructure/sse/` - Server-Sent Events writer
- `infrastructure/repositories/` - PostgreSQL repositories (ai_configurations, conversations, usage_tracking)
- `infrastructure/toolimpls/` - GenericAPIToolExecutor, tool catalogue assembly, composite tools
- `publishedlanguage/` - AgentToolSpec, AccessClass, AIConfigProvider, AIConfigInfo

### Technical Patterns
- **No CQRS/Event Sourcing**: Uses CRUD repositories (not event-sourced aggregates) for both AIConfiguration and Conversation
- **Agent Loop**: Iterative LLM → tool call → result cycle with budget enforcement and iteration limits
- **Loopback HTTP**: Tool execution via HTTP calls back to the EASI API, authenticated with short-lived AgentToken
- **SSE Streaming**: Real-time response delivery to frontend
- **Multi-provider LLM**: Abstracted client interface supporting OpenAI and Anthropic APIs
- **Tool Catalogue**: Tools assembled from `AgentToolSpec` implementations in other contexts' published languages, plus 4 composite tools defined locally

### API Style
- REST Level 3 with HATEOAS
- SSE streaming for message responses
- Paginated conversation listing

### Cross-Context Integration
- **Downstream of Auth**: Listens to `TenantCreated` to auto-provision AI configuration
- **Loopback consumer of all contexts**: Executes tools via HTTP against Architecture Modeling, Capability Mapping, Enterprise Architecture, Value Streams, and MetaModel APIs
- **Published Language provider**: Defines `AgentToolSpec` contract that other contexts implement to contribute tools
