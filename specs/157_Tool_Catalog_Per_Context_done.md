# Tool Catalog per Bounded Context

**Status**: done

**Series**: Architecture Assistant Evolution (3 of 6)
- Spec 155: Agent Permission Ceiling
- Spec 156: Generic Tool Executor
- **Spec 157**: Tool Catalog per Bounded Context (this spec)
- Spec 158: Domain Knowledge Injection
- Spec 159: Expand Tool Coverage
- Spec 160: Agent Audit Events

**Depends on:** Spec 156

## User Value

> "As an architect maintaining EASI, I want each bounded context to declare which operations the AI assistant can perform, so that tool coverage grows naturally as the platform evolves — and new contexts are blocked by default."

## Problem

Tool registration lives centrally in `archassistant/infrastructure/toolimpls/registration.go`. The archassistant context "knows" the internals of every other context's API. This creates tight coupling and means tool coverage only grows when someone edits the archassistant package.

## Solution

### Context-Owned Tool Specs

Each bounded context declares its agent tools in its `publishedlanguage` package:

```
capabilitymapping/publishedlanguage/agent_tools.go
architecturemodeling/publishedlanguage/agent_tools.go
enterprisearchitecture/publishedlanguage/agent_tools.go
valuestreams/publishedlanguage/agent_tools.go
metamodel/publishedlanguage/agent_tools.go
```

Each file exports a function returning `[]tools.AgentToolSpec`. The archassistant context collects and registers them.

### Context Allowlist

The archassistant context maintains a hard-coded allowlist of contexts it will load tools from:

```go
var allowedContexts = []func() []tools.AgentToolSpec{
    cmpublished.AgentTools,
    ampublished.AgentTools,
    eapublished.AgentTools,
    vspublished.AgentTools,
    mmpublished.AgentTools,
}
```

Contexts not in this list are excluded by default. Adding a new context requires a one-line change here — a conscious opt-in.

### CI Drift Detection

An architecture guard test verifies:
1. Every `AgentToolSpec` references an API path that exists in the route registrations
2. Every spec's permission string is a valid permission constant
3. No specs reference blocked contexts

This catches stale tools (endpoint removed but spec remains) and missing coverage (new endpoint in an allowed context without a spec — reported as info, not failure).

### Security Model: Exclusion by Omission

Tools are included only if:
1. The bounded context is in the allowlist
2. The context's `AgentTools()` function declares the spec
3. The agent permission ceiling (spec 155) allows the permission
4. The user has the permission

New endpoints are excluded until someone adds an `AgentToolSpec` entry. This is fail-closed.

## Checklist

- [x] Specification approved
- [x] `AgentTools() []AgentToolSpec` function in `capabilitymapping/publishedlanguage/`
- [x] `AgentTools()` in `architecturemodeling/publishedlanguage/`
- [x] `AgentTools()` in `enterprisearchitecture/publishedlanguage/`
- [x] `AgentTools()` in `valuestreams/publishedlanguage/`
- [x] `AgentTools()` in `metamodel/publishedlanguage/`
- [x] Context allowlist in archassistant with central collector
- [x] `registration.go` loads from context catalogs instead of manual registration
- [x] Architecture guard test: tool specs reference valid routes
- [x] Architecture guard test: tool specs reference valid permissions
- [x] Existing 24 tools migrated to context-owned specs (no behavior change)
- [x] Build passing
