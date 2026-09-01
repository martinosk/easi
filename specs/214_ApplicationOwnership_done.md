# 214 — Application Ownership

> **Status:** pending
> **Depends on:** —
> **Roadmap alignment:** SD6 / H1-2

---

## Problem Statement

An ApplicationComponent has no owner. Experts are free-text value objects, unconnected to real users or teams, so nobody can answer "who is accountable for this application?" or "which applications are orphaned?". This blocks stewardship work and the ServiceNow trust model (SD3), which depends on an Unknown/orphaned state that auto-registered applications can land in safely.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | See which applications lack an accountable owner; drive the orphan count down |
| **Application Owner** | Be recorded as accountable for the applications they own |
| **Admin** | Confirm or correct ownership claims |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Application ownership

  Scenario: New applications are orphaned by default
    Given an architect creates an application component
    Then its ownership state is "Unknown"

  Scenario: Nominating an owner
    Given an application in state "Unknown"
    When an architect nominates a user as its owner
    Then the ownership state is "Nominated" and the candidate is shown

  Scenario: Confirming a nominated person
    Given an application in state "Nominated" with a user candidate
    When an architect confirms the nomination
    Then the ownership state is "Owned"

  Scenario: Confirming a nominated team
    Given an application in state "Nominated" with an internal-team candidate
    When an architect confirms the nomination
    Then the ownership state is "Managed"

  Scenario: Assigning an owner directly
    Given an application in state "Unknown"
    When an architect assigns an internal team as its owner
    Then the ownership state is "Managed" without passing through "Nominated"

  Scenario: Clearing ownership
    Given an application in state "Owned"
    When an architect clears its ownership
    Then the ownership state is "Unknown" and no owner is shown

  Scenario: Owning team is deleted
    Given an application in state "Managed" by team "Platform Ops"
    When "Platform Ops" is deleted
    Then the application's ownership state reverts to "Unknown"

  Scenario: Ownership statistics
    Given applications in a mix of ownership states
    When an architect opens the application list
    Then counts per ownership state are shown, including the orphan count
```

---

## Business Rules & Invariants

1. **Single state** — every ApplicationComponent has exactly one ownership state: `unknown`, `nominated`, `owned`, or `managed`; components start as `unknown`.
2. **Typed owner reference** — the owner is a reference to an Auth user or an ArchitectureModeling InternalTeam, never free text.
3. **Nomination carries the candidate** — entering `nominated` requires an owner reference; confirming resolves to `owned` (user reference) or `managed` (team reference) without changing the reference.
4. **Direct assignment** — an owner reference may be assigned without nomination, resolving directly to `owned` or `managed`.
5. **Clearing** — clearing ownership removes the reference and returns the state to `unknown`, from any state.
6. **Deleted team reverts ownership** — when a referenced InternalTeam is deleted, affected components revert to `unknown`.
7. **HATEOAS-driven transitions** — the transitions available to the caller are exposed as affordance links on the component; the UI derives validity only from links, never from a client-side state table.

---

## Acceptance Criteria

- [ ] Component responses and list rows carry ownership state and resolved owner display name
- [ ] Nominate, confirm, assign, and clear operate per rules 1–5, each raising its own domain event
- [ ] Team deletion reverts ownership per rule 6 without any cross-context call
- [ ] An ownership statistics query returns counts per state for the tenant
- [ ] Existing components read as `unknown` after deployment with no data loss
- [ ] Affordance links appear exactly when the corresponding transition is legal for the actor

---

## Architecture

### Ownership

ArchitectureModeling owns the change entirely. Auth is upstream for user identity (display names via published `UserCreated` events, mirroring the Capability Mapping user-name cache); InternalTeam is AM-local.

### Domain Model

Ownership state and owner reference live on the ApplicationComponent aggregate (no independent lifecycle warrants a separate aggregate). New events: `ApplicationOwnerNominated`, `ApplicationOwnershipConfirmed`, `ApplicationOwnerAssigned`, `ApplicationOwnershipCleared`. Aggregates with no ownership events read as `unknown` — no upcasters. The team-deletion revert is an AM-internal reactor on its own `InternalTeamDeleted` event.

### API Surface

Ownership transitions as sub-resource operations on the component, gated by `components:write`; statistics as a read endpoint under `components:read`. Affordances per rule 7.

### Persistence

Ownership columns on the component read model plus an owner-name resolution via a new AM user-name cache (projector on `authPL.UserCreated`, seeded by a backfill migration — the `capabilitymapping.user_names` pattern). Backfill defaults every existing component to `unknown`.

### Frontend

Components feature: ownership section on the details panel, state facet on the list, statistics tile above the list. Mutation effects invalidate component list, detail, and statistics queries.

### Cross-Context Integration

Consumes `authPL.UserCreated` into a local name cache. Publishes the four ownership events for future read sides; no context consumes them in this slice.

---

## Design Decisions

1. **Ownership on the component aggregate, not a separate aggregate** — the state has no lifecycle apart from the component and its invariants are component-local. Alternative: standalone Ownership aggregate (rejected — adds identity and coordination without independent behavior).
2. **`owned` vs `managed` resolved by reference kind** — one nomination/confirmation flow; the distinction the roadmap requires falls out of whether a person or team is referenced. Alternative: separate command sets per kind (rejected — doubles the surface for no expressiveness).
3. **Reference validity checked at the command handler via read models** — user existence via the name cache, team existence via AM's own read model, per the cross-aggregate-invariant rule. DB constraints are a backstop only.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Owner display names from a local cache | Eventual consistency for brand-new users | Cache is projected in the write transaction; backfill covers pre-deploy users |
| Team revert via reactor | Momentary window where a deleted team is still referenced | Reactor runs in the deleting transaction (in-tx projection) |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
