# Event Constant Consistency in Subscriptions

## Description
Replace brittle event-type string literals in subscription wiring with published language constants. Event name constants are the one published language artifact (alongside ID types) that legitimately crosses the ACL boundary at runtime — a string constant is wire protocol identification, not domain model leakage.

## Architectural Context
Per spec 146 and 147, downstream BCs maintain ACL isolation by:
- Keeping locally owned structs for event deserialization (ACL boundary)
- NOT importing upstream payload DTOs (`publishedlanguage/contracts/`) in production code

Event name constants are different. They identify which event arrived on the wire so the ACL can dispatch to the right handler. Importing `cmPL.CapabilityCreated` (a string constant) creates no more coupling than the subscription itself — the downstream BC already knows this event exists because it subscribes to it. Making that subscription type-safe via a constant is a strict improvement over a duplicated string literal.

## Scope
- Subscription registration in API route setup and projector dispatch tables.
- All bounded contexts where event names are duplicated as raw string literals.

## Out of Scope
- Enforcing a single projector routing style (`switch` vs map) when behavior is equivalent.

## Required Changes

### 1) Subscription constants only
Use constants from the owning BC's published language package for all cross-BC event bus subscriptions and event-type comparisons.

### 2) Local event ownership
A BC should reference its own `publishedlanguage` event constants for internal subscriptions rather than duplicating literals. If no `publishedlanguage` package exists yet for an internally-only-consumed event, a local constant in the domain or application layer is acceptable.

### 3) Keep behavior unchanged
This is a contract-hardening refactor. Do not alter event routing semantics.

## Acceptance Criteria
- No raw event-name literals in subscription wiring where a published constant exists.
- Renaming a published constant produces compile-time breakage in subscribers.
- All affected tests pass.
- Only constants and ID types from upstream `publishedlanguage` are imported — never `publishedlanguage/contracts/` (enforced by spec 147 guardrails).

## Verification
- `go test ./internal -run Architecture`
- `go test ./internal/...`
- Focused search for event-name literals in route/projector subscription code.
