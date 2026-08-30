# 207 — Direction-Derived Reads Owned by Architecture Direction

> **Status:** ongoing
> **Depends on:** 172 (Direction Is the Association), 136 (EA Read Model ACL Decoupling)
> **Amends:** 172 (bounded-context ownership table), 178 is amended by 208
> **Amended by:** 209 (Events-Only Context Integration) — the composition-root lookups into Capability Mapping and Architecture Modeling are replaced by Architecture Direction's own reference and realization caches

---

## Problem Statement

Since spec 172, the association between an enterprise capability and its domain capabilities is the `Direction.sources` set owned by Architecture Direction. Everything that is *derived* from that association — the composition of an enterprise capability, source eligibility, composition preview, the included-capability and domain counts on the enterprise capability DTO, and the maturity-gap analysis — still lives in the Enterprise Architecture context. To compute it, EA reads Architecture Direction's read model through a query-time adapter in the composition root, while Architecture Direction reacts to EA's `EnterpriseCapabilityDeleted` event and checks enterprise-capability existence through EA's read model.

The result is a bidirectional runtime dependency EA ⇄ Architecture Direction. It contradicts the architecture README ("dependency graph is acyclic", "local caches over shared state"), it is invisible to the existing architecture tests (they check package imports, and the composition root is exempt), and it is exactly the shape the dependency guard in spec 206 forbids. The data that composition needs — direction sources — is Architecture Direction's own; the reads should live there.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Sees an enterprise capability's composition, counts and maturity gaps exactly as today |
| **Domain Architect** | Captures and edits directions with the same eligibility checks and composition preview as today |
| **Platform Operator** | Deploys the change to existing tenants without a period of empty compositions or zero counts |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Composition, source candidates and maturity analysis served by Architecture Direction

  Scenario: Composition view is unchanged
    Given an enterprise capability with an active direction whose sources have subtrees in two business domains
    When the architect opens the enterprise capability's composition
    Then the included capabilities, roles, carve-outs, domain grouping and counts are identical to the pre-207 response

  Scenario: Enterprise capability list still shows counts
    Given three enterprise capabilities, two with active directions
    When the architect opens the enterprise capabilities table
    Then each row shows its included-capability count and domain count
    And an enterprise capability without an active direction shows 0 and 0

  Scenario: Enterprise capability detail still shows counts
    When the architect opens an enterprise capability's detail panel
    Then the included-capability count and domain count are shown

  Scenario: Source eligibility still rejects a capability claimed by another active direction
    Given capability C is an explicit source of an active direction on enterprise capability X
    When the architect captures a direction on enterprise capability Y with C as a source
    Then the request is rejected with 409 naming C and X, exactly as before

  Scenario: Composition preview is unchanged
    When the architect previews a proposed source set in the capture form
    Then the included capabilities and eligibility per source are identical to the pre-207 response

  Scenario: Maturity analysis is unchanged
    Given an enterprise capability with a target maturity and included capabilities below it
    When the architect opens the maturity analysis tab and the gap detail
    Then candidates, distributions and investment priorities are identical to the pre-207 response

  Scenario: Renaming, re-parenting or deleting a domain capability is reflected in composition
    Given capability C is included in enterprise capability X's composition
    When C is renamed, moved under a different parent, or deleted in Capability Mapping
    Then X's composition reflects the change after the event is projected

  Scenario: Changing an enterprise capability is reflected in Architecture Direction
    When an enterprise capability is renamed, deleted, or given a new target maturity
    Then source-candidate conflicts name it correctly, its direction is rejected on deletion, and the maturity gap uses the new target

  Scenario: Existing tenants see complete data immediately after deployment
    Given a tenant with capabilities, domains, enterprise capabilities and directions created before this change
    When the migrations have run and the backend starts
    Then compositions, counts and maturity analysis are complete without waiting for any event
```

---

## Business Rules & Invariants

1. **Composition is derived from Architecture Direction's own data** — the composition algorithm (172 R2) runs over Architecture Direction's direction sources, its capability-node cache and its enterprise-capability cache. It never reads another context's tables or read models.
2. **Enterprise Architecture has no dependency on Architecture Direction** — no import, no event subscription, no composition-root bridge with EA as consumer and AD as supplier.
3. **The enterprise capability DTO carries no direction-derived fields** — `includedCapabilityCount` and `domainCount` are removed from it. They are served by Architecture Direction as a collection of composition summaries, one per enterprise capability.
4. **URLs and response shapes of moved endpoints do not change** — `/enterprise-capabilities/{id}/composition`, `/capabilities/source-candidates`, `/enterprise-capabilities/maturity-analysis`, `/enterprise-capabilities/{id}/maturity-gap` keep their paths, DTOs, permissions and HATEOAS links.
5. **Caches are eventually consistent and always backfilled** — both new caches are populated by projectors on the upstream published events and by a one-time backfill migration from the upstream tables, so a deployment never starts with empty caches.
6. **Capability-node cache mirrors Capability Mapping's hierarchy semantics** — level, parent, L1 ancestor, effective business domain (the L1's domain, inherited by the subtree) and maturity value follow the same rules the EA metadata cache follows today.
7. **Existence and activity of an enterprise capability are answered from the local cache** — Architecture Direction's reference checks for enterprise capabilities read its own cache, not EA's read model.
8. **Composition preview remains stateless** and runs the same algorithm as the persisted composition.

---

## Acceptance Criteria

- [x] `GET /enterprise-capabilities/{id}/composition`, `GET /capabilities/source-candidates`, `GET /enterprise-capabilities/maturity-analysis`, `GET /enterprise-capabilities/{id}/maturity-gap` and `POST /enterprise-capabilities/{id}/direction/composition-preview` are served by Architecture Direction with unchanged paths, DTOs, permissions and links; their existing handler and read-model tests pass in the new location.
- [x] `GET /enterprise-capability-compositions` returns one summary per enterprise capability (`enterpriseCapabilityId`, `sourceCount`, `includedCount`, `domainCount`, `hasActiveDirection`, `directionStatus`) under the same read permission as the composition view.
- [x] The enterprise capability list and detail DTOs no longer contain `includedCapabilityCount` or `domainCount`; the frontend table and detail panel show the counts from the composition summaries / composition meta.
- [x] Capturing a direction with a source already claimed by another active direction returns 409 with the conflicting capability and enterprise capability — verified by the existing handler test, now exercising Architecture Direction's local eligibility.
- [x] Architecture Direction's `capability_node_cache` is maintained by a projector for `CapabilityCreated/Updated/Deleted/ParentChanged/LevelChanged/AssignedToDomain/UnassignedFromDomain/MetadataUpdated` and `BusinessDomainUpdated` — mirroring EA's proven metadata cache projector, which likewise reacts to `BusinessDomainUpdated` only (a domain's name can change after creation; nothing needs projecting at the moment it is created, before any capability is assigned to it) — with projector unit tests per event and an integration test for hierarchy recalculation.
- [x] Architecture Direction's `enterprise_capability_cache` is maintained by a projector for `EnterpriseCapabilityCreated/Updated/Deleted/TargetMaturitySet`, with projector unit tests.
- [x] Backfill migrations populate both caches from `capabilitymapping.*` and `enterprisearchitecture.enterprise_capabilities` idempotently; an integration test runs the backfill against seeded source tables and asserts the cache contents.
- [x] `enterprisearchitecture` contains no reference to Architecture Direction and no composition, eligibility or maturity-analysis code; `go test ./...` and the spec 206 dependency guard show no EA → AD edge.
- [x] OnePagers' enterprise-capability relation field ("composed capabilities") is sourced from Architecture Direction's composition.
- [x] Agent tools for composition, source candidates and maturity analysis are contributed by Architecture Direction's published language; the tool catalog test passes.
- [x] Swagger regenerated; frontend `npm run build`, `npm test -- --run`, `npm run lint` green; MSW spec-172 stubs updated to the new contract.

---

## Architecture

### Ownership

Architecture Direction owns everything derived from direction sources: composition, source eligibility, composition preview, composition summaries, maturity analysis. Enterprise Architecture keeps the enterprise capability aggregate, target maturity, strategic importance, TIME suggestions and its existing ACL caches. OnePagers consumes composition from Architecture Direction through its existing relation-field port.

### Domain Model

No aggregate changes. The composition resolver (domain service) and composition service (application service) move to Architecture Direction unchanged in behaviour. Two local caches are added to Architecture Direction:

- **Capability node cache** — one row per domain capability: name, level, parent, L1 ancestor, effective business domain (id + name), maturity value. Domain names come from Architecture Direction's existing reference-name cache, which already tracks `BusinessDomainCreated/Updated`.
- **Enterprise capability cache** — one row per enterprise capability: name, active flag, target maturity.

The `ReferenceChecker`'s enterprise-capability existence and activity checks, `SourceEligibility` and `CompositionPreviewProvider` become in-context implementations; the corresponding composition-root adapters are deleted.

### API Surface

Moved without change: composition, source candidates, maturity analysis, maturity gap, composition preview. New: `GET /enterprise-capability-compositions` (composition summaries collection). Changed: enterprise capability list/detail DTO loses `includedCapabilityCount` and `domainCount`; `x-composition` stays. Permissions unchanged: the moved reads keep the permission constants they have today.

### Persistence

Two tables in the `architecturedirection` schema with the standard tenant RLS policy, plus two idempotent backfill migrations (`INSERT … SELECT … ON CONFLICT DO UPDATE`) from `capabilitymapping.capabilities`, `capabilitymapping.domain_capability_assignments`, `capabilitymapping.business_domains` and `enterprisearchitecture.enterprise_capabilities`. The legacy `link_count`/`domain_count` columns on `enterprisearchitecture.enterprise_capabilities` were already dropped by migration 120 (spec 172); there is nothing left to migrate off of them.

### Frontend

Enterprise capability table reads counts from the composition summaries collection (new query hook, query key, invalidated by direction mutations). Detail panel reads counts from the composition meta it already loads. Maturity analysis tab is unaffected (its counts come from the analysis DTO). Spec-172 MSW stubs updated. Source-candidate and preview calls are unchanged.

### Cross-Context Integration

- Architecture Direction ← Capability Mapping: `Capability*` and `BusinessDomain*` events (already subscribed) now also feed the capability node cache.
- Architecture Direction ← Enterprise Architecture: `EnterpriseCapability*` events (already subscribed for deletion) now also feed the enterprise capability cache.
- Enterprise Architecture → Architecture Direction: **removed** (no adapter, no subscription).
- OnePagers ← Architecture Direction: composition for the enterprise-capability relation field (declared bridge per spec 206).

---

## Design Decisions

1. **Move the derived reads, not the aggregate** — Architecture Direction is downstream of the enterprise capability's identity (a direction needs an existing, active EC) and upstream of everything derived from sources. Moving the reads leaves aggregate ownership and event streams untouched. Alternative: moving the `EnterpriseCapability` aggregate into Architecture Direction (rejected: dissolves EA into AD, relocates 69 files and all importers, and forces a context rename for no behavioural gain).
2. **Composition stays on the fly** — 172 O5 chose per-request computation over a projected composition table; volumes have not changed. Alternative: a materialised composition read model kept by a projector (rejected: adds a second derived table and replay logic without a demonstrated need).
3. **Counts become a separate collection resource** — the counts are direction-derived, so they leave the EC DTO. One collection call serves the table; the detail panel already loads the composition meta. Alternative: keeping the fields on the EC DTO via an AD-published "composition changed" event cached by EA (rejected: EA → AD dependency reintroduces the cycle).
4. **Architecture Direction keeps its own capability-node cache** — EA still needs its metadata cache for TIME suggestions, and the README principle is one local cache per consuming context. Alternative: sharing EA's cache table (rejected: cross-context table access, exactly what 136/137 removed).
5. **Existence checks for enterprise capabilities read the local cache** — the cache is required for composition anyway, so the EA read-model bridge is removed rather than merely declared.
6. **Permissions are preserved constant-for-constant** — the moved endpoints keep the permission they enforce today, so no role gains or loses access.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Two more caches in Architecture Direction | More projector code and one more table to backfill | Projectors mirror EA's proven metadata projector; backfill is idempotent and integration-tested |
| Counts leave the EC DTO | One extra request on the EC table | Single collection call, cached by React Query, invalidated by direction mutations |
| Eventual consistency for names/maturity in composition | Sub-second lag between a rename in CM and the composition view | Same characteristic EA's caches already have (136 Part 5) |

---

## Checklist

- [x] Specification ready — decision W approved in session 2026-08-29
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
