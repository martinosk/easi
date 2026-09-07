# 210 — Relocate EnterpriseArchitecture into Architecture Direction

> **Status:** done
> **Depends on:** — (design: [docs/specs/enterprise-capability.md](../docs/specs/enterprise-capability.md))
> **Roadmap alignment:** SD1 / H1-1

---

## Problem Statement

EnterpriseArchitecture is a two-aggregate fragment whose language belongs to Architecture Direction (design doc, Problem Statement). This slice performs the mechanical merge — the Platform→Auth playbook — with zero behaviour change: same endpoints, same events on the wire, same UI. It exists so that the remodelling slices (211–213) are in-context work.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Everything works exactly as before the deployment |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Behaviour-preserving relocation

  Scenario: Enterprise capabilities remain fully operable
    Given enterprise capabilities existed before the deployment
    When an architect lists, creates, updates, or deletes an enterprise capability
    Then behaviour, URLs, and permissions are identical to before

  Scenario: One-pagers for enterprise capabilities are unaffected
    Given an enterprise capability with recorded one-pager facts
    When its name changes or it is deleted
    Then the subject index updates and facts archive exactly as before

  Scenario: Direction reactions still fire
    Given an enterprise capability with an active direction
    When the enterprise capability is deleted
    Then the active direction is rejected as before

  Scenario: TIME suggestions still serve
    When an architect opens the TIME suggestions tab
    Then suggestions compute and render as before
```

---

## Business Rules & Invariants

1. **No behaviour change** — every route, permission (`enterprise-arch:*`), response shape, event type string, and UI flow is unchanged.
2. **Ownership moves whole** — all EA aggregates, events, read models, routes, and agent tools become Architecture Direction code; the `enterprisearchitecture` schema and Go package cease to exist.
3. **Stored events are untouched** — the global event store is not modified; only read-model tables re-parent.
4. **Redundant cache dissolves** — AD's `enterprise_capability_cache` (a cache of now-local data) is dropped in favour of the relocated `enterprise_capabilities` read model; the EC-deleted reaction becomes in-context.

---

## Acceptance Criteria

- [x] `backend/internal/enterprisearchitecture/` is deleted; all its behaviour lives under `backend/internal/architecturedirection/`
- [x] Read-model tables re-parent via `SET SCHEMA` migration; `enterprise_capability_cache` and its projector are removed; `DROP SCHEMA enterprisearchitecture` succeeds; DB user provisioning updated
- [x] OnePagers subscribes to the same event type strings via `adPL` constants (import swap only)
- [x] Agent tools formerly in `eaPL` are served from AD's published language; the assistant catalog is unchanged for users
- [x] All architecture guard tests pass without modification; `components.csv` regenerated; no shipped migration is edited
- [x] Context map, canvases (EnterpriseArchitecture.md folded into ArchitectureDirection.md), `cross-context-events.md`, and `INDEX.md` updated

---

## Architecture

### Ownership

Architecture Direction absorbs everything. OnePagers is the only external consumer touched (import-constant swaps).

### Domain Model

Unchanged in this slice; aggregates relocate verbatim.

### API Surface

Unchanged; route registration moves into AD's setup (the composition/maturity routes under `/enterprise-capabilities/…` already live there).

### Persistence

One migration, 150: `ALTER TABLE enterprisearchitecture.* SET SCHEMA architecturedirection` for `enterprise_capabilities`, `enterprise_strategic_importance`, `ea_importance_cache`, `ea_fit_score_cache`, `ea_strategy_pillar_cache`, `ea_realization_cache`, `domain_capability_metadata`, `business_domain_name_cache`; `DROP TABLE architecturedirection.enterprise_capability_cache`; `DROP SCHEMA enterprisearchitecture`.

No earlier migration is touched. Migrations run once, in order, recorded by filename in `schema_migrations`: on a fresh database 138/144/148 execute while `enterprisearchitecture` still exists, and on an existing database they have already executed. Verified by running 001–150 against an empty database and comparing the end state.

### Frontend

No changes; API paths are stable.

### Cross-Context Integration

Published constants relocate to `architecturedirection/publishedlanguage` with byte-identical strings. OnePagers' three EA subscriptions swap constants. Nothing else in the graph changes; contexts go 15 → 14.

---

## Design Decisions

1. **Follow the Platform→Auth playbook exactly** — proven mechanics: `SET SCHEMA`, stable event strings, dynamic guards (design doc decisions 1–2). Alternative — thin EA before moving — rejected: double contract churn.
2. **Drop the EC cache now, not later** — keeping a cache of local data would violate the pattern that caches exist only for upstream data. Its three consumers (composition service, reference checker, composition handlers) read the relocated `EnterpriseCapabilityReadModel` instead; `MaturityAnalysisReadModel` reads `architecturedirection.enterprise_capabilities`.
3. **Shipped migrations are immutable; tests assert the target database state** — three integration tests read a historical migration file and execute it against the current database. That only ever worked while no schema had moved, and it asserts a database state that no longer exists. The tests are deleted rather than the migrations bent to keep them running; migrations 138, 144 and 148 stay exactly as they shipped.
4. **Enterprise capability routes stay a separate setup function inside AD's API package** — `setupEnterpriseCapabilityRoutes` is called from `SetupRoutes` with the shared `RoutesDeps` (extended with `SessionProvider`), so the composition root registers one Architecture Direction entry point. Route patterns, permissions and HATEOAS base paths are unchanged.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Wholesale move in one change | Large mechanical diff | No behaviour change; guards and full test suite verify; precedent commit as checklist |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
