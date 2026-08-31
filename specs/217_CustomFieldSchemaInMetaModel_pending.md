# 217 — Custom-Field Schema in MetaModel

> **Status:** pending
> **Depends on:** —
> **Roadmap alignment:** SD5 / H1-3

---

## Problem Statement

The tenant's extensible attribute schema — custom field names, types, selection options, number bounds — is owned by OnePagers, so MetaModel is a metamodel in name only and any context wanting a custom attribute must go through the fact-sheet context (coverage finding G1). The schema of the model belongs to the context whose language is "the vocabulary of the model"; how a one-pager displays and requires fields is presentation policy and stays with OnePagers.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Admin / metamodel steward** | Define the tenant's attribute vocabulary in one place |
| **Enterprise Architect** | Unchanged one-pager configuration and facts |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Attribute schema owned by MetaModel

  Scenario: Defining an attribute
    Given a steward with meta-model write permission
    When they define a "selection" attribute "Hosting Region" for subject type "application"
    Then the attribute is available in the application one-pager configuration

  Scenario: Retiring an attribute
    Given attribute "Hosting Region" is included in the application one-pager
    When the steward retires "Hosting Region"
    Then the one-pager stops offering the field for editing
    And previously recorded facts remain readable

  Scenario: Existing definitions survive the move
    Given a tenant with custom fields defined before this change
    When the change is deployed
    Then every field keeps its identity, type, options, and bounds
    And every recorded fact still resolves to its field

  Scenario: Requiredness stays a one-pager concern
    Given attribute "Hosting Region" exists in MetaModel
    When an architect marks it required on the application one-pager
    Then completeness counts it, and MetaModel is unchanged
```

---

## Business Rules & Invariants

1. **MetaModel owns schema** — per subject type: attribute name, data type (`text`, `number`, `date`, `link`, `selection`, `contact-person`), selection options, number bounds, active/retired. Managed only through MetaModel.
2. **OnePagers owns presentation policy** — inclusion, display order, and required-ness (built-in and custom uniformly). Required-ness leaves the field definition.
3. **Identity is preserved** — every existing field definition keeps its FieldID through the move; facts keyed by FieldID are untouched.
4. **Events-only consumption** — OnePagers consumes MetaModel's published schema events into a local, backfilled cache; no live queries.
5. **Permissions unchanged** — schema management stays behind `meta-model:write` (already the permission on today's definition endpoints).

---

## Acceptance Criteria

- [ ] MetaModel exposes attribute-schema management per subject type; OnePagers' definition endpoints are removed
- [ ] OnePagers reads definitions from its local cache of MetaModel events, seeded by a backfill migration
- [ ] A one-time transfer moves every existing definition into MetaModel with FieldIDs preserved; recorded facts resolve unchanged
- [ ] Required-ness is managed on the OnePagers configuration for built-in and custom fields uniformly
- [ ] Retiring an attribute in MetaModel removes it from editing while keeping recorded facts readable

---

## Architecture

### Ownership

MetaModel gains the schema; OnePagers becomes a downstream consumer for definitions while keeping facts, completeness, display order, inclusion, and required-ness.

### Domain Model

New MetaModel aggregate per tenant and subject type holding the attribute set (mirroring the granularity of OnePagers' per-subject-type configuration, avoiding contention on the single MetaModelConfiguration aggregate). Published events: `SubjectAttributeDefined`, `SubjectAttributeRenamed`, `SubjectAttributeRetired`, `SubjectAttributeReactivated`, `SubjectAttributeOptionAdded`, `SubjectAttributeOptionRetired`, `SubjectAttributeBoundsChanged`. OnePagers' configuration aggregate drops schema mutation; its field references carry required-ness for custom fields the way they already do for built-ins.

### API Surface

MetaModel: attribute management under the meta-model resource per subject type. OnePagers: definition endpoints (define/rename/retire/reactivate/options/bounds) removed; requirement, inclusion, and display-order endpoints remain.

### Persistence

OnePagers schema cache table projected from MetaModel events (the maturity-scale-cache pattern: projector plus backfill migration). One-time transfer seeds MetaModel aggregates from the existing OnePagers configuration state via MetaModel's own commands, preserving FieldIDs; OnePagers' historical schema events remain inert in the store.

### Frontend

The one-pager configuration page keeps its place; schema operations call MetaModel endpoints, presentation operations call OnePagers. Cache invalidation spans both features' query keys.

### Cross-Context Integration

New published language: MetaModel schema events consumed by OnePagers. No other context consumes them in this slice.

---

## Design Decisions

1. **Per-subject-type schema aggregate in MetaModel** — matches consumer granularity and bounds transaction contention. Alternative: extend the tenant-wide MetaModelConfiguration aggregate (rejected: one aggregate would serialize all schema edits and bloat an already multi-purpose aggregate).
2. **Required-ness moves to presentation, not schema** — required-ness drives completeness, a OnePagers judgement; the field VO currently couples them and this is the seam to cut. Alternative: keep required on the definition in MM (rejected: re-creates G1 one level down).
3. **Seeding through MetaModel commands with preserved IDs** — same mechanism the Importing context proved for cross-context data movement; avoids writing synthetic events directly into the store.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Two APIs behind one config page | The page orchestrates two contexts | HATEOAS links on the configuration response name both surfaces |
| Historical schema events stay in OnePagers streams | Two homes for schema history in the event store | Audit trail remains truthful; new history accrues only in MetaModel |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
