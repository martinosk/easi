# Agent Permission Ceiling

**Status**: done

**Series**: Architecture Assistant Evolution (1 of 6)
- **Spec 155**: Agent Permission Ceiling (this spec)
- Spec 156: Generic Tool Executor
- Spec 157: Tool Catalog per Bounded Context
- Spec 158: Domain Knowledge Injection
- Spec 159: Expand Tool Coverage
- Spec 160: Agent Audit Events

## User Value

> "As a platform operator, I need confidence that the AI assistant cannot access user management, access delegation, or audit APIs — even when an admin uses it — so that AI-assisted workflows cannot escalate privileges."

## Problem

The agent inherits the user's full permission set. An admin's agent has `users:manage`, `invitations:manage`, `edit-grants:manage`, and `audit:read`. Today this is masked because no tools exist for those contexts. When we move to auto-generated tools (spec 157), this becomes the primary vulnerability.

## Solution

Introduce an **agent permission ceiling** — a hard-coded allowlist of permissions the agent can ever hold. The effective permission for any tool call is `intersection(user.Permissions, agentCeiling)`.

### Agent Permission Ceiling

```go
var agentPermissionCeiling = map[string]bool{
    "components:read":       true,
    "components:write":      true,
    "capabilities:read":     true,
    "capabilities:write":    true,
    "domains:read":          true,
    "domains:write":         true,
    "valuestreams:read":     true,
    "valuestreams:write":    true,
    "enterprise-arch:read":  true,
    "enterprise-arch:write": true,
    "views:read":            true,
    "metamodel:read":        true,
    "assistant:use":         true,
}
```

Permanently excluded: `users:manage`, `invitations:manage`, `edit-grants:manage`, `audit:read`, `metamodel:write`, `importing:write`.

### Enforcement Point

In `buildToolContext` (`conversation_handlers.go`), wrap the actor permissions:

```go
type agentScopedPermissions struct {
    actor sharedctx.Actor
}

func (a *agentScopedPermissions) HasPermission(permission string) bool {
    return a.actor.HasPermission(permission) && agentPermissionCeiling[permission]
}
```

### Risk-Based Write Limits

Replace flat `maxSameToolCalls = 500` with access-class-based limits:

| Access Class | Max per message | Rationale |
|---|---|---|
| Read | 500 | Reading 100+ entities is normal exploration |
| Create | 50 | Bulk creation is rare in conversation |
| Update | 100 | Bulk updates are rare in conversation |
| Delete | 5 | Destructive, should be deliberate |

This requires `AccessClass` to differentiate create/update/delete (currently only read vs write). Add `AccessCreate`, `AccessUpdate`, `AccessDelete` as sub-categories of `AccessWrite`.

## Checklist

- [x] Specification approved
- [x] `agentPermissionCeiling` defined as package-level constant
- [x] `agentScopedPermissions` wrapper implemented
- [x] `buildToolContext` uses wrapper instead of raw actor permissions
- [x] `AccessClass` expanded: `AccessRead`, `AccessCreate`, `AccessUpdate`, `AccessDelete`
- [x] Per-access-class tool call limits in orchestrator
- [x] Existing tools annotated with correct access sub-class
- [x] Unit test: admin actor filtered to ceiling
- [x] Unit test: delete tool capped at 5 calls per message
- [x] Build passing
