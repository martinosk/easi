# 226 — Domain Stewardship

> **Status:** ongoing
> **Depends on:** 214 (typed user references on an artifact), 126 (Access Delegation shape and deletion cascade)
> **Roadmap alignment:** `SD8 (clarified 2026-09-29) / H2-6` — slice A of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) (decisions D2, D3, D4)

---

## Problem Statement

Nobody can be accountable for anything in a business domain today. The domain carries a domain architect, who curates the model, and nothing else — and that field, recorded since spec 117, is read by nothing: no page shows it, no check or routing uses it; an application has an owner, a capability an EA owner, but "who answers for this domain's applications all having owners" or "who answers for its TIME assessments being current" cannot be recorded. Without that fact, a home page cannot know whose attention a domain's problems deserve, and a stakeholder accountable for part of a domain's landscape has no standing in the tool at all.

This slice records the fact. A **stewardship** is one user accountable for one **concern** in one domain. Concerns are the categories of things a landscape can be wrong about, and the list is fixed in code because each concern exists only through the checks that will produce its attention items (design doc D4). This slice ships the write side and its surface on the domain board; the attention items that use it are slices C and D.

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Concern** | One kind of thing a domain's landscape can be wrong about: `ownership`, `assessment`, `documentation`, `planning`, `structure` (defined in the design doc). Fixed in code; not tenant vocabulary. |
| **Stewardship** | The fact that one user is accountable for one concern in one domain. At most one per (domain, concern). |
| **Steward** | The user holding a stewardship. Distinct from the domain architect (curates the model) and from an application owner (fixes their own application). Not the "ownership steward" of spec 214's route catalog, which names whoever curates an application's ownership record. |
| **Assign** | The command that creates a stewardship, or changes the steward of an existing one. |
| **Release** | The command that ends a stewardship. |
| **Unassigned** | The read-side label for a concern with no stewardship in a domain. |
| **Fallback** | The domain architect, shown beside an unassigned concern as steward of last resort. Derived on read, never recorded. |
| **Candidate** | Any active user of the tenant, whatever their role. |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Admin / Architect** (roles `admin`, `architect`) | Record who answers for each concern in each domain, change it when people move, see at a glance which concerns nobody answers for. |
| **Stakeholder** (role `stakeholder`) | Be assignable as a steward without any write role, and see who the stewards of a domain are. |
| **Domain Architect** (the user set on the `BusinessDomain`; always an admin or architect today per spec 117, but the fallback is by that field, not by role) | See which concerns in their domain have a steward and which fall back to them. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Domain stewardship

  Scenario: Every concern is listed, assigned or not
    Given the domain "Customer Engagement" has a steward for "ownership" and none for the other concerns
    When a user who may read domains opens Stewards for "Customer Engagement"
    Then every concern is listed with its label and description
    And "ownership" shows the steward's name
    And every other concern reads "Unassigned"
    And, "Customer Engagement" having the domain architect "Alice Smith", every unassigned concern names her as the fallback

  Scenario: Unassigned concern in a domain without a domain architect
    Given the domain "Logistics" has no domain architect and no stewards
    When a user who may read domains opens Stewards for "Logistics"
    Then every concern reads "Unassigned" with no fallback named

  Scenario: Fallback whose name is unknown
    Given the domain "Logistics" has a domain architect whose user is not in the user cache
    When a user who may read domains opens Stewards for "Logistics"
    Then every unassigned concern names the fallback as "Unknown user"

  Scenario: Assigning a steward
    Given an architect opens Stewards for "Customer Engagement"
    When they assign the stakeholder "Mette Gram" to "assessment"
    Then "assessment" shows "Mette Gram"
    And the assignment is recorded with who assigned it and when

  Scenario: Any active user is a candidate
    Given an architect assigns a steward
    Then the candidates are every active user of the tenant, whatever their role

  Scenario: Assigning a disabled user is rejected
    Given "Jonas Holm" is a disabled user of the tenant
    When an architect assigns "Jonas Holm" to "assessment" in "Customer Engagement"
    Then the request is rejected and nothing is recorded

  Scenario: Reassigning replaces the steward
    Given "assessment" in "Customer Engagement" is stewarded by "Mette Gram"
    When an architect assigns "Jonas Holm" to "assessment"
    Then "assessment" shows "Jonas Holm"
    And "Mette Gram" no longer stewards "assessment" in that domain

  Scenario: Assigning the current steward again changes nothing
    Given "assessment" in "Customer Engagement" was assigned to "Mette Gram" by "Alice Smith"
    When another architect assigns "Mette Gram" to "assessment"
    Then the request succeeds
    And "assessment" still shows "Mette Gram", assigned by "Alice Smith" at the original time
    And nothing is recorded

  Scenario: One user may steward many concerns
    Given "Mette Gram" stewards "assessment" in "Customer Engagement"
    When an architect assigns "Mette Gram" to "documentation" in the same domain
    Then both concerns show "Mette Gram"

  Scenario: Releasing a steward
    Given "assessment" in "Customer Engagement" is stewarded by "Mette Gram"
    When an architect releases the steward of "assessment"
    Then "assessment" reads "Unassigned"

  Scenario: Releasing an unassigned concern changes nothing
    Given "planning" in "Customer Engagement" is unassigned
    When an architect releases the steward of "planning"
    Then the request succeeds
    And nothing is recorded

  Scenario: Readers see, writers change
    Given a stakeholder opens Stewards for "Customer Engagement"
    Then the stewards are listed
    And no assign or release control is shown
    And the candidate list is not requested
    And an assignment request from that stakeholder is rejected as forbidden

  Scenario: Unknown user
    When an architect assigns a user id that does not exist in the tenant to a concern
    Then the request is rejected and nothing is recorded

  Scenario: Unknown concern
    When a request names a concern that is not in the fixed list
    Then the request is rejected as invalid

  Scenario: Unknown domain
    When a request reads, assigns or releases stewardships of a domain id that does not exist in the tenant
    Then the request is rejected as not found

  Scenario: Domain deletion releases its stewardships
    Given "Customer Engagement" has two stewards
    When the domain is deleted
    Then both stewardships are released
    And Stewards for that domain is no longer available

  Scenario: A disabled steward stays recorded
    Given "Mette Gram" stewards "assessment" in "Customer Engagement"
    When the admin disables "Mette Gram"
    Then "assessment" still shows "Mette Gram"

  Scenario: Reaching Stewards from the board
    Given a user who may read domains is on the Business Domains board
    When they open a domain's menu
    Then it offers "Stewards…"
```

---

## Business Rules & Invariants

1. **Concerns are fixed** — exactly `ownership`, `assessment`, `documentation`, `planning`, `structure`, each with a label and a one-line description served by the API. Any other value is invalid. Labels and descriptions are code, not tenant-configurable vocabulary (invariant 5 is about the model's vocabulary, design doc D4).
2. **One steward per (domain, concern)** — at most one live stewardship exists for a pair. Assigning to an already stewarded pair replaces the steward in one step.
3. **Steward is an active user reference** — an existing, active user of the tenant, any role, referenced by id; a disabled user cannot be assigned; never free text, never a team.
4. **Assignment requires `domains:write`; reading requires `domains:read`.** An edit grant on the domain does not confer assignment.
5. **Every assignment is attributed** — who assigned, when.
6. **Domain deletion releases** — every stewardship of a deleted domain is released by the system, attributed to the deletion.
7. **User disablement does not release** — a steward disabled after assignment stays recorded; surfacing it is an attention item (slice D).
8. **Stewardship is tenant-scoped** — rows carry the tenant and are isolated by row-level security.
9. **Affordance-gated UI** — the frontend shows assign and release controls only when the response carries the corresponding link, and requests the candidate list only when an assign link is present; it never inspects the role.
10. **Events state the current fact** — `StewardAssigned` carries domain id, concern, steward user id, actor and time, and means "this user is now the steward of this pair"; `StewardReleased` carries domain id, concern, actor and time. Neither carries the previous steward: the pair's stream is ordered, and a consumer that needs the transition reads the preceding event or its own state. Publishing them is an additive change to the context's contract; no consumer exists in this slice.
11. **The domain architect is the steward of last resort** — an unassigned concern falls back to the domain's architect on the read side (design doc). The fallback is derived, never recorded as a stewardship, and is absent when the domain has no architect.
12. **Commands are idempotent** — assigning the user who already stewards the pair, or releasing an unassigned concern, succeeds, changes nothing and records no event; attribution stays as it was.
13. **Commands and reads address an existing domain** — a domain id unknown to the tenant, or deleted, is not found.

---

## Acceptance Criteria

- [x] `GET` of a domain's stewardships lists all five concerns with label, description, steward (id and name, or null), assignedBy and assignedAt, in a fixed concern order, plus the domain's fallback (domain architect id and name, or null; id with null name when the user is not in the cache, rendered "Unknown user")
- [x] Assign and release commands behave per rules 1–7, 12 and 13, with at least one test per rule at the layer that enforces it (aggregate, handler, middleware or reactor)
- [x] Assign with an unknown user id, a disabled user id or an unknown concern is rejected with the status codes `easi-api-standards` prescribes; nothing is stored
- [x] Reads and commands for an unknown or deleted domain answer 404
- [x] Assigning the current steward again, or releasing an unassigned concern, succeeds without appending an event
- [x] A caller without `domains:write` receives no assign or release links and gets 403 on the commands
- [x] Deleting a domain releases its stewardships (integration test over the published `BusinessDomainDeleted` event) and the collection answers 404 for the deleted domain
- [x] Steward names resolve for every user created before or after deployment (backfilled user cache); a user's active status follows `UserDisabled` and `UserEnabled`
- [x] The domain menu on the board offers "Stewards…" when the domain carries the `x-stewardships` link; the dialog lists concerns, shows assign and release controls only when their links are present, and fetches candidates from the tenant's active users only when an assign link is present
- [x] Mutations invalidate the stewardship query through the feature's mutation effects
- [x] Every BDD scenario has at least one corresponding test
- [x] The architecture guard tests pass with the new context, and the dependency graph shows exactly two edges from `stewardship`: to `capabilitymapping` and to `auth`
- [x] `docs/architecture/components.csv` is regenerated, `docs/backend/cross-context-events.md` lists the context's published events and subscriptions, and the canvas matches the implementation
- [x] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

A new bounded context, **`stewardship`** (`/backend/internal/stewardship/`), classified Supporting. It owns the `Stewardship` aggregate and the concern vocabulary. It depends on Capability Mapping (domain lifecycle events) and Auth (user events and permission constants) through their published languages only. Capability Mapping is unchanged apart from the `x-stewardships` link the shared HATEOAS helper adds to domain responses, following the `x-edit-grants` precedent. The context's canvas is [`docs/architecture/Stewardship.md`](../docs/architecture/Stewardship.md).

**Boundary guards.** The architecture tests in `/backend/internal/` discover contexts by scanning directories, so `stewardship/` is covered the moment it exists, with no allowlist entry:

| Guard | What it enforces for this context |
|-------|-----------------------------------|
| `TestNoCrossBoundedContextImports` | Production code imports nothing of another context except its `publishedlanguage`; this context imports only `capabilitymapping/publishedlanguage` and `auth/publishedlanguage`, and no context imports `stewardship/` |
| `TestContextDependencyGraphIsAcyclic` | Any published-language import from Capability Mapping or Auth back into this context closes a cycle and fails |
| `TestSharedAndInfrastructureImportNoContext` | The `x-stewardships` link in `shared/api` is a URL string, never an import of this context |
| `TestCompositionRootOnlyRegistersRoutes` | The composition root imports only `stewardship/infrastructure/api`; subscriptions and projectors are wired inside `SetupRoutes` |
| `TestSQLSchemaOwnership`, `TestNoSQLOutsideApprovedLocations` | Runtime SQL names only the `stewardship` schema and lives in `application/readmodels`, `application/projectors` or `infrastructure/repositories` |
| `TestNewMigrationsCrossSchemasOnlyInBackfills` | The domain and user cache seeds read `capabilitymapping` and `auth` tables only in migrations whose name contains `backfill` |

Routes are not registered in the Arch Assistant tool catalog, as Access Delegation's are not. `docs/architecture/components.csv` is regenerated so CI does not fail on drift.

### Domain Model

- **`Stewardship` aggregate**, one per live (domain, concern) pair, with its own identity. State: domain id, concern, steward user id, assigned by, assigned at. Commands: `AssignSteward` (creates the aggregate for an unassigned pair; changes the steward on an existing one; no-op when the steward is unchanged), `ReleaseSteward` (ends the aggregate). Events: `StewardAssigned`, `StewardReleased`. The handler locates the pair's live stewardship through the read model; rule 2's uniqueness is enforced at the handler with a unique index on the live-stewardship read model as backstop (a release removes the row, so the index covers only live pairs) — the one-per-owner shape of spec 216's `ComponentContainments`.
- **`Concern` value object** — the fixed list with label and description.
- **`StewardRef` value object** — a user id validated against the local user cache, which must hold the user as active.
- Rules 1, 3, 5 and 12 live in the aggregate and value objects; 2 and 13 at the handler; 4 in middleware; 6 is a reactor on `BusinessDomainDeleted` that dispatches `ReleaseSteward` for every live stewardship of the domain, attributed to `system:domain-deleted` (the Access Delegation artifact-deletion precedent); 7 holds because no reactor subscribes to `UserDisabled`.

### API Surface

- `/stewardships`, a collection filtered by domain, readable with `domains:read`, returning every concern in fixed order with its steward and, when the caller holds `domains:write`, an assign affordance per concern and a release affordance per assigned concern.
- `/stewardships/{domainId}/{concern}`, readable with `domains:read` (one concern), accepting assign (`PUT`, body `stewardId`) and release (`DELETE`), both idempotent (rule 12) and both requiring `domains:write`. Item links: `self`, `collection`, and for `domains:write` `x-assign` plus `x-release` when assigned; collection links: `self`, `x-domain`, and for `domains:write` `x-candidates` (the tenant's active users). Unknown or disabled steward and unknown concern answer 400; unknown domain 404. `assignedBy` records the actor's e-mail.
- Domain responses from Capability Mapping carry `x-stewardships` for any caller who may read domains.
- Shapes, status codes, link rels and Swagger per `easi-api-standards`.

### Persistence

Schema `stewardship`: a stewardships read model, a domain cache (existence, name and domain architect id, fed by domain lifecycle events) and a user cache (id, name and active status, fed by `UserCreated`, `UserDisabled` and `UserEnabled`; the status backs rule 3). All tenant-scoped with row-level security; caches backfilled by `backfill` migrations per spec 209. Events append to the shared event store.

### Frontend

Feature `stewardship` under `/frontend/src/features/`: API client, query keys, `useDomainStewardships` and assign/release mutation hooks with `mutationEffects.ts`, and a `StewardsDialog` (Mantine) listing concerns with steward name or "Unassigned" (naming the domain architect as fallback when the domain has one), a user select fed by the existing users hook and enabled only when an assign link is present (stakeholders lack `users:read`), and controls gated on the row's links. The Business Domains board's domain menu gains "Stewards…" next to "Invite to Edit…", shown when the domain carries `x-stewardships`.

### Cross-Context Integration

| Direction | Event / command | Purpose |
|-----------|-----------------|---------|
| Capability Mapping → Stewardship | `BusinessDomainCreated`, `BusinessDomainUpdated`, `BusinessDomainDeleted` | Domain cache (existence, name, architect); deletion release |
| Auth → Stewardship | `UserCreated`, `UserDisabled`, `UserEnabled` | User existence, name and active status |
| Stewardship → (any) | `StewardAssigned`, `StewardReleased` | Published for the home read side (slice C); no consumer in this slice |

---

## Design Decisions

1. **Own context, not a field on `BusinessDomain`** — the concern vocabulary is defined by the checks that will use it, and Capability Mapping should not learn it (design doc D3). Alternative: `stewards` map on the domain aggregate (rejected — every future concern would be a Capability Mapping change).
2. **Aggregate per (domain, concern) with its own identity, not per domain** — assignments change independently and the event history reads per concern; one aggregate per domain would replay every concern's history to change one. The aggregate id is intrinsic and the pair is located through the read model, so a released pair leaves no aggregate to resurrect. Alternatives: one `DomainStewards` aggregate (rejected — coarser than the fact); identity derived from (domain id, concern) (rejected — an aggregate id built from another context's id, and a "released" flag to reuse it).
3. **Replace in one command, one event, no previous steward** — reassigning is `AssignSteward` on a stewarded pair, yielding a single `StewardAssigned` that states the new steward. Alternatives: require release then assign (rejected — a two-step change leaves a window with nobody accountable and doubles the UI work); carry the previous steward on the event (rejected — duplicates what the ordered stream already holds and makes one event mean two things depending on a field's presence); a separate `StewardReassigned` event (rejected — same information, one more event type for consumers to handle).
4. **Any active user is a candidate** — accountability is not a modelling role; the stakeholder story in the design doc depends on it. Alternative: reuse the admin-or-architect candidate list of the domain architect (rejected — excludes the very people the feature exists for).
5. **Existing `domains:*` permissions, no new permission** — assigning accountability is domain curation. Alternative: `stewardships:manage` (rejected — a fourth permission group with the same holders).
6. **Disablement does not release, but a disabled user cannot be assigned** — a silent release would hide a gap; the gap becomes an attention item in slice D. Assigning someone who cannot sign in is accountability for nobody, so the write side refuses it and the user cache tracks status for that check. Alternative: accept any existing user (rejected — the UI would offer active users while the API accepted disabled ones).
7. **Link on the domain via the shared HATEOAS helper** — the `x-edit-grants` precedent: a URL convention in the shared kernel, not a code dependency between contexts.
8. **Domain architect stays a Capability Mapping field, surfaced as fallback** — the design doc keeps the architect where it is and names them steward of last resort; this slice is the first consumer of the field. Alternative: fold the architect into the concern list as a "model" stewardship and migrate the field (rejected — contradicts the design doc's placement; an amendment for a human, not a spec).
9. **A steward is one person, never a team** — a concern needs one accountable person (design doc D2). This differs deliberately from SD6, where an application owner may be a team: an application is fixed by whoever owns it, a concern is answered for by someone.
10. **Idempotent commands** — assigning the current steward again or releasing an unassigned concern is a successful no-op with no event. Alternative: emit anyway to refresh attribution (rejected — an event records a change, and nothing changed).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| New context for one aggregate | Wiring, schema and guard-test cost before any read side exists | The context is the home of slices C–E; the cost is paid once |
| Fixed concern list in code | Adding a concern is a deploy | A concern without checks is meaningless; each arrives with its spec |
| User existence and status checked against a local cache | Cache lag between `UserCreated` / `UserDisabled` and assignment | Same-process event bus; backfill covers pre-existing users |
| No previous steward on `StewardAssigned` | A stateless consumer wanting "who was replaced" must read the stream | No such consumer exists; slices C–D are projectors holding the prior row |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
