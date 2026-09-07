# 221 — Origin Entity Details: One Surface, Edited In Place

> **Status:** ongoing
> **Depends on:** 191 (one-pager subject drawer), 219 (application details edited in place)
> **Roadmap alignment:** outside roadmap: presentation consistency for the origin-entity records (acquired entity, vendor, internal team), the third slice of the in-place editing rule that 219 settled; no domain change. Respects SD6's principle that every affordance is a HATEOAS link.

---

## Problem Statement

Spec 219 settled one rule for applications and spec 220 extends it to capabilities: one detail surface, every field edited where it is displayed, every edit control gated by the link that authorises it, no whole-record edit dialog. The three origin entity types still follow the old model.

An acquired entity, a vendor or an internal team is viewed on two surfaces: the Architecture Canvas details pane and the one-pager subject drawer. Both render through one shared panel already, but every field is read-only and the only way to change anything is a whole-record "Edit" button that opens one of three modal forms. Each form edits the name, the type-specific fields and the notes together. The "Remove from view" control sits in the same action row as "Edit", although it belongs to the view placement rather than the record. Each type renders its own "… Details" heading above the name. The navigation tree declares Edit handlers for these types that nothing wires, so the tree offers no Edit item at all. None of the three edit dialogs has a test.

This spec extends the rule to origin entities: every field edited in place, gated by `edit`, view membership in its own section, and the three edit dialogs retire.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | Correct a vendor's implementation partner, a team's contact person or an acquisition's integration status the moment they see it wrong |
| **Invited editor (190 grantee)** | The same edit controls on the canvas and from a one-pager |
| **Read-only viewer** | A clean read-only panel with no disabled or hidden-then-revealed controls |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One origin entity detail surface, edited in place

  Scenario Outline: Every surface shows the same sections in the same order
    Given <type> "<name>" has every optional field filled and one related application
    When I open its details on the <surface>
    Then I see, in order: name, <type-fields>, notes, created, type, related applications, one-pager action, history
    And no section present on one surface is absent on another

    Examples:
      | type            | name          | type-fields                             | surface                  |
      | acquired entity | Nordic Cargo  | acquisition date, integration status    | Architecture Canvas      |
      | acquired entity | Nordic Cargo  | acquisition date, integration status    | One-pager subject drawer |
      | vendor          | SAP           | implementation partner                  | Architecture Canvas      |
      | vendor          | SAP           | implementation partner                  | One-pager subject drawer |
      | internal team   | Platform Team | department, contact person              | Architecture Canvas      |
      | internal team   | Platform Team | department, contact person              | One-pager subject drawer |

  Scenario: The canvas adds a view-membership section, the subject drawer does not
    Given vendor "SAP" is placed on the current view
    When I open its details on the Architecture Canvas
    Then an "In this view" section, placed after related applications and before the one-pager action, shows "Remove from view"
    And that section is absent on the one-pager subject drawer

  Scenario: Rename in place
    Given the entity resource carries an "edit" link
    When I activate the name field, type "SAP SE" and confirm
    Then the entity is renamed without a modal
    And the name updates on the canvas node and the navigation tree

  Scenario: Edit a text field in place
    Given the entity resource carries an "edit" link
    When I activate the notes field, change the text and confirm
    Then the new notes are saved and shown in read mode

  Scenario: Empty optional field invites an edit
    Given internal team "Platform Team" has no department
    And the entity resource carries an "edit" link
    Then the department field shows an "Add a department" prompt that starts editing

  Scenario: Change integration status in place
    Given acquired entity "Nordic Cargo" carries an "edit" link
    When I activate the integration status field, pick "Completed" and confirm
    Then the status badge shows "Completed"
    And the name, acquisition date and notes are unchanged

  Scenario: Change acquisition date in place
    Given acquired entity "Nordic Cargo" carries an "edit" link
    When I activate the acquisition date field, pick a date and confirm
    Then the date is saved and shown formatted in read mode

  Scenario: Cancel discards
    Given I am editing any field
    When I cancel
    Then the field returns to read mode showing the previous value
    And no request is sent

  Scenario: Invalid name stays open with an error
    Given I am editing the name
    When I clear the name and confirm
    Then the field stays in edit mode and shows the validation message
    And no request is sent

  Scenario: Read-only viewer sees no edit controls
    Given the entity resource carries no "edit" link
    When I open its details on any surface
    Then every field renders as plain text or a badge
    And there is no edit control and no empty-field prompt

  Scenario: No whole-record edit dialog remains
    When I open origin entity details on any surface
    Then no "Edit" action opens a modal form

  Scenario: An edit on one surface is fresh on the others
    Given vendor "SAP" is open on the canvas details pane and on a one-pager subject drawer
    When I change its implementation partner in the drawer
    Then the canvas pane shows the new value on next render without reopening
```

---

## Business Rules & Invariants

1. **One surface** — exactly one origin entity detail panel exists, parameterised by entity type; each host (canvas details pane, one-pager subject drawer) contributes only its container and supplies the type and id. Hosts cannot opt sections out; the panel owns its data fetching.
2. **Field-owned affordances** — every editable field carries its own control, rendered exactly when `edit` is present on the entity resource. No control is rendered disabled as a substitute for absence.
3. **No whole-record edit mode** — there is no panel-level "Edit" button, no edit-mode toggle, and no modal that edits several fields together. The three edit dialogs are removed along with their call sites and the unwired tree Edit handlers.
4. **No dialogs** — every origin entity field is a single value (text, date or status), so no field opens a dialog.
5. **Per-field commit** — each in-place field commits on confirm (Enter or the field's confirm control) and discards on cancel (Escape or the field's cancel control). A failed request keeps the field in edit mode and shows the error; a rejected value never reaches the API.
6. **Whole-record request per field** — the update request for each type replaces the record, so a single-field commit sends the entity's other fields unchanged alongside the edited value.
7. **Same validation as creation** — every field uses the same schema as the corresponding create dialog; the shared name and notes schemas are exported so the in-place fields can reuse them.
8. **Editable fields per type** — acquired entity: name, acquisition date, integration status, notes. Vendor: name, implementation partner, notes. Internal team: name, department, contact person, notes. Created, type and related applications are read-only.
9. **View membership is a separate section** — "Remove from view" is a property of the view placement. It renders in one "In this view" section after related applications and before the one-pager action, only when the panel is hosted on the canvas and the entity is on the current view, gated by the view-origin-entity's `x-remove` link.
10. **Freshness on every host** — all reads go through TanStack Query; the existing mutation effects refresh every open host.
11. **One heading** — the panel renders the entity name as its heading; a host adds a container title at most, never a second "… Details" heading. The type label stays as a read-only field.
12. **Creation is unchanged** — the three create dialogs stay; this spec covers editing an existing record only.

---

## Acceptance Criteria

- [x] The canvas details pane and the one-pager subject drawer render the same panel for each of the three types with the same sections in the same order; a test per type and host asserts the type-specific fields are present.
- [x] Every field in rule 8 is edited in place, gated on `edit`; confirm persists through the existing update request with the other fields unchanged, cancel sends nothing, an empty name is rejected client-side with the message shown in the field.
- [x] Integration status is an in-place select over the four known statuses; acquisition date is an in-place date field; both share the confirm, cancel and error behaviour of the text field.
- [x] Empty optional fields render an "Add a …" prompt when `edit` is present, and nothing when it is absent.
- [x] With no `edit` link, every field renders as text or a badge with no edit control.
- [x] The three edit dialogs, the panel's dialog state and action row, and the unwired tree Edit handlers for origin entities are removed.
- [x] "Remove from view" renders in one "In this view" section only on the canvas when the entity is on the current view.
- [x] Each host renders one heading for the entity.
- [x] The hand-rolled type switch that picks a fetch hook per type stays behind the query layer; no host fetches an entity outside it.
- [x] E2E (mock-mode Playwright project, `e2e/mock/origin-entity-details.spec.ts`): rename a vendor in place on the canvas with the tree following, and change an acquired entity's integration status in place on the canvas. The mock project has no one-pager page, so the subject-drawer surface is covered by the panel unit tests and the subject drawer test rather than end to end.

---

## Architecture

### Ownership

Frontend only, in the origin-entities feature. Architecture Modeling's contract is unchanged: the same `edit` link and per-type update request carry every field; a single-field commit sends the other fields unchanged.

### Domain Model

No change.

### API Surface

No change. The `edit` link is already emitted under the write-or-grant rule for all three types.

### Persistence

No change.

### Implementation Notes

- The acquired-entity status badge keyed on form-style values while the API returns upper-case constants, so every status rendered as its raw value; the panel now keys on the API values.
- The per-type update requests follow the entity's `edit` link instead of a hard-coded path, matching 219 and 220.
- The mock API now serves and persists single origin entities and derives the origin-relationships response from the mock database, so the panel tests and the mock-mode Playwright project can exercise the in-place flows.

### Frontend

- One panel component takes an entity type and id, resolves the entity through the query layer, and renders the sections in the scenario order.
- The in-place text field primitive from 219 serves name, notes, implementation partner, department and contact person. The in-place select from 220 serves integration status. A sibling in-place date field, wrapping the existing date input, serves acquisition date with the same confirm, cancel and error behaviour.
- Each type has a small field-definition table (label, schema, value accessor, request builder) so the panel renders type-specific fields without per-type components carrying their own headings and action rows.
- The view-membership section is passed in by the canvas host only; the subject drawer passes nothing.
- The three edit dialogs, the panel's dialog state and action row, and the tree's origin entity Edit handler props are deleted.

### Cross-Context Integration

None. Spec 191's freshness rule continues to hold through the existing mutation effects.

---

## Design Decisions

1. **Same rule as 219** — origin entities have only single-value fields, so in-place editing removes the dialog with nothing left over. Alternative: _keep the dialogs because there is no inline editing to be consistent with yet_ (rejected: the user meets these nodes on the same canvas and one-pager surfaces as applications and capabilities).
2. **A date field primitive** — acquisition date is the first date edited in place. Wrapping the existing date input with the shared confirm and cancel chrome keeps one interaction model. Alternative: _plain text input for the date_ (rejected: the create dialog uses a date input, and rule 7 asks for the same validation).
3. **Field-definition table instead of three panels** — the three details components differ only in their field list; a table per type removes the three copies of heading, action row and history wiring. Alternative: _keep three components and edit each_ (rejected: three places to drift for one rule).
4. **Drop the unwired tree Edit handlers rather than wire them** — the tree already offers delete and click-to-select for origin entities, and 219's tree Edit maps to selection anyway; keeping dead props invites a future dialog.
5. **Origin entities only** — canvas edges (relation, realization) still have whole-record edit dialogs and get their own spec (222).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Whole-record request per field | Two requests when two fields change, instead of one | Fields are small; the update request is idempotent on unchanged fields |
| Removing the edit dialogs | Users who learned "Edit" lose a familiar button | Pencil affordance on hover and empty-field prompts make the in-place controls discoverable |
| Field-definition table | Type-specific rendering becomes data rather than JSX | Definitions are typed per entity type; a wrong accessor fails at compile time |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (no backend change)
- [x] API documentation updated (no API change)
- [ ] User sign-off
