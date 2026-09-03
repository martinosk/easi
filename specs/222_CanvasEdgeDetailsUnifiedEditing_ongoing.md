# 222 — Canvas Edge Details: One Surface, Edited In Place

> **Status:** ongoing
> **Depends on:** 219 (application details edited in place), 220 (capability details edited in place)
> **Roadmap alignment:** outside roadmap: presentation consistency for the two editable canvas edges (application relation, capability realization), the last slice of the in-place editing rule that 219 settled; no domain change. Respects SD6's principle that every affordance is a HATEOAS link, and fixes two places where it was not.

---

## Problem Statement

Spec 219 settled one rule for applications, and 220 and 221 extend it to capabilities and origin entities: every field edited where it is displayed, every edit control gated by the link that authorises it, no whole-record edit dialog. The two editable canvas edges still follow the old model.

Selecting a relation edge on the Architecture Canvas shows a details pane with a whole-record "Edit" button that opens a modal editing name and description together, wired through the canvas dialog manager. Selecting a realization edge shows a pane whose "Edit" button opens a modal editing realization level and notes together, with the dialog state held inside the pane. Both are the last "Edit" buttons left on the canvas once 219 to 221 ship.

Both edges also break the link rule on the backend: the relation resource and the realization resource emit `edit` and `delete` for every actor, so a read-only viewer sees an Edit button whose request is rejected. Origin relationship edges are read-only by design and already carry no edit affordance.

This spec extends the rule to relation and realization edges and makes their links honest.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | Rename a relation, describe it, or change a realization level the moment they see it wrong on the canvas |
| **Read-only viewer** | A clean read-only pane with no control that fails on click |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Canvas edge details edited in place

  Scenario: Relation pane sections
    Given a "Serves" relation from "Billing Engine" to "Order Management" with a name and a description
    When I select its edge on the Architecture Canvas
    Then I see, in order: name, type, source, target, description, created, reference documentation, history

  Scenario: Realization pane sections
    Given "Billing Engine" realizes capability "Order Management" at level "Partial" with notes
    When I select its edge on the Architecture Canvas
    Then I see, in order: capability, application, realization level, origin, notes, linked, history

  Scenario: Rename a relation in place
    Given the relation resource carries an "edit" link
    When I activate the name field, type "Sends invoices" and confirm
    Then the relation is renamed without a modal
    And the edge label on the canvas updates

  Scenario: Describe a relation in place
    Given the relation resource carries an "edit" link
    And the relation has no description
    Then the description field shows an "Add a description" prompt that starts editing
    When I type a description and confirm
    Then the description is saved and shown in read mode

  Scenario: Change a realization level in place
    Given the realization resource carries an "edit" link
    When I activate the realization level field, pick "Full" and confirm
    Then the level badge shows "Full"
    And the notes are unchanged

  Scenario: Add realization notes in place
    Given the realization resource carries an "edit" link
    When I activate the notes field, type "Covers invoicing only" and confirm
    Then the notes are saved and shown in read mode

  Scenario: Inherited realizations are read-only
    Given a realization whose origin is "Inherited"
    When I select its edge
    Then the level and notes render as read-only
    And the inherited explanation is shown

  Scenario: Cancel discards
    Given I am editing any edge field
    When I cancel
    Then the field returns to read mode showing the previous value
    And no request is sent

  Scenario: Read-only viewer sees no edit controls
    Given the actor cannot write components or capabilities
    When I select a relation edge or a realization edge
    Then the resource carries no "edit" link
    And every field renders as plain text or a badge with no edit control

  Scenario: No whole-record edit dialog remains
    When I select any edge on the Architecture Canvas
    Then no "Edit" action opens a modal form
```

---

## Business Rules & Invariants

1. **One pane per edge kind** — the relation pane and the realization pane each own their data fetching and render their sections in the scenario order; the canvas details renderer supplies only the edge id.
2. **Field-owned affordances** — name and description of a relation, and level and notes of a direct realization, each carry their own control, rendered exactly when `edit` is present on the resource. No control is rendered disabled as a substitute for absence.
3. **No whole-record edit mode** — there is no pane-level "Edit" button and no modal that edits several edge fields together. The relation edit dialog, its dialog-manager entry, its canvas opener, and the realization edit dialog are removed.
4. **No dialogs** — every edge field is a single value, so no field opens a dialog.
5. **Per-field commit** — each in-place field commits on confirm and discards on cancel. A failed request keeps the field in edit mode and shows the error; a rejected value never reaches the API.
6. **Whole-record request per field** — the relation update sends name and description; the realization update sends level and notes. A single-field commit sends the other field unchanged.
7. **Same validation as today** — relation name and description use the existing relation schemas; realization level and notes use the existing realization schemas.
8. **Inherited realizations are read-only** — an inherited realization renders no edit control regardless of links, as today.
9. **Honest links** — the relation resource emits `edit` only when the actor may write components and `delete` only when the actor may delete them; the realization resource emits `edit` only when the actor may write capabilities and `delete` only when the actor may delete them. A resource never carries a link whose request the same actor would be refused.
10. **Structure is not edited on the pane** — relation type, source and target, and a realization's capability and application, are read-only; changing them is a delete and re-create, as today.
11. **Origin relationship edges are unchanged** — they remain read-only with their own pane.

---

## Acceptance Criteria

- [x] The relation pane renders name and description in place, gated on `edit`; confirm persists through the existing relation update with the other field unchanged, cancel sends nothing.
- [x] The realization pane renders level as an in-place select over Full, Partial and Planned, and notes as an in-place text field, both gated on `edit` and on the realization being direct; confirm persists through the existing realization update with the other field unchanged.
- [x] Empty relation description and empty realization notes render an "Add …" prompt when editable, and nothing when not.
- [x] With no `edit` link, or for an inherited realization, every field renders as text or a badge with no edit control.
- [x] The relation edit dialog, the `edit-relation` dialog type, its dialog-manager entry and its canvas opener, and the realization edit dialog with the pane's dialog state, are removed. The canvas dialog hook no longer exposes an edit opener of any kind.
- [x] Backend: relation links emit `edit` and `delete` only for actors with the matching components permission; realization links emit `edit` and `delete` only for actors with the matching capabilities permission. Link tests cover both the permitted and the refused actor for each resource.
- [x] Frontend tests exist for both panes covering editable, read-only and inherited states; the relation pane test replaces the absent edit dialog test.
- [x] E2E (mock-mode Playwright project, `e2e/mock/edge-details.spec.ts`): rename a relation in place with the edge label following, and change a realization level in place.

---

## Architecture

### Ownership

Frontend in the relations feature. Architecture Modeling and Capability Mapping each gain actor-aware link emission for one resource; no endpoints, request or response shapes change.

### Domain Model

No change.

### API Surface

No new endpoints. Two link corrections:

| Resource | Today | After |
|----------|-------|-------|
| Relation | `edit`, `delete` always | `edit` when the actor may write components; `delete` when the actor may delete components |
| Capability realization | `edit`, `delete` always | `edit` when the actor may write capabilities; `delete` when the actor may delete capabilities |

Both match the middleware already guarding those routes.

### Persistence

No change.

### Implementation Notes

- The relation and realization link builders now take the actor, and every handler that emits those links reads the actor from the request; the unconditional builders are gone.
- The mock API persists relation and realization updates so the panes' tests and the mock-mode Playwright project can exercise the in-place flows.

### Frontend

- The relation pane renders name and description with the in-place text field from 219, and drops its edit action and the dialog-opening prop.
- The realization pane renders level with the in-place select from 220 and notes with the in-place text field, and drops its dialog state.
- The relation edit dialog, its dialog type, its dialog-manager entry and the canvas dialog hook's opener are deleted; the realization edit dialog is deleted.
- The canvas details renderer stops threading an edit callback to edge panes.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Same rule as 219** — edges have two single-value fields each; in-place editing removes the dialog with nothing left over.
2. **Fix the links in the same spec** — the frontend rule "render the control exactly when the link is present" is only correct if the link is honest. Leaving the backend as is would make read-only viewers see pencils that fail. Alternative: _a separate backend spec_ (rejected: the fix is a handful of lines under the same rule, and this spec's read-only scenario cannot pass without it).
3. **Level as a select, not a segmented control** — three values with a confirm step matches the in-place select from 220; a segmented control that commits on click would break the per-field confirm and cancel rule.
4. **Structure stays read-only** — retargeting an edge is a different operation with different invariants, and today it is delete and re-create; this spec does not change that.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Whole-record request per field | Two requests when both fields change | Fields are small; the update request is idempotent on the unchanged field |
| Actor-aware links | Read-only users lose an Edit button that used to fail | It failed before; now the pane is honest |
| Removing the last canvas edit dialogs | The canvas dialog hook shrinks to create dialogs only | Simpler hook; create flows are untouched |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (link unit tests; no handler contract change)
- [x] API documentation updated (links only, no annotation change)
- [ ] User sign-off
