# 216 — Application Composition

> **Status:** pending
> **Depends on:** —
> **Roadmap alignment:** SD6 / H1-2

---

## Problem Statement

Application components are flat. Real landscapes contain suites and platforms whose parts deserve their own record (own experts, own realisations, own one-pager) while still reading as one product. There is no way to express "CRM Suite contains the Quoting module". ComponentRelation covers behavioural coupling (Triggers, Serves) but not structural containment.

The roadmap deliberately caps composition at two levels: a parent and its children, never deeper.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Model suites/platforms and their parts without unbounded hierarchies |
| **Stakeholder** | Read a component page and understand what it is part of, or contains |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Application composition, maximum two levels

  Scenario: Composing a component into a parent
    Given standalone components "CRM Suite" and "Quoting"
    When an architect composes "Quoting" into "CRM Suite"
    Then "Quoting" shows it is part of "CRM Suite"
    And "CRM Suite" lists "Quoting" among its parts

  Scenario: A child cannot accept parts
    Given "Quoting" is part of "CRM Suite"
    Then "Quoting" offers no compose affordance for accepting parts

  Scenario: A parent cannot become a part
    Given "CRM Suite" has parts
    Then "CRM Suite" offers no affordance to become part of another component

  Scenario: Extracting a part
    Given "Quoting" is part of "CRM Suite"
    When an architect extracts "Quoting"
    Then "Quoting" is standalone again

  Scenario: Deleting a parent releases its parts
    Given "CRM Suite" contains "Quoting" and "Billing"
    When an architect deletes "CRM Suite"
    Then "Quoting" and "Billing" still exist as standalone components
```

---

## Business Rules & Invariants

1. **Single parent** — a component is part of at most one parent.
2. **Two levels maximum** — a component that has parts cannot become a part; a component that is a part cannot accept parts. No component may be composed into itself.
3. **Composition is not a relation** — Triggers/Serves relations are unchanged and may exist independently of composition.
4. **Deletion releases, never cascades** — deleting a parent extracts its parts; parts are never deleted, hidden, or orphaned by a parent's deletion.
5. **HATEOAS-driven legality** — compose and extract affordances appear exactly when rules 1–2 permit the operation for the actor; the UI derives legality only from links.

---

## Acceptance Criteria

- [ ] Compose and extract operate per rules 1–2, each raising its own domain event
- [ ] Component detail responses carry the parent reference and the list of parts
- [ ] Deleting a parent releases all parts per rule 4
- [ ] Affordances appear exactly per rule 5, including their absence on children and populated parents
- [ ] Relations, realisations, experts, and one-pagers of a part are unaffected by composition changes

---

## Architecture

### Ownership

ArchitectureModeling only.

### Domain Model

A parent reference on the ApplicationComponent aggregate. Events: `ApplicationComponentComposedInto`, `ApplicationComponentExtracted`. The depth and single-parent invariants span two aggregates, so they are enforced at the command handler via the component read model, with the aggregate guarding its own local facts (a component with a parent refuses parts; a component with parts refuses a parent, based on state it learns through its own events). Parent deletion triggers an AM-internal reaction extracting each part.

### API Surface

Compose/extract operations on the component (`components:write`); parent and parts in component representations; `x-compose-into` / `x-extract` affordances.

### Persistence

Parent column and part counts on the component read model; no migration backfill values needed (all existing components are standalone).

### Frontend

Components feature: "Part of" / "Contains" section on the details panel with compose and extract actions; the components list shows containment as an annotation. Views and canvases render components exactly as today.

### Cross-Context Integration

None. Composition events are published for future read sides; no consumer in this slice.

---

## Design Decisions

1. **Parent reference on the aggregate, not a composition aggregate** — containment has no identity or lifecycle of its own; a link aggregate would add ceremony without behaviour. Alternative: `ComponentComposition` aggregate (rejected).
2. **Depth invariant at the command handler with read-model lookup** — the two-level rule inspects two aggregates and follows the cross-aggregate-invariant convention; a DB check constraint is a backstop only.
3. **Release-on-delete** — follows the standing rule that removal never cascades to other entities.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Handler-level depth check | Concurrent compose commands could race past the check | DB backstop constraint; conflict surfaces on the second write |
| Two-level cap | Deep product hierarchies cannot be modelled | Deliberate roadmap decision (SD6); revisit only via amendment |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
