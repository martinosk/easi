# 166 — Logical Capability Rename

> **Status:** ongoing
> **Depends on:** —
> **Coordinates with:** [167 — Direction on a Logical Capability](167_Direction_Aggregate_Capture_pending.md) (not yet drafted; this rename is its prerequisite vocabulary)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md), [`mockups/architecture-direction-ddd.md`](../mockups/architecture-direction-ddd.md)

---

## Problem Statement

The conceptual model spec establishes that what EASI today calls an **Enterprise Capability** is, properly named, a **Logical Capability** — an abstract grouping that spans physical capabilities living in business domains. The "Enterprise Capability" name conflates two genuinely different activities (physical consolidation and logical grouping) that the next slices need to keep apart.

The architecture group already speaks the new vocabulary. The screens and code do not. Every alignment conversation that references a screenshot today requires an inline translation step. Every later slice that ships against the new model has to inline-explain the rename. The cheapest way to clear both costs is to do the rename once, alone, before any new behaviour rides along.

The rename is purely structural — no new fields, no new flows, no new endpoints. The risk is not in the design but in the surface area: ~74 backend Go files, ~17 frontend TypeScript files, six event types, three database tables, and twenty-plus Swagger annotations. The spec exists to make the surface area explicit so the work doesn't sprawl.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | The vocabulary in the tool matches the vocabulary used in architecture-group conversations; a screenshot referenced in a meeting reads identically to the discussion. |
| **Domain Owner / Product Manager (downstream consumer)** | When the architecture group says "this is a Logical Capability decision," they can find it in the tool by that name. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Logical Capability Rename

  Scenario: Existing capabilities appear under the new label
    Given the system has capabilities created before the rename
    When I open the capabilities listing
    Then the navigation label reads "Logical Capabilities"
    And the page title reads "Logical Capabilities"
    And every capability that was previously visible is still listed
    And every existing mapping, strategic-importance entry, and target-maturity setting is preserved

  Scenario: API routes serve under the new path only
    Given the rename is deployed
    When the frontend calls GET /api/v1/logical-capabilities
    Then it returns the same set of capabilities as before
    And the response shape matches the pre-rename response, with field names updated where they referenced "enterprise"
    When any client calls GET /api/v1/enterprise-capabilities
    Then the response status is 404 Not Found

  Scenario: Event replay produces identical projections
    Given the event store contains historical EnterpriseCapability* events written before the rename
    When projections are rebuilt from the event store from scratch
    Then the projected read models match the pre-rename projections row-for-row, modulo column-name renames

  Scenario: User-visible strings are renamed everywhere
    Given the rename is deployed
    When an architect navigates the application end-to-end
    Then the strings "Logical Capability" / "Logical Capabilities" appear wherever the old terms used to surface
    And the strings "Enterprise Capability" / "Enterprise Capabilities" do not appear in any user-facing surface
    And the strings "Enterprise Capability Link" / "linked" in the link-management UI read as "mapping" / "mapped"

  Scenario: Existing flows behave identically end-to-end
    Given the rename is deployed
    When I create a logical capability
    Then it persists with the same fields and behaviour as a pre-rename Enterprise Capability
    When I map a domain (physical) capability to it
    Then the mapping is recorded with the same semantics that linking previously carried
    When I set target maturity, set strategic importance, or fetch a maturity-gap analysis
    Then each surface persists and reads back with the same fields and behaviour as before

  Scenario: A consumer of the cross-context published language receives the renamed events
    Given the capabilitymapping context subscribes to logical-capability events
    When a logical capability is created or mapped after the rename
    Then the subscriber observes events of types LogicalCapabilityCreated and LogicalCapabilityMapped
    And subscribers that previously consumed the old type names receive equivalent events through the upcaster bridge
```

---

## Business Rules & Invariants

1. **The rename is structural, not behavioural.** No business rule on the underlying concept changes. Every existing invariant on Enterprise Capability carries forward to Logical Capability without modification.
2. **Aggregate renames.**
   - `EnterpriseCapability` → `LogicalCapability`
   - `EnterpriseCapabilityLink` → `LogicalCapabilityMapping`
   - `EnterpriseStrategicImportance` → `StrategicImportance` (the "Enterprise" prefix is redundant — the type is scoped by foreign key)
3. **Value object renames.** `EnterpriseCapabilityID` → `LogicalCapabilityID`. `EnterpriseCapabilityName` → `LogicalCapabilityName`. `EnterpriseCapabilityLinkID` → `LogicalCapabilityMappingID`. Branded TypeScript ID types follow the same renames.
4. **Domain event renames.** Past-tense names follow the renamed aggregates:
   - `EnterpriseCapabilityCreated` → `LogicalCapabilityCreated`
   - `EnterpriseCapabilityUpdated` → `LogicalCapabilityUpdated`
   - `EnterpriseCapabilityDeleted` → `LogicalCapabilityDeleted`
   - `EnterpriseCapabilityLinked` → `LogicalCapabilityMapped`
   - `EnterpriseCapabilityUnlinked` → `LogicalCapabilityUnmapped`
   - `EnterpriseCapabilityTargetMaturitySet` → `LogicalCapabilityTargetMaturitySet`
   - The strategic-importance events follow analogous renames: `EnterpriseStrategicImportanceSet/Updated/Removed` → `StrategicImportanceSet/Updated/Removed`
5. **Event-store backward compatibility via upcasters.** Historical events on disk keep their old type names. Upcasters in the published-language deserialization layer translate `EnterpriseCapability*` and `EnterpriseStrategicImportance*` types to their renamed equivalents at read time. The event store is **not** rewritten.
6. **Database schema renames in place via SQL migration.**
   - Table `enterprisearchitecture.enterprise_capabilities` → `enterprisearchitecture.logical_capabilities`
   - Table `enterprisearchitecture.enterprise_capability_links` → `enterprisearchitecture.logical_capability_mappings`
   - Table `enterprisearchitecture.enterprise_strategic_importance` → `enterprisearchitecture.strategic_importance`
   - All indexes, foreign keys, and column references rename to match
   - The migration is forward-only per EASI database conventions (the project does not maintain down migrations); a corrective follow-up migration would be issued if the rename ever needs to be undone
7. **API routes rename without aliases or redirects.**
   - `/api/v1/enterprise-capabilities` → `/api/v1/logical-capabilities`
   - `/api/v1/enterprise-capabilities/{id}/links` → `/api/v1/logical-capabilities/{id}/mappings`
   - `/api/v1/enterprise-capabilities/{id}/strategic-importance` → `/api/v1/logical-capabilities/{id}/strategic-importance`
   - `/api/v1/enterprise-capabilities/{id}/target-maturity` → `/api/v1/logical-capabilities/{id}/target-maturity`
   - `/api/v1/enterprise-capabilities/{id}/maturity-gap` → `/api/v1/logical-capabilities/{id}/maturity-gap`
   - `/api/v1/enterprise-capabilities/maturity-analysis` → `/api/v1/logical-capabilities/maturity-analysis`
   - `/api/v1/domain-capabilities/{id}/enterprise-capability` → `/api/v1/domain-capabilities/{id}/logical-capability`
   - `/api/v1/domain-capabilities/{id}/enterprise-link-status` → `/api/v1/domain-capabilities/{id}/logical-mapping-status`
   - Old routes are removed in the same release. The frontend is the only known consumer; it renames in lockstep.
8. **Frontend renames in lockstep with backend.** User-visible strings, query keys, hook names, component names, mock fixtures, branded ID types, and the API module rename together. Where the user-facing language used "link" / "linked" in the mapping UI, it becomes "map" / "mapped" to match the renamed aggregate.
9. **The package name `enterprisearchitecture` is NOT renamed in this slice.** The Go package and the frontend feature folder both represent the bounded context, not the aggregate; renaming them is high-churn structural movement with zero semantic gain. Inside the package, every aggregate / event / handler / route / DB type renames. The DDD memo accepts this trade-off explicitly.
10. **HATEOAS link rels that reference the old type name update consistently.** The `x-related` array entries on canvas-renderable types whose `targetType` or `relationType` referenced "enterprise-capability" rename to reference "logical-capability" — verified during Swagger regeneration.

---

## Acceptance Criteria

- [ ] Every Go identifier containing `EnterpriseCapability*` or `EnterpriseStrategicImportance*` is renamed per rules 2/3/4; `grep -r EnterpriseCapability backend/` and `grep -r EnterpriseStrategicImportance backend/` both return zero matches
- [ ] Every event constant in `backend/internal/enterprisearchitecture/publishedlanguage/events.go` is renamed per rule 4
- [ ] The event-deserialization layer translates all six `EnterpriseCapability*` event types and the three `EnterpriseStrategicImportance*` types from on-disk wire format to the renamed Go structs (via dual deserializer registrations and explicit upcasters where field-name renames apply); an integration test loads a fixture event store containing only old-named events, runs projections, and asserts the produced read model is identical to one produced by writing equivalent renamed events from scratch
- [ ] DB migration renames the three tables and every dependent index / foreign key / column per rule 6 (forward-only — no down migration, per EASI conventions)
- [ ] HTTP routes in `backend/internal/enterprisearchitecture/infrastructure/api/routes.go` reflect rule 7 verbatim
- [ ] Calling any old route under `/api/v1/enterprise-capabilities` returns 404 Not Found (no redirect)
- [ ] Swagger documentation regenerates with the renamed types, routes, and DTO fields; `grep -r EnterpriseCapability backend/docs/` returns zero matches
- [ ] Frontend branded ID types in `frontend/src/api/types.ts` rename per rule 3
- [ ] Frontend feature folder content renames consistently inside `frontend/src/features/enterprise-architecture/` (folder name itself unchanged per rule 9): types, hooks, components, query keys, mock fixtures
- [ ] Every user-visible string referencing "Enterprise Capability" / "Enterprise Capabilities" / "linked" (in the mapping context) is replaced per rule 8; `grep -ri "enterprise capabilit" frontend/src/` returns zero matches
- [ ] All existing backend tests pass with the renamed types; tests reference the new names; the integration test in rule-3 above is added
- [ ] All existing frontend tests pass with the renamed types; tests reference the new names
- [ ] All existing UX flows (create / read / update / delete logical capability; map / unmap a physical capability; set target maturity; set / update / remove strategic importance; fetch maturity gap) work end-to-end without regression
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file
- [ ] A manual screenshot walkthrough confirms the strings "Enterprise Capability" / "Enterprise Capabilities" / "Enterprise Capability Link" appear nowhere in the UI

---

## Architecture

### Ownership

The `enterprisearchitecture` bounded context owns this rename. The Go package itself is not renamed (rule 9). Every aggregate / value object / event / handler / repository / read model / projector / API route / DTO / database table inside the context renames.

The `capabilitymapping` context is structurally unaffected: it holds no compile-time references to the renamed types. Its existing event subscriptions translate via the published-language upcaster.

### Domain Model

Aggregates renamed per rule 2; value objects renamed per rule 3; events renamed per rule 4. Existing invariants are preserved exactly. No new aggregate is introduced. No new event is introduced. No new field on any aggregate.

### API Surface

Routes rename per rule 7. The `/links` URL segment becomes `/mappings` to match the renamed aggregate. The `/target-maturity`, `/strategic-importance`, and `/maturity-gap` segments are unchanged because they refer to entities scoped to the parent capability, not to the parent capability's name.

DTO field names that referenced the old name update (e.g. `enterpriseCapabilityId` → `logicalCapabilityId`); shape and semantics otherwise unchanged.

### Persistence

Tables renamed in place via SQL migration per rule 6. Existing data preserved. Indexes, foreign keys, and triggers update to match.

The event store is not rewritten. Events on disk keep their original type names. Upcasters in the published-language deserialization layer translate at read time. This is the standard EASI pattern for renames (see `easi-go-backend-patterns`, `easi-database-migrations` skills during implementation).

### Frontend

Affected feature: `frontend/src/features/enterprise-architecture/`. Folder name unchanged (rule 9); contents rename. The exported API module becomes `logicalCapabilityApi`; the file may be renamed opportunistically if it does not multiply diff size.

Branded ID types in `frontend/src/api/types.ts` rename per rule 3.

Query key namespace migrates: `enterpriseCapabilitiesQueryKeys` → `logicalCapabilitiesQueryKeys`. Mutation cache invalidation in `mutationEffects.ts` updates to match.

User-facing strings update per rule 8.

### Cross-Context Integration

The `capabilitymapping` context's projectors that subscribe to renamed events (per the explore findings, `domain_capability_metadata_projector` is one) update their subscriptions to consume the renamed event types. Because events on disk keep old names, the upcaster bridges the wire format until the event store is naturally written-over by future events.

No published-language event escapes the renamed naming once upcasters are in place: every consumer downstream sees the renamed types at the deserialization boundary.

---

## Design Decisions

1. **API routes rename without 301 redirects or duplicate handlers.** Rationale: the EASI frontend is the only known API consumer; it renames in the same release. Deprecation overhead doesn't pay back. Alternative considered: 301-redirect old routes for one release — rejected as code to delete later. If a previously-unknown consumer surfaces post-deploy, a follow-up spec adds redirects on a per-route basis.

2. **`EnterpriseCapabilityLink` → `LogicalCapabilityMapping` (and `/links` → `/mappings`).** Per the DDD memo. The relationship between a logical capability and the physical capabilities under it is a *mapping*, not a generic *link*. Naming alignment with the domain concept reads more clearly in code and in the URL. Alternative considered: keep "link" as the user-facing word while renaming the type — rejected because frontend / backend vocabulary divergence is the exact problem this slice exists to fix.

3. **`EnterpriseStrategicImportance` → `StrategicImportance` (no qualifier).** The type is scoped by its foreign key (`logical_capability_id`); repeating the parent name in the type name is noise. Alternative considered: `LogicalCapabilityStrategicImportance` — rejected as verbose with no informational gain. If a second strategic-importance concept is introduced in another context later, it gets its own qualifier at that point, not pre-emptively.

4. **Event store events keep old type names on disk; upcasters translate at read.** Standard EASI pattern (per `easi-go-backend-patterns` and `easi-database-migrations`). Avoids a destructive event-store rewrite and keeps historical events as the immutable source of truth.

5. **Package and folder names retained.** The Go `enterprisearchitecture` package and the frontend `enterprise-architecture` folder represent the bounded context, not the aggregate. Renaming them ripples into ~30+ import paths for purely cosmetic gain. Alternative considered: rename to `logicalarchitecture` or similar — rejected as out of proportion to value. The DDD memo accepts this trade-off and notes the bounded context carries the historical name.

6. **Frontend "link" / "linked" UI vocabulary becomes "map" / "mapped" in the mapping context.** Internal consistency: when the aggregate is `LogicalCapabilityMapping`, every surface that touches it should use the same word. Alternative considered: keep colloquial "link" in the UI — rejected as the exact split-vocabulary problem this slice fixes.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Hard rename of API routes (no redirects) | Breaks any unknown external consumer | Search internal repos before deploy; if a consumer appears post-deploy, add a redirect via follow-up spec |
| Event store keeps old type names; upcasters translate at read | Upcaster code is permanent | Standard EASI pattern; the cost is one-time and bounded |
| Bounded-context folder names retained | Folder name diverges from displayed vocabulary | Acceptable: developers reading code understand the context's history; users see only the labels, which fully rename |
| `StrategicImportance` loses the `Enterprise` qualifier | Slightly less self-documenting in isolated grep results | Naming follows DDD principle: the type is scoped by its foreign key, not by repeating context names |
| Renaming "link" → "map" in the UI | Architects in active use today are accustomed to "link" / "linked" | A single release-note line covers the change; the new word fits the renamed aggregate consistently |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
