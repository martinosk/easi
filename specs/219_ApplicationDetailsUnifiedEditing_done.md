# 219 — Application Details: One Surface, Edited In Place

> **Status:** done
> **Depends on:** 077 (application details reused on Business Domains), 114 (application experts), 191 (one-pager subject drawer), 214 (application ownership), 215 (application hosting)
> **Roadmap alignment:** outside roadmap: presentation consistency for the application record that H1-2 (214–216) is growing; no domain change, no new contract. Respects SD6 — every affordance stays a HATEOAS link.

---

## Problem Statement

An application is viewed and edited on three surfaces: the Architecture Canvas details pane, the Business Domains drawer, and the one-pager subject drawer. All three render the same body, yet they are not the same. The Business Domains drawer never receives the experts callback, so its experts section is silently absent, and it fetches the application through a hand-rolled effect that TanStack Query never invalidates, so ownership and hosting edits made there can render stale. Spec 077 promised parity with the canvas; spec 191 relies on it.

Within the pane, editing follows two competing models. Ownership, hosting, experts and fit scores are edited where they are shown, each gated by its own link (214, 215, 114). Name and description are the only fields still behind a whole-record "Edit" button that opens a modal form, and that modal embeds a second experts editor, so on the canvas two live experts lists exist at once. A user learns one rule for hosting and another for description.

This spec settles one rule: an application has one detail surface, every field is edited where it is displayed, and every edit control is gated by the link that authorises it. The whole-record edit dialog retires.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | Fix a name, description, owner, hosting or expert the moment they see it wrong, wherever they happen to be looking |
| **Invited editor (190 grantee)** | The same edit controls on whichever surface they arrive at, including from a one-pager |
| **Read-only viewer** | A clean read-only panel with no disabled or hidden-then-revealed controls |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: One application detail surface, edited in place

  Scenario Outline: Every surface shows the same sections in the same order
    Given application "Billing Engine" has a description, an owner, a hosting classification, two experts, one realized capability, and fit scores
    When I open its details on the <surface>
    Then I see, in order: name, description, ownership, hosting, experts, created, type, realized capabilities, origins, fit scores, one-pager action, history
    And no section present on one surface is absent on another

    Examples:
      | surface                     |
      | Architecture Canvas         |
      | Business Domains drawer     |
      | One-pager subject drawer    |

  Scenario: The canvas adds a view-membership section, other surfaces do not
    Given "Billing Engine" is placed on the current view
    When I open its details on the Architecture Canvas
    Then an "In this view" section, placed after the fit scores and before the one-pager action, shows the custom colour control and "Remove from view"
    And that section is absent on the Business Domains drawer and the one-pager subject drawer

  Scenario: Rename in place
    Given the application resource carries an "edit" link
    When I activate the name field and type "Billing Platform" and confirm
    Then the application is renamed without a modal
    And the name updates on the canvas node, the domain board chip and the navigation tree

  Scenario: Edit the description in place
    Given the application resource carries an "edit" link
    When I activate the description field, change the text and confirm
    Then the new description is saved and shown in read mode

  Scenario: Empty description invites an edit
    Given the application has no description
    And the application resource carries an "edit" link
    Then the description field shows an "Add a description" prompt that starts editing

  Scenario: Cancel discards
    Given I am editing the name
    When I cancel
    Then the field returns to read mode showing the previous name
    And no request is sent

  Scenario: Invalid name stays open with an error
    Given I am editing the name
    When I clear the name and confirm
    Then the field stays in edit mode and shows the validation message
    And no request is sent

  Scenario: Read-only viewer sees no edit controls
    Given the application resource carries no "edit", "x-add-expert", "x-classify-hosting" or ownership links
    When I open its details on any surface
    Then the name and description render as plain text
    And hosting renders as a badge, experts have no add or remove control, and ownership shows no action

  Scenario: Experts are editable on Business Domains
    Given the application resource carries an "x-add-expert" link
    When I open its details from a domain board chip
    Then I can add an expert and remove an existing one from the drawer

  Scenario: No whole-record edit dialog remains
    When I open application details on any surface, or right-click an application in the navigation tree
    Then no "Edit" action opens a modal form for name and description
    And the tree's "Edit" action selects the application so its details pane shows it

  Scenario: An edit on one surface is fresh on the others
    Given the application was opened from a domain board chip and was not already cached in the applications list
    When I change its hosting classification in the drawer
    Then the drawer shows the new classification without reopening
    And the tree's hosting facet and the one-pager reflect it on next render
```

---

## Business Rules & Invariants

1. **One surface** — exactly one application detail panel component exists; each host (canvas details pane, Business Domains drawer, one-pager subject drawer) contributes only its container and supplies the application id. Hosts cannot opt sections out; the panel owns its dialogs and its data fetching.
2. **Field-owned affordances** — every editable field carries its own control, rendered exactly when the link that authorises it is present: `edit` for name and description, `x-classify-hosting` for hosting, `x-add-expert` and the expert's `x-remove` for experts, the four ownership links for ownership, fit-score links for fit scores. No control is rendered disabled as a substitute for absence.
3. **No whole-record edit mode** — there is no panel-level "Edit" button, no edit-mode toggle, and no modal that edits name and description together. The edit dialog for applications is removed along with its call sites.
4. **Focused dialogs only for compound input** — a dialog may open from a field only when the input needs a picker or several values at once: owner (user or team lookup) and add expert (name, role, contact). A single value never opens a dialog.
5. **Per-field commit** — each in-place field commits on confirm (Enter or the field's confirm control) and discards on cancel (Escape or the field's cancel control). A failed request keeps the field in edit mode and shows the error; a rejected value never reaches the API.
6. **Same validation as creation** — name and description use the same schema as the create application form.
7. **View membership is a separate section** — custom colour and "Remove from view" are properties of the view placement, not the application. They render in one "In this view" section after the application's own fields and before the one-pager action and history, only when the panel is hosted on the canvas and the application is on the current view, gated by the view-component's own links.
8. **Freshness on every host** — all application reads go through TanStack Query so the existing mutation effects (list, detail, statistics, one-pager, one-pager quality) refresh every open host. No host may fetch the application outside the query layer.
9. **One heading** — the panel renders the application name as its heading; a host adds a container title at most, never a second "Application Details" heading.
10. **Tree Edit selects** — the navigation tree's context-menu "Edit" selects the application so the details pane shows it; inline rename in the tree is unchanged.
11. **Creation is unchanged** — the create application dialog stays; this spec covers editing an existing record only.

---

## Acceptance Criteria

- [x] The canvas details pane, the Business Domains drawer and the one-pager subject drawer render the same panel component with the same sections in the same order; a test per host asserts the experts section is present.
- [x] Name and description are edited in place, gated on the `edit` link; confirm persists through the existing update request, cancel sends nothing, an empty name is rejected client-side with the message shown in the field.
- [x] With no `edit` link, name and description render as text with no edit control.
- [x] Empty description renders an "Add a description" prompt when `edit` is present, and nothing when it is absent.
- [x] The application edit dialog and every call site (canvas dialog manager, canvas dialog hook, navigation provider, Business Domains drawer, one-pager panel) are removed; the tree context-menu "Edit" selects the application.
- [x] Custom colour and "Remove from view" render in one "In this view" section only on the canvas when the application is on the current view.
- [x] The Business Domains drawer reads the application through a query hook; the hand-rolled fetch hook is removed; a test shows a hosting change made in the drawer re-renders from the invalidated detail query.
- [x] Exactly one experts list is mounted per open panel.
- [x] Each host renders one heading for the application.
- [x] E2E (mock-mode Playwright project, `e2e/mock/application-details.spec.ts`): rename in place on the canvas with the tree following, and describe in place from a domain board chip with experts visible in the drawer. The real-backend project needs Docker, which the implementation environment lacked; the scenario is covered by the mock project until it can run there.

---

## Architecture

### Ownership

Frontend only, in the components feature. Architecture Modeling's contract is unchanged: the same `edit` link and update request carry name and description; a single-field commit sends the other field unchanged. Business Domains and OnePagers keep hosting the panel but stop owning any of its behaviour.

### Domain Model

No change.

### API Surface

No change. Every affordance in this spec is already a link on the application or view-component resource.

### Persistence

No change.

### Frontend

- One panel component takes an application id, resolves the application through the query layer, and renders the sections in rule-1 order. It owns the owner dialog and the add-expert dialog.
- An in-place text field primitive (read text, pencil affordance, input with confirm and cancel, error state) serves name and description; the navigation tree's rename may adopt it but is not required to.
- The view-membership section is passed in by the canvas host only, as the one host-specific slot; the panel renders it after the application's own fields, before the one-pager action and history.
- Hosts become thin: the canvas details pane, the Business Domains drawer and the one-pager subject drawer each render the panel inside their container.
- The application edit dialog, its dialog-manager entry, the canvas hook that opens it and the navigation provider wiring are deleted; the tree's Edit handler dispatches node selection instead.
- The hand-rolled fetch hook in Business Domains is replaced by the components detail query hook.

### Cross-Context Integration

None. Spec 191's freshness rule (subject mutations invalidate the one-pager queries) continues to hold through the existing mutation effects.

---

## Implementation Notes

- The unused navigation context module under `frontend/src/contexts/navigation/` was the only other consumer of the retired dialog action; it had no importers anywhere and was removed.
- The mock API's domain realizations endpoint now serves the seeded realizations grouped by capability, and the mock seed carries one application, so the Business Domains drawer can be exercised in mock mode.

---

## Design Decisions

1. **Edit in place, not a dialog** — the panel already edits ownership, hosting, experts and fit scores where they are shown; extending that to name and description gives one rule and removes the only form that could go stale against the panel. Alternatives considered: _put every field behind the Edit dialog_ (rejected: ownership is a state machine with four distinct affordances that do not fit a form, and 214/215 already shipped inline); _a panel-level edit-mode toggle like the one-pager page_ (rejected: the one-pager toggles because its custom facts save as one batch; application fields commit individually, and a mode that flips some fields while others stay always-live is the confusion this spec removes).
2. **Hosts cannot hide sections** — the current body hides experts when a callback is omitted, which is how Business Domains lost them. Making the panel own its dialogs removes the optional-callback surface entirely. Alternative: _keep callbacks and fix the one missing call site_ (rejected: the same bug recurs at the next host).
3. **View membership as the one host slot** — colour and view removal belong to the view placement, so they are legitimately canvas-only. Naming the section makes the exception visible rather than accidental.
4. **Tree Edit selects instead of being removed** — the item is gated on `edit` and users know it; pointing it at the details pane keeps the affordance while honouring rule 3.
5. **Applications only** — capabilities have the same split (drawer versus canvas pane, Edit dialog for name and description). They get their own spec once this pattern has shipped, so one slice can be validated before the second follows.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Per-field commit for name and description | Two requests when both change, instead of one | Both fields are small; the update request is idempotent on the unchanged field |
| Removing the edit dialog | Users who learned "Edit" lose a familiar button | Pencil affordance on hover and the empty-description prompt make the in-place controls discoverable; the tree keeps its Edit item |
| Panel owns data fetching | Hosts can no longer pass a preloaded object | The list query is already warm on every host, so the detail read is a cache hit |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated (no API change)
- [x] User sign-off
