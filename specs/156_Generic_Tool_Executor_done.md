# Generic Tool Executor

**Status**: done

**Series**: Architecture Assistant Evolution (2 of 6)
- Spec 155: Agent Permission Ceiling
- **Spec 156**: Generic Tool Executor (this spec)
- Spec 157: Tool Catalog per Bounded Context
- Spec 158: Domain Knowledge Injection
- Spec 159: Expand Tool Coverage
- Spec 160: Agent Audit Events

**Depends on:** Spec 155

## User Value

> "As a developer, I want to add a new agent tool by declaring a spec — not by writing a new Go struct, handler function, and registration code — so that expanding the assistant's capabilities is fast and safe."

## Problem

Each of the 24 tools today is a separate Go struct with a dedicated `Execute` method, registration code, and parameter handling. Adding a tool requires ~30 lines of boilerplate. The `query` and `mutation` helpers in `toolimpls/` already demonstrate that execution is generic HTTP — the per-tool code is just glue.

## Solution

### AgentToolSpec

A declarative structure describing a tool without any execution code:

```go
type AgentToolSpec struct {
    Name        string
    Description string
    Access      AccessClass
    Permission  string
    Method      string        // GET, POST, PUT, DELETE
    Path        string        // "/capabilities/{capabilityID}/realizations"
    PathParams  []ParamSpec
    QueryParams []ParamSpec
    BodyParams  []ParamSpec
}

type ParamSpec struct {
    Name        string
    Type        string // "string", "integer", "boolean"
    Description string
    Required    bool
}
```

### GenericAPIToolExecutor

A single executor that handles any `AgentToolSpec`:

1. Validates and extracts parameters from LLM arguments
2. Substitutes path parameters into the URL
3. Builds query string or JSON body
4. Calls `agenthttp.Client` with the appropriate method
5. Returns the response body as-is (JSON) for LLM consumption

Lives in `archassistant/infrastructure/toolimpls/generic_executor.go`.

### Composite Tools

3-5 tools that call multiple endpoints remain as explicit implementations: `search_architecture`, `get_portfolio_summary`, and any future cross-endpoint tools. These implement a `CompositeToolExecutor` interface alongside the generic executor.

### Migration Path

1. Implement `GenericAPIToolExecutor` and `AgentToolSpec`
2. Re-express all 24 existing tools as `AgentToolSpec` declarations
3. Verify behavior is identical (same tool names, same parameters, same responses)
4. Delete the 24 per-tool struct files
5. Keep composite tools as explicit implementations

## Design Decisions

- **Raw JSON responses, not formatted text.** Current tools use helpers like `formatApplicationDetails` to produce human-readable output. Modern LLMs handle JSON well. Drop custom formatting; return API responses directly. This eliminates the largest source of per-tool code.
- **Parameter validation in the executor.** UUID format, string length, integer range — all handled generically based on `ParamSpec.Type`.
- **No code generation.** The specs are Go data literals, not generated artifacts. They are readable, reviewable, and testable.

## Checklist

- [x] Specification approved
- [x] `AgentToolSpec` and `ParamSpec` types defined
- [x] `GenericAPIToolExecutor` implemented with method dispatch (GET/POST/PUT/DELETE)
- [x] Path parameter substitution
- [x] Query parameter building
- [x] JSON body building
- [x] Generic parameter validation (UUID, string length, integer range)
- [x] All 22 single-endpoint tools re-expressed as `AgentToolSpec` declarations
- [x] Composite tools (`search_architecture`, `get_portfolio_summary`, `list_application_relations`) kept as explicit implementations
- [x] Old per-tool structs deleted
- [x] Behavior-identical: same tool names, same parameters
- [x] Unit tests for generic executor
- [x] Integration test: round-trip through registry → executor → mock HTTP
- [x] Build passing
