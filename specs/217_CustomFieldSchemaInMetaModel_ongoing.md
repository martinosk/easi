# 217 — Custom-Field Schema in MetaModel

> **Status:** ongoing
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

- [x] MetaModel exposes attribute-schema management per subject type; OnePagers' definition endpoints are removed
- [x] OnePagers reads definitions from its local cache of MetaModel events, seeded by a backfill migration
- [x] A one-time transfer moves every existing definition into MetaModel with FieldIDs preserved; recorded facts resolve unchanged
- [x] Required-ness is managed on the OnePagers configuration for built-in and custom fields uniformly
- [x] Retiring an attribute in MetaModel removes it from editing while keeping recorded facts readable

---

## Architecture

### Ownership

MetaModel gains the schema; OnePagers becomes a downstream consumer for definitions while keeping facts, completeness, display order, inclusion, and required-ness.

### Domain Model

New MetaModel aggregate `SubjectAttributeSchema` per tenant and subject type holding the attribute set (mirroring the granularity of OnePagers' per-subject-type configuration, avoiding contention on the single MetaModelConfiguration aggregate); created lazily on first read, with an internal `SubjectAttributeSchemaCreated` event. Published events: `SubjectAttributeDefined` (carries options and number bounds), `SubjectAttributeRenamed`, `SubjectAttributeRetired`, `SubjectAttributeReactivated`, `SubjectAttributeOptionAdded`, `SubjectAttributeOptionRetired`, `SubjectAttributeBoundsChanged`. OnePagers' configuration aggregate drops schema mutation and the custom-field value object; it keeps the display order plus a required flag per custom field (the way it already does for built-ins) and gains `IncludeCustomField` / `ExcludeCustomField`, raised only by an in-context reactor on `SubjectAttributeDefined` / `Reactivated` / `Retired` so the display order stays in step with the schema (the reactor creates the configuration when the subject type has none yet). Legacy `CustomFieldDefined` and related events still replay their inclusion and requirement effects.

### API Surface

MetaModel: `/meta-model/subject-types/{subjectType}/attributes` — GET (lazy create, `meta-model:read`), POST define, PUT `/{attributeID}` rename, POST `/{attributeID}/retire|reactivate`, POST `/{attributeID}/options`, POST `/{attributeID}/options/{optionID}/retire`, PUT `/{attributeID}/bounds` (`meta-model:write`, optimistic `version`). OnePagers: definition endpoints (define/rename/retire/reactivate/options/bounds) removed; requirement, inclusion, and display-order endpoints remain. The configuration response keeps `customFields` (name, type, options, bounds, active from the cache; `required` and `included` from the configuration) and names the schema surface with `x-attribute-schema`; custom fields carry only `x-set-requirement`.

### Persistence

`metamodel.subject_attribute_schemas` read model (migration 159). `onepagers.custom_field_definition_cache` (migration 160) projected from MetaModel events (the maturity-scale-cache pattern), backfilled by migration 161 from the definitions embedded in the legacy configuration documents with `pending_transfer = TRUE`. At startup OnePagers dispatches MetaModel's published `ImportSubjectAttribute` command for every pending row (preserving FieldID, option IDs, bounds and active state); MetaModel's resulting events re-project the row and clear the flag, so the transfer is idempotent and self-confirming. OnePagers' historical schema events remain inert in the store.

### Frontend

The one-pager configuration page keeps its place; it reads the configuration and, through `x-attribute-schema`, the MetaModel schema. Schema operations follow the schema's and attributes' links with the schema version; presentation operations follow configuration links with the configuration version. Defining a field sends its bounds in the same request; a field defined as required is marked required on the configuration once the refreshed configuration offers `x-set-requirement`. Schema mutations invalidate the MetaModel schema key and every one-pager configuration key.

### Cross-Context Integration

New published language: MetaModel schema events consumed by OnePagers. No other context consumes them in this slice.

---

## Design Decisions

1. **Per-subject-type schema aggregate in MetaModel** — matches consumer granularity and bounds transaction contention. Alternative: extend the tenant-wide MetaModelConfiguration aggregate (rejected: one aggregate would serialize all schema edits and bloat an already multi-purpose aggregate).
2. **Required-ness moves to presentation, not schema** — required-ness drives completeness, a OnePagers judgement; the field VO currently couples them and this is the seam to cut. Alternative: keep required on the definition in MM (rejected: re-creates G1 one level down).
3. **Seeding through MetaModel commands with preserved IDs** — same mechanism the Importing context proved for cross-context data movement; avoids writing synthetic events directly into the store.
4. **Inclusion follows the schema through a reactor** — defining an attribute puts it on the one-pager, retiring removes it, exactly as before the move; the OnePagers aggregate keeps its display-order invariant without knowing the schema. Alternative: manual include/exclude endpoints for custom fields (rejected: a second step for every definition and stale references in the display order).
5. **Transfer marker lives in the cache row** — `pending_transfer` makes the startup transfer idempotent without a separate bookkeeping table and without MetaModel reading OnePagers' tables.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Two APIs behind one config page | The page orchestrates two contexts | HATEOAS links on the configuration response name both surfaces |
| Historical schema events stay in OnePagers streams | Two homes for schema history in the event store | Audit trail remains truthful; new history accrues only in MetaModel |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (cross-context define/retire flow and legacy transfer; compiled, to be run against the compose database)
- [x] API documentation updated
- [ ] User sign-off
