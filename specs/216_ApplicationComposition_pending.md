# 216 — Application Composition

> **Status:** pending
> **Depends on:** —
> **Roadmap alignment:** SD6 / H1-2

---

## Problem Statement

Application components are flat. Real landscapes contain suites and platforms whose parts deserve their own record (own experts, own realisations, own one-pager) while still reading as one product. There is no way to express "CRM Suite contains the Quoting module". ComponentRelation covers behavioural coupling (Triggers, Serves) but not structural containment.

Containment comes in two strengths, and the difference shows up exactly when the parent goes away. A monolith is **composed of** its modules: the modules have no existence outside the monolith, so retiring the monolith retires them. A platform **aggregates** autonomous services: the services stand on their own, so retiring the platform releases them. A single containment semantics would silently misrepresent one of these two landscapes; the kind must be declared per part.

The roadmap deliberately caps containment at two levels: a parent and its parts, never deeper.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Model suites, platforms and monoliths and their parts, declaring whether a part can outlive its parent, without unbounded hierarchies |
| **Stakeholder** | Read a component page and understand what it is part of, or contains, and what happens to the parts if the parent is retired |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Application containment, maximum two levels, composition or aggregation

  Scenario: Attaching a part by composition
    Given standalone components "Monolith" and "Order Module"
    When an architect attaches "Order Module" to "Monolith" as a composition
    Then "Order Module" shows it is part of "Monolith" by composition
    And "Monolith" lists "Order Module" among its parts

  Scenario: Attaching a part by aggregation
    Given standalone components "Platform" and "Search Service"
    When an architect attaches "Search Service" to "Platform" as an aggregation
    Then "Search Service" shows it is part of "Platform" by aggregation
    And "Platform" lists "Search Service" among its parts

  Scenario: A part has exactly one parent
    Given "Quoting" is part of "CRM Suite"
    Then "Quoting" offers no affordance to attach to another parent
    And a request to attach "Quoting" to "ERP Suite" is rejected

  Scenario: A part cannot accept parts
    Given "Quoting" is part of "CRM Suite"
    Then "Quoting" offers no affordance for accepting parts
    And a request to attach "Billing" to "Quoting" is rejected

  Scenario: A parent cannot become a part
    Given "CRM Suite" has parts
    Then "CRM Suite" offers no affordance to attach to another component
    And a request to attach "CRM Suite" to "ERP Suite" is rejected

  Scenario: Detaching a part
    Given "Quoting" is part of "CRM Suite"
    When an architect detaches "Quoting"
    Then "Quoting" is standalone again
    And "CRM Suite" no longer lists "Quoting"

  Scenario: Detaching a composed part is allowed
    Given "Order Module" is part of "Monolith" by composition
    When an architect detaches "Order Module"
    Then "Order Module" is standalone again

  Scenario: Deleting a parent deletes composed parts and releases aggregated parts
    Given "CRM Suite" contains "Quoting" by composition and "Billing" by aggregation
    When an architect deletes "CRM Suite"
    Then "Quoting" is deleted
    And "Billing" still exists as a standalone component

  Scenario: Delete confirmation names the composed parts
    Given "CRM Suite" contains "Quoting" by composition and "Billing" by aggregation
    When an architect opens the delete confirmation for "CRM Suite"
    Then the confirmation states that "Quoting" will be deleted with it
    And the confirmation states that "Billing" will be released

  Scenario: Deleting a part leaves the parent intact
    Given "Quoting" is part of "CRM Suite"
    When an architect deletes "Quoting"
    Then "CRM Suite" still exists and no longer lists "Quoting"
```

---

## Business Rules & Invariants

1. **Single parent** — a component is part of at most one parent. Moving a part is detach then attach.
2. **Two levels maximum** — a component that has parts cannot become a part; a component that is a part cannot accept parts. No component may be attached to itself.
3. **Declared kind** — every containment is exactly one of `composition` (the part exists only within its parent) or `aggregation` (the part is autonomous). The kind is chosen at attach time and is fixed for the life of the containment; changing it is detach then attach.
4. **Containment is not a relation** — Triggers/Serves relations are unchanged and may exist independently of containment.
5. **Parent deletion follows the kind** — deleting a parent deletes its composed parts and releases its aggregated parts. A released part is standalone; it is never hidden or left referencing a deleted parent.
6. **Parts keep their own record** — a part's relations, realisations, experts, ownership, hosting and one-pager are untouched by attach, detach and release. A composed part deleted with its parent is deleted through the ordinary component deletion, so its own cascades run.
7. **Detach and delete are always available on a part** — composition restricts what happens to a part when its parent is deleted, not what an architect may do to the part directly.
8. **HATEOAS-driven legality** — attach and detach affordances appear exactly when rules 1–2 permit the operation for the actor; the UI derives legality only from links.

---

## Acceptance Criteria

- [ ] Attach and detach operate per rules 1–3, each raising its own domain event carrying the part and parent references; attach also carries the kind
- [ ] Rules 1–2 are enforced inside one aggregate, so two conflicting attach commands can never both succeed
- [ ] Component detail responses carry the parent reference with its kind, and the list of parts with their kinds
- [ ] Deleting a parent deletes composed parts and releases aggregated parts per rule 5, parts first, then the parent
- [ ] The delete confirmation for a parent names the composed parts that will be deleted and the aggregated parts that will be released
- [ ] Affordances appear exactly per rule 8, including their absence on parts and on populated parents
- [ ] Relations, realisations, experts, ownership, hosting and one-pagers of a part are unaffected by attach, detach and release
- [ ] Existing components read as standalone after deployment

---

## Architecture

### Ownership

ArchitectureModeling only.

### Domain Model

A new `ComponentContainments` aggregate, one per tenant, holding every containment in the landscape as entries of (part reference, parent reference, `ContainmentKind`). The ApplicationComponent aggregate is untouched: it neither knows its parent nor its parts.

Because every containment in the tenant lives in one aggregate, rules 1–3 are intrinsic invariants enforced in its command methods from its own state: `Attach` refuses a part that already has a parent, a part that has parts, a parent that is itself a part, a part equal to its parent, and an unknown kind; `Detach` refuses a part that is standalone. Concurrent attach and detach commands in a tenant serialise on this one stream through the event store's version check, so no cross-aggregate check and no read-model lookup is involved in the invariants.

Events on the containment stream: `ComponentAttached` (part, parent, kind) and `ComponentDetached` (part, parent).

The command handler only checks references: both components must exist and not be deleted, per the reference-validity convention. The aggregate is provisioned lazily on the tenant's first attach and located through its read model, with one-per-tenant uniqueness enforced at the handler and a unique index as backstop, the established shape for one-per-owner aggregates.

The component delete handler, before deleting the component, detaches it if it is a part, and for each of its parts dispatches the ordinary delete command (composition) or the detach command (aggregation), resolved from the read model. This mirrors the existing relation cascade in the same handler. Parts go first so that a failure midway leaves the parent intact and the deletion retryable. A composed part's own deletion runs the same handler, so it detaches itself and cascades its relations like any component.

### API Surface

Attach and detach operations addressed at the component (`components:write`); the attach request names the parent and the kind. Parent reference with kind, and parts with kinds, in component representations. Affordances `x-attach-to` (standalone components only) and `x-detach` (parts only), derived from the component read model. No new permission.

### Persistence

Nullable parent reference and kind columns on the component read model, projected from the containment events and indexed on the parent reference; parts and part counts are derived by query on the parent reference. A small registry table maps the tenant to its containment aggregate. No backfill values are needed since all existing components are standalone. No database constraint enforces a containment rule.

### Frontend

Components feature: a "Part of" / "Contains" section on the details panel with attach (parent and kind picker) and detach actions; the parts list shows each part's kind. The components list shows containment as an annotation. The attach picker offers only components that are not themselves parts, and the server rejects any ineligible target regardless. The delete confirmation for a parent lists what will be deleted and what will be released. Views and canvases render components exactly as today.

### Cross-Context Integration

None. Containment events are published for future read sides; no consumer in this slice. Composed parts deleted with their parent publish the ordinary `ApplicationComponentDeleted`, so existing cross-context deletion handlers need no change.

---

## Design Decisions

1. **One containment aggregate per tenant, not a parent reference on the component and not one aggregate per parent** — the two-level rule compares two facts, "P has parts" and "P is a part". Any boundary that records those on different streams leaves them a cross-aggregate invariant with a race between commands touching different streams. A parent reference on the component records them on the part's and the parent's streams; a per-parent aggregate records "P has parts" on P's containment and "P is a part" on Q's containment. Only a boundary that holds every containment in the tenant makes the rule intrinsic. The transactional cost is right: components are created and edited concurrently by many people, so they stay their own aggregate; containment changes are rare structural edits that can serialise per tenant. Precedents: `RealizationRoles` (spec 181) chose its boundary to make the single-standard invariant intrinsic; `MetaModelConfiguration` is one aggregate per tenant. Alternatives: parent reference on ApplicationComponent with handler-level depth check (rejected: race accepted by design); one aggregate per parent (rejected: closes nothing, makes even single-parent cross-aggregate).
2. **Composition and aggregation as a declared kind, not two features** — the two share every rule except what parent deletion does to the part, so one containment concept with a kind value object is the smallest model that keeps the deletion semantics honest. Alternatives: a single "composition" semantics with release on delete (rejected: misrepresents monoliths and hides the fact that the modules are gone); a single semantics with cascade on delete (rejected: destroys autonomous services with their platform); two separate concepts with separate events and affordances (rejected: doubles the surface for one differing branch).
3. **Kind fixed at attach time** — a reclassify operation would be a third command, event and affordance for a change no one has asked for; detach then attach expresses it today. Alternative: `ContainmentReclassified` event (rejected for now as unpulled).
4. **Detach and delete stay available on composed parts** — strict UML composition forbids a part leaving its whole, but EASI records landscape change, and extracting a module into a service is an ordinary architecture move. Alternative: block detach on composed parts (rejected).
5. **No database constraint** — the invariants are intrinsic to the aggregate, which is where the DDD convention puts them. EASI's backstop pattern is a unique index for one-per-owner uniqueness, and it is used here only for the one-containment-aggregate-per-tenant registry. A CHECK constraint cannot express a cross-row rule, and a trigger would be a new pattern moving domain rules into the database, the direction spec 128 explicitly reversed. Alternatives: CHECK constraint (rejected as inexpressible); trigger (rejected as a new pattern).
6. **Composed parts are deleted through the ordinary delete command** — each part then runs its own relation cascade and publishes its own deletion event, so views, realisations, grants and one-pagers clean up through the handlers that already exist. Alternative: a bulk "parent deleted" event consumed by every context (rejected: a new contract for every consumer).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| One containment aggregate per tenant | Every attach and detach in a tenant serialises on one stream; two architects attaching in the same instant get one concurrency conflict | Structural edits are rare and short; the conflict is a clean retryable error, and the aggregate replays a few hundred small events at most |
| Containment lives outside the component aggregate | The component representation joins parent and parts from a second source | Both are projected onto the component read model, so the API and HATEOAS derive from one row as today |
| Two-level cap | Deep product hierarchies cannot be modelled | Deliberate roadmap decision (SD6); revisit only via amendment |
| Kind fixed at attach time | Turning a monolith modular reads in history as detach then attach per part | One-off; a reclassify event is a small follow-up if the history fidelity matters to someone |
| Composition deletes parts with the parent | A mis-declared composition makes one delete remove several records | The delete confirmation names every part that will go; parts are deleted through the ordinary command, so nothing bypasses existing deletion behaviour |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
