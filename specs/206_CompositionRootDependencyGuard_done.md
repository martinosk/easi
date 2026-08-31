# 206 — Composition-Root Dependency Guard

> **Status:** done
> **Depends on:** 207 (Direction-Derived Reads Owned by Architecture Direction), 208 (One-Pager Completeness Served by OnePagers) — the guard is green only once both have landed
> **Amends:** 134 (Architecture Guard Tests), 125/135 (Published Language)
> **Amended by:** 209 (Events-Only Context Integration) — declared composition-root bridges are no longer permitted; rule 9 and design decision 8 are superseded

---

## Problem Statement

`docs/architecture/README.md` states that the bounded-context dependency graph is acyclic and that downstream contexts keep local caches rather than querying upstream contexts at read time. The architecture tests enforce only that no context *imports* another's internals; files under `internal/infrastructure/` — the composition root — are exempt. Every context therefore looks clean while the composition root wires read models and services of one context into the ports of another, in both directions. A survey on 2026-08-29 found five runtime cycles (EA ⇄ Architecture Direction; AM, CM and EA ⇄ OnePagers; Access Delegation ⇄ Auth; Auth ⇄ Arch Assistant) and two compile-time cycles created by contracts published by the *consumer* (`AgentToolSpec` in Arch Assistant, import-gateway inputs in Importing). None of them is recorded in a canvas or the context map.

The dependency graph must be made acyclic, and a test must make every composition-root bridge explicit and reject any change that closes a cycle.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Backend Developer** | Learns from a failing test, not a review, that a new adapter creates an undeclared or cyclic dependency |
| **Architect / Reviewer** | Sees every cross-context runtime dependency in one declared, justified list |
| **End user (grantee, assistant user)** | Notices no change: edit grants still invite non-users; the assistant is still offered only when configured |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Composition-root dependency guard

  Scenario: Undeclared bridge fails the build
    Given a new file in the composition root that imports a read model of context A and a port of context B
    And no bridge declaration names that file
    When the architecture tests run
    Then they fail naming the file and the contexts it reaches

  Scenario: A declared bridge that closes a cycle fails the build
    Given context B already depends on context A through a published-language import
    When a bridge declaring A as consumer and B as supplier is added
    Then the architecture tests fail and print the cycle A → B → A

  Scenario: Stale or inaccurate declarations fail the build
    Given a declaration for a file that no longer bridges, or that omits a context the file imports
    When the architecture tests run
    Then they fail naming the declaration

  Scenario: The router only registers routes
    Given `router.go` imports a package of a context other than its `infrastructure/api` package
    When the architecture tests run
    Then they fail naming the import

  Scenario: Edit grant for a non-user still invites them
    Given an admin grants edit access to an email that has no user account
    When the grant is created
    Then an invitation for that email exists, exactly as before

  Scenario: Assistant is offered only when configured
    Given a user with the assistant permission in a tenant whose assistant is not configured
    When the user signs in
    Then the chat entry point is not shown
    And once a tenant admin configures the assistant, the chat entry point is shown on the next load
```

---

## Business Rules & Invariants

1. **Dependency edges** — context Y depends on context X when (a) any file of Y imports `X/publishedlanguage`, or (b) a declared composition-root bridge names Y as consumer and X as supplier. `shared/`, `infrastructure/` and `testing/` are not contexts.
2. **Acyclic** — the graph formed by all edges has no cycle. There is no allowlist for cycles.
3. **Router registers only** — `internal/infrastructure/api/router.go` imports, from contexts, only their `infrastructure/api` packages.
4. **Every other composition-root file that imports a context package is a bridge** and must have exactly one declaration listing its edges (consumer → supplier) with a one-line justification.
5. **Declarations are exact** — the set of contexts a declared file imports equals the set of contexts in its edges; a declaration for a file that no longer bridges is a failure.
6. **Event subscriptions are wired inside the consuming context** — never in the composition root.
7. **A contract implemented by many contexts is shared kernel** — it lives under `shared/`, not in the consumer's published language.
8. **A gateway input type belongs to the supplier's published language** — the context that executes the operation publishes the input shape; the caller imports it.
9. **Non-cyclic bridges may remain** — they are declared, justified and visible; the README's "local caches over shared state" is a preference, enforced only for cycles.

---

## Acceptance Criteria

- [x] `TestEveryCompositionRootBridgeIsDeclaredExactly` (with `TestRouterOnlyRegistersRoutes`, `TestNoStaleBridgeDeclarations`, `TestNonContextPackagesImportOnlyPublishedLanguage`, `TestProductionCodeDoesNotImportTestSupport`) fails on an undeclared bridge, on a declaration whose contexts differ from the file's imports, on a stale declaration, and on `router.go` importing anything but `infrastructure/api` packages of contexts.
- [x] `TestContextDependencyGraphIsAcyclic` builds the graph from published-language imports plus declared bridges and fails with the printed cycle on any cycle; it passes on the codebase after 207 and 208.
- [x] `AgentToolSpec`, `ParamSpec`, `AccessClass` and the param constructors live in `shared/agenttools`; no context imports `archassistant/publishedlanguage`.
- [x] Import-gateway input types live in the published language of Architecture Modeling and Capability Mapping; neither imports `importing/publishedlanguage`.
- [x] Access Delegation requests invitations for non-users through its own `InvitationRequester` port, implemented in the composition-root bridge (`accessdelegation_bridges.go`) by dispatching Auth's `CreateInvitation` command — Auth itself must not import the port; Auth no longer subscribes to `EditGrantForNonUserCreated`; the existing auto-invitation behaviour tests pass against the port. **Removed 2026-08-31** — `EditGrantForNonUserCreated` was never persisted to the event store (it was a direct in-process bus publish) and had zero subscribers; the auto-invitation behaviour lives entirely in the `EnsureInvitation` dispatch, so the event, its published-language constant, and the `EventBus` dependency it required were deleted.
- [x] Auth's session links `x-assistant` / `x-assistant-write` depend on permission only; Arch Assistant exposes `GET /assistant/status` (`{ configured }`) to holders of the assistant permission; the frontend shows the chat entry point when the link is present and the status is configured; Auth has no dependency on Arch Assistant.
- [x] Arch Assistant subscribes to `TenantCreated` inside its own route setup, not in the router.
- [x] Every cross-context adapter in the composition root lives in a `*_bridges.go` file declared in the bridge registry; `router.go` constructs no adapters.
- [x] `docs/architecture/README.md` states the enforced rule and links the declared-bridge registry; `docs/backend/cross-context-events.md` "Query-Based Integration" is replaced by the declared bridges; the canvases of Access Delegation (README row), Arch Assistant and Auth reflect the new directions.
- [x] `go test ./...` green including the two new tests.

---

## Architecture

### Ownership

The rule and the tests live in `backend/internal` next to the existing architecture tests. Each hygiene change is owned by the context it touches: `shared/agenttools` (new shared kernel package), Architecture Modeling and Capability Mapping (gateway input types), Access Delegation (`InvitationRequester` port, implemented in the declared bridge), Auth (session links), Arch Assistant (`/assistant/status`, `TenantCreated` subscription).

### Domain Model

No aggregate changes. Access Delegation gains a consumer-defined port `InvitationRequester` invoked after a grant for a non-user is created. **Removed 2026-08-31**: the published event `EditGrantForNonUserCreated` was deleted — it was never persisted and had no subscribers; the auto-invitation behaviour lives entirely in the `EnsureInvitation` dispatch.

### API Surface

New: `GET /assistant/status` on Arch Assistant, permission `assistant:use`. Changed: `GET /auth/sessions/current` `_links.x-assistant` and `x-assistant-write` are present whenever the user holds the assistant permission (previously also gated on configuration).

### Persistence

None.

### Frontend

The chat entry point is shown when the session carries `x-assistant` and `GET /assistant/status` reports configured; the status query is invalidated when the assistant configuration is saved.

### Cross-Context Integration

- Access Delegation → Auth (bridge: user lookup, invitation check, domain allowlist, invitation request). Auth → Access Delegation: **removed**.
- Auth → Arch Assistant: **removed**. Arch Assistant → Auth, → Platform (events): unchanged and now wired inside Arch Assistant.
- All contexts → Arch Assistant (contract import): **removed** via `shared/agenttools`.
- Architecture Modeling / Capability Mapping → Importing (contract import): **removed**; Importing → Architecture Modeling / Capability Mapping / Value Streams remain as declared bridges.
- Declared, non-cyclic bridges after 207/208: Access Delegation ← CM, AM, AV, Auth; Architecture Direction ← CM, AM; Enterprise Architecture ← CM; OnePagers ← CM, AM, AD, EA, MM; Importing ← AM, CM, VS; Capability Mapping ← AM (component gateway).

---

## Design Decisions

1. **Direction is declared, not inferred** — an adapter file cannot reveal which side is the port and which the implementation; a declaration reviewed with the code is the honest mechanism, mirroring the existing (empty) import allowlist. Alternative: static analysis of interface satisfaction (rejected: fragile, and still blind to closures such as `existsByID(readModel.GetByID)`).
2. **Cycles are fixed, never allowlisted** — user decision 2026-08-29; an exception list would restore the blind spot the guard exists to remove.
3. **Access Delegation asks Auth for the invitation** — flipping the auto-invitation from an Auth-side event reaction to an Access-Delegation-side port removes the cycle without three event-fed caches. Alternative: Access Delegation caching users, invitations and allowed domains from Auth events (rejected: three backfilled caches to serve a write-time validation).
4. **Assistant availability is the assistant's resource** — the session says *may use*; the assistant says *is ready*. Alternative: Auth caching the assistant's configuration status from a new Arch Assistant event (rejected: keeps Auth → Arch Assistant while Arch Assistant already depends on Auth).
5. **`AgentToolSpec` is shared kernel** — one contract implemented by seven contexts is not any single context's language. Alternative: each context defining its own tool type and Arch Assistant adapting (rejected: seven copies of one shape).
6. **Gateway inputs move to the supplier** — the executing context publishes what it accepts, so the caller depends on the callee, matching the runtime direction.
7. **`router.go` registers, bridges wire** — one file per consuming context keeps each declaration reviewable and lets the test compute "contexts reached" per file.
8. **Non-cyclic bridges stay declared rather than forced into caches** — scope stays on acyclicity; the registry makes the remaining query-time reads visible for later decisions.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Declared direction can be mis-declared | A wrong declaration could hide a cycle | Declarations live beside the adapters in one reviewed file; the acyclicity test prints the full graph on failure |
| `x-assistant` link no longer implies "configured" | One extra status request when the link is present | Tiny resource, cached; assistant-owned readiness is the more accurate contract |
| Event `EditGrantForNonUserCreated` has no in-process subscriber | Audit-only event | **Removed 2026-08-31** — the event was never persisted and had no subscribers; the auto-invitation behaviour lives in the `EnsureInvitation` dispatch |

---

## Checklist

- [x] Specification ready — approved in session 2026-08-29 ("make sure it cannot happen again", no cycle allowlist)
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off (2026-08-31)
