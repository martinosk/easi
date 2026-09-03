# 220 — Capability Details: One Surface, Edited In Place

> **Status:** pending
> **Depends on:** 042 (capability UI consistency), 096 (maturity slider), 114 (capability experts), 191 (one-pager subject drawer), 200 (EA owner name display), 219 (application details edited in place)
> **Roadmap alignment:** outside roadmap: presentation consistency for the capability record, the second slice of the in-place editing rule that 219 settled for applications; no domain change. Respects SD6's principle that every affordance is a HATEOAS link.

---

## Problem Statement

Spec 219 settled one rule for applications: one detail surface, every field edited where it is displayed, every edit control gated by the link that authorises it, no whole-record edit dialog. Capabilities still follow the old model on every surface that shows them.

A capability is viewed on three surfaces: the Architecture Canvas details pane, the Business Domains capability drawer, and the one-pager subject drawer. The canvas pane and the subject drawer share a body, but each wraps it in its own "Capability Details" heading and owns its own copy of the edit and add-expert dialog state. The Business Domains drawer renders a bespoke details section that omits status, maturity value, created date and the realized-capability list the other surfaces show. All three keep a whole-record "Edit" button that opens a modal editing name, description, status, maturity, ownership model, primary owner and EA owner together, with a second experts list and the only tag editor embedded inside it. Experts are edited in place; everything else is not. A user who learned the application rule in 219 meets the opposite rule on the next node they click.

This spec extends 219's rule to capabilities: one capability panel, every field edited in place, gated by its link, and the capability edit dialog retires.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | Fix a capability's name, description, status, maturity, ownership or tags the moment they see it wrong, on whichever surface they are looking at |
| **Invited editor (190 grantee)** | The same edit controls on the canvas, the domain board and the one-pager |
| **Read-only viewer** | A clean read-only panel with no disabled or hidden-then-revealed controls |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One capability detail surface, edited in place

  Scenario Outline: Every surface shows the same capability sections in the same order
    Given capability "Order Management" has a description, a status, a maturity value, an ownership model, a primary owner, an EA owner, two tags, two experts and one realising application
    When I open its details on the <surface>
    Then I see, in order: name, description, level, status, maturity, ownership model, primary owner, EA owner, tags, experts, created, realising applications, one-pager action, history
    And no section present on one surface is absent on another

    Examples:
      | surface                      |
      | Architecture Canvas          |
      | Business Domains drawer      |
      | One-pager subject drawer     |

  Scenario: The canvas adds a view-membership section, other surfaces do not
    Given "Order Management" is placed on the current view
    When I open its details on the Architecture Canvas
    Then an "In this view" section, placed after realising applications and before the one-pager action, shows the custom colour control and "Remove from view"
    And that section is absent on the Business Domains drawer and the one-pager subject drawer

  Scenario: The Business Domains drawer adds domain-context sections, other surfaces do not
    Given "Order Management" is opened from the domain board of "Sales"
    When the drawer opens
    Then the journeys section and the strategic importance section for "Sales" render above the capability's own fields
    And those sections are absent on the Architecture Canvas and the one-pager subject drawer

  Scenario: Rename in place
    Given the capability resource carries an "edit" link
    When I activate the name field, type "Order Handling" and confirm
    Then the capability is renamed without a modal
    And the name updates on the canvas node, the domain board and the navigation tree

  Scenario: Edit the description in place
    Given the capability resource carries an "edit" link
    When I activate the description field, change the text and confirm
    Then the new description is saved and shown in read mode

  Scenario: Empty description invites an edit
    Given the capability has no description
    And the capability resource carries an "edit" link
    Then the description field shows an "Add a description" prompt that starts editing

  Scenario: Change status in place
    Given the capability resource carries an "x-update-metadata" link
    When I activate the status field, pick "Retiring" and confirm
    Then the status is saved and shown as a badge in read mode
    And the maturity, ownership model, primary owner and EA owner are unchanged

  Scenario: Change maturity in place
    Given the capability resource carries an "x-update-metadata" link
    When I activate the maturity field, move the slider to 42 and confirm
    Then the maturity badge shows the section name for 42 and the value 42

  Scenario: Change ownership in place
    Given the capability resource carries an "x-update-metadata" link
    When I activate the ownership model field and pick "Federated" and confirm
    And I activate the primary owner field, type "Sales Platform Team" and confirm
    And I activate the EA owner field, pick a user and confirm
    Then each field is saved on its own confirm and shown in read mode
    And the EA owner shows the user's display name, never the id

  Scenario: Add a tag in place
    Given the capability resource carries an "x-add-tag" link
    When I activate the tag input, type "core" and confirm
    Then "core" appears among the tags without a modal

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

  Scenario: A rejected EA owner stays open with the server's message
    Given I am editing the EA owner
    When I confirm a value the server rejects as ambiguous
    Then the field stays in edit mode and shows the server's message

  Scenario: Read-only viewer sees no edit controls
    Given the capability resource carries no "edit", "x-update-metadata", "x-add-tag" or "x-add-expert" links
    When I open its details on any surface
    Then every field renders as plain text or a badge
    And there is no edit control, tag input, add-expert or remove-expert control

  Scenario: No whole-record edit dialog remains
    When I open capability details on any surface, or right-click a capability in the navigation tree
    Then no "Edit" action opens a modal form
    And the tree's "Edit" action selects the capability so its details pane shows it

  Scenario: An edit on one surface is fresh on the others
    Given "Order Management" is open on the canvas details pane and on the domain board drawer
    When I change its status in the drawer
    Then the canvas pane shows the new status on next render without reopening
```

---

## Business Rules & Invariants

1. **One surface** — exactly one capability detail panel component exists; each host (canvas details pane, Business Domains drawer, one-pager subject drawer) contributes only its container and supplies the capability id. Hosts cannot opt sections out; the panel owns its dialogs and its data fetching.
2. **Field-owned affordances** — every editable field carries its own control, rendered exactly when the link that authorises it is present: `edit` for name and description; `x-update-metadata` for status, maturity, ownership model, primary owner and EA owner; `x-add-tag` for tags; `x-add-expert` and the expert's `x-remove` for experts. No control is rendered disabled as a substitute for absence.
3. **No whole-record edit mode** — there is no panel-level "Edit" button, no edit-mode toggle, and no modal that edits several capability fields together. The capability edit dialog, its form fields, its form hook and the add-tag dialog are removed along with their call sites.
4. **Focused dialogs only for compound input** — the only dialog opened from the panel is add expert (name, role, contact). A single value, including a tag, the EA owner and the maturity value, never opens a dialog.
5. **Per-field commit** — each in-place field commits on confirm (Enter or the field's confirm control) and discards on cancel (Escape or the field's cancel control). A failed request keeps the field in edit mode and shows the error; a rejected value never reaches the API.
6. **Metadata commits are whole** — a single metadata field commit sends the capability's current status, maturity value, ownership model, primary owner and EA owner unchanged alongside the edited value, because the metadata request has no partial semantics and an omitted maturity value is read as zero.
7. **Same validation as creation** — name, description, status and maturity use the same schemas and bounds as the create capability form; the maturity bounds come from the maturity scale.
8. **Level is structural** — the level renders as a read-only badge; it is derived from the hierarchy and is not edited on the panel.
9. **Host slots are named and placed** — the canvas contributes one "In this view" section (custom colour, "Remove from view") after realising applications and before the one-pager action, only when the capability is on the current view, gated by the view-capability's own links. The Business Domains drawer contributes its domain-context sections (journeys, strategic importance) above the capability's own fields. No other host content exists.
10. **One realising-applications section** — the realising-applications list, with its assessment and role controls gated by their own links, is part of the panel and renders identically on every host.
11. **Freshness on every host** — all capability reads go through TanStack Query; the existing mutation effects refresh every open host. No host fetches the capability outside the query layer.
12. **One heading** — the panel renders the capability name as its heading; a host adds a container title at most, never a second "Capability Details" heading.
13. **Tree Edit selects** — the navigation tree's context-menu "Edit" selects the capability so the details pane shows it.
14. **Links name their targets** — the capability resource emits `x-update-metadata` (PUT metadata) and `x-add-tag` (POST tags) under the same write-or-grant rule that already emits `edit` and `x-add-expert`; the frontend follows those links and keeps no hard-coded metadata or tag URL.
15. **Creation is unchanged** — the create capability dialog stays; this spec covers editing an existing record only.

---

## Acceptance Criteria

- [ ] The canvas details pane, the Business Domains drawer and the one-pager subject drawer render the same panel component with the same sections in the same order; a test per host asserts the status and experts sections are present.
- [ ] Name and description are edited in place, gated on `edit`; confirm persists through the existing update request, cancel sends nothing, an empty name is rejected client-side with the message shown in the field.
- [ ] Status, maturity, ownership model, primary owner and EA owner are edited in place, gated on `x-update-metadata`; each commit sends the full current metadata set with one field changed; a server rejection keeps the field open with the server's message.
- [ ] Tags are added in place, gated on `x-add-tag`; the add-tag dialog is removed.
- [ ] With none of the links present, every field renders as text or a badge with no edit control.
- [ ] The capability edit dialog, its form fields, its form hook, the `edit-capability` dialog type, the canvas dialog opener and every Edit button (canvas pane, subject drawer panel, Business Domains drawer) are removed; the tree context-menu "Edit" selects the capability.
- [ ] Custom colour and "Remove from view" render in one "In this view" section only on the canvas when the capability is on the current view.
- [ ] The Business Domains drawer renders the shared panel plus its journeys and strategic importance sections; its bespoke details section is removed.
- [ ] The realising-applications section, with link-gated assessment and role controls, renders on every host.
- [ ] Exactly one experts list is mounted per open panel; each host renders one heading for the capability.
- [ ] Backend: the capability resource carries `x-update-metadata` and `x-add-tag` when the actor may write or holds an edit grant, and neither otherwise; link tests cover both cases.
- [ ] E2E (mock-mode Playwright project, `e2e/mock/capability-details.spec.ts`): rename in place on the canvas with the tree following, and change status from a domain board drawer with the canvas pane reflecting it.

---

## Architecture

### Ownership

Frontend in the capabilities feature. Capability Mapping's contract gains two additive links on the capability resource; no new endpoints, no request or response shape change. Business Domains and OnePagers keep hosting the panel but stop owning any of its behaviour.

### Domain Model

No change.

### API Surface

Two additive links on the capability resource, emitted alongside `edit` and `x-add-expert` under the existing write-or-grant rule:

| Link | Method | Target |
|------|--------|--------|
| `x-update-metadata` | PUT | `/capabilities/{id}/metadata` |
| `x-add-tag` | POST | `/capabilities/{id}/tags` |

Both endpoints already exist and already sit behind the matching authorisation middleware.

### Persistence

No change.

### Frontend

- One panel component takes a capability id, resolves the capability through the query layer, and renders the sections in the scenario order. It owns the add-expert dialog.
- The in-place text field primitive from 219 serves name, description and primary owner. Two sibling primitives cover the remaining single-value fields: an in-place select (status, ownership model, EA owner) and an in-place maturity field wrapping the existing maturity slider, both with the same confirm, cancel and error behaviour.
- The tags section renders the existing tag list and, when `x-add-tag` is present, an inline tag input with confirm and cancel.
- The metadata mutation hook follows `x-update-metadata`; the tag mutation hook follows `x-add-tag`.
- Hosts become thin: the canvas details pane passes the "In this view" section as its one slot; the Business Domains drawer passes its journeys and strategic importance sections as its slot; the one-pager subject drawer passes nothing.
- The capability edit dialog, its form fields, its form hook, the add-tag dialog, the `edit-capability` dialog-manager entry and the canvas hook that opens it are deleted; the tree's Edit handler dispatches capability selection instead.

### Cross-Context Integration

None. Spec 191's freshness rule continues to hold through the existing mutation effects.

---

## Design Decisions

1. **Same rule as 219** — the panel already edits experts in place; extending that to every other field gives capabilities the rule applications have. Alternatives considered: _keep the dialog for the metadata fields only_ (rejected: this recreates the two-model split 219 removed, with status behind a modal and experts inline).
2. **Links for metadata and tags instead of reusing `edit`** — the backend authorises the metadata and tag endpoints separately from the main update, and the frontend currently hard-codes both URLs. Emitting a link per affordance keeps rule 2 literal and removes the hard-coded paths. Alternative: _gate all three on `edit`_ (rejected: the frontend would still need to know the URLs, and a future permission split would silently mis-gate).
3. **Whole metadata commit per field** — the metadata request replaces every metadata field and reads an omitted maturity as zero, so a single-field commit must carry the others. Alternative: _add PATCH semantics to the endpoint_ (rejected: a backend contract change for a presentation spec; the whole commit is idempotent on the unchanged fields).
4. **Single-value selects stay inline** — status, ownership model and EA owner each pick one value, so they get an in-place select rather than a dialog. The EA owner list is the same user lookup the dialog used.
5. **One realising-applications section everywhere** — the drawer's richer list (assessment, role) differs from the canvas list only by controls that are already link-gated, so it becomes the panel's list. Alternative: _keep two lists and let the drawer suppress the panel's_ (rejected: hosts cannot opt sections out).
6. **Domain-context sections stay with the drawer** — journeys and strategic importance need the domain, so they are the drawer's slot, named and placed, mirroring the canvas's "In this view".
7. **Level stays read-only** — it is derived from the hierarchy and changed by re-parenting, not by editing the record.
8. **Capabilities only** — origin entities (acquired entity, vendor, internal team) and canvas edges (relation, realization) still have whole-record edit dialogs. They are covered by 221 and 222 so each slice ships and validates on its own.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Per-field commit for metadata | Five requests to change five metadata fields, instead of one | Each request is small and idempotent; changing several metadata fields at once is rare |
| Removing the edit dialog | Users who learned "Edit" lose a familiar button | Pencil affordance on hover and empty-field prompts make the in-place controls discoverable; the tree keeps its Edit item |
| Two new links | Backend touch for a presentation change | Links only, no endpoint or shape change, covered by link tests |
| Richer realising-applications list on every host | Assessment and role controls appear on the canvas for the first time | They are gated by their own links, so they appear only where already permitted |

---

## Checklist

- [x] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
