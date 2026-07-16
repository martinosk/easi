# 187 — Edit One-Pager Facts on the One-Pager Page

> **Status:** done
> **Depends on:** [176 — OnePagerFacts capture](176_OnePagerFacts_done.md) (the facts aggregate, per-type inputs, and save flow being relocated), [177 — OnePagerView](177_OnePagerView_done.md) (the page and composed read gaining the Edit action), [178 — One-Pager Completeness](178_OnePagerCompleteness_done.md) (the completeness block already carried by the composed response, re-rendered after save)

---

## Problem Statement

Spec 176 put One-Pager fact editing on the entity's own detail panel — a form embedded
below the entity's tags, relationships, and audit history. Spec 177 then built the
One-Pager itself: the clean, shareable, stakeholder-facing fact sheet architects present
and send links to. Today that page is read-only. An architect who spots a stale fact
while presenting the one-pager has to leave it, find the right detail panel, locate the
field inside a section that also holds unrelated entity data, fix it there, and switch
back to confirm it rendered — for a fact whose entire purpose is to be seen and edited in
the one-pager, not on the entity's working detail view.

This slice moves editing to where the one-pager already lives: the page itself. The six
detail panels keep only the navigation link to the one-pager; the one-pager page gains an
Edit action that turns its custom fields into an editable form in place, reusing the
per-type inputs and save flow spec 176 already proved. As a side effect, this also
removes an unconditional configuration+facts fetch pair the inline section triggered on
every detail-panel open — read-only viewers of a panel no longer pay for data they never
see.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Fix a stale or missing one-pager fact from the same page they are presenting, without detouring through the entity's detail panel |
| **Stakeholder / shared-link viewer** | See a clean, read-only fact sheet with no edit affordance they cannot use |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Edit One-Pager facts on the one-pager page

  Scenario Outline: Detail panel shows only the One-Pager link
    Given a "<subject type>" Subject whose detail response carries the one-pager link
    When I open the Subject's detail panel
    Then I see a "One-Pager" action
    And I do not see an inline One-Pager edit form on the panel

    Examples:
      | subject type          |
      | Capability            |
      | Enterprise Capability |
      | Application           |
      | Acquired Entity       |
      | Vendor                |
      | Internal Team         |

  Scenario: One-pager page shows an Edit action when the user may edit
    Given I hold the write permission of a Subject whose one-pager I am viewing
    When the one-pager loads
    Then I see the sheet in read mode
    And I see an "Edit" action

  Scenario: One-pager page shows no Edit action without write permission
    Given I lack the write permission of a Subject whose one-pager I am viewing
    When the one-pager loads
    Then I see the sheet in read mode
    And I do not see an "Edit" action

  Scenario: A shared read-only link never exposes Edit
    Given I open a one-pager through a shared URL without the subject's write permission
    When the one-pager loads
    Then I do not see an "Edit" action

  Scenario: Entering edit mode makes only custom fields editable
    Given a one-pager whose configuration interleaves built-in and custom fields
    When I click "Edit"
    Then every active custom field renders an editable input matching its field type
    And every built-in field still renders read-only in its original position

  Scenario: Saving several edited custom fields records every one and returns to read mode
    Given I am editing a one-pager with three custom fields
    When I change two of the three fields and click "Save"
    Then both changed fields are recorded
    And the unchanged field is not written
    And the page returns to read mode showing the updated values
    And the completeness summary reflects the updated values

  Scenario: Cancel discards changes
    Given I am editing a one-pager and have changed a custom field's value
    When I click "Cancel"
    Then the page returns to read mode
    And the field shows its previously saved value
    And no field value was recorded

  Scenario: Share remains available in read mode
    Given I am viewing a one-pager in read mode
    Then a "Share (copy URL)" action is available
    And it is replaced by Save and Cancel actions while I am editing
```

---

## Business Rules & Invariants

1. **No inline edit form on detail panels** — none of the six subject detail panels
   render a One-Pager edit form; each keeps only the "One-Pager" navigation action, gated
   on the subject's `x-one-pager` link exactly as spec 177 left it.
2. **Edit affordance travels with the composed read** — the composed one-pager response
   carries an `x-record` link precisely when the requesting actor holds the write
   permission of the subject, mirroring the same permission check and permission map the
   facts resource already uses (spec 176 rule 11).
3. **Edit action gated on that link, not on role inspection** — the page's Edit action
   renders only when `x-record` is present on the loaded one-pager view; absent for a
   viewer without the subject's write permission, including a shared read-only link.
4. **Only custom fields are editable on the one-pager** — built-in fields remain read-only
   in edit mode; they are edited on the entity itself, not on the one-pager. The composed
   response carries no write affordance for built-in field values.
5. **Interleaved order holds in edit mode** — the configuration's single interleaved
   display order (spec 177 rule 5) is unchanged by entering edit mode; only the
   presentation of custom-field rows changes, built-in row order and content do not move.
6. **Dirty-only save** — clicking Save records exactly the custom fields that changed
   (spec 176 rule 7a: no-op suppression) and clears any field emptied during the edit;
   untouched fields produce no write.
7. **Save returns to read mode with fresh data** — on save success the page shows the
   updated field values and an updated completeness summary without a manual reload;
   cache invalidation follows the existing facts-save effect (spec 176).
8. **Cancel is a pure discard** — cancelling an edit issues no write and returns to read
   mode showing the values as they were before editing began.
9. **Facts business rules are unchanged** — spec 176's aggregate, validation, and
   permission rules continue to hold unchanged; this spec only relocates the entry point
   to those same commands.

---

## Acceptance Criteria

- [x] All six subject detail panels (Capability, Enterprise Capability, Application,
      Acquired Entity, Vendor, Internal Team) render only the "One-Pager" action; no
      inline one-pager edit form remains on any panel.
- [x] `GET /api/v1/one-pagers/{subjectType}/{subjectID}` carries an `x-record` link
      exactly when the requesting actor holds the subject's write permission, and omits
      it otherwise.
- [x] The one-pager page renders an "Edit" action only when `x-record` is present on the
      loaded view; the action is absent for a viewer lacking the subject's write
      permission, including a shared read-only link.
- [x] Clicking Edit renders every active custom field as an editable input matching its
      field type, in its configured interleaved position; every built-in field remains
      read-only in its original position.
- [x] Editing several custom fields and clicking Save records exactly the dirty fields,
      writes nothing for untouched fields, returns the page to read mode, and shows the
      updated values and updated completeness summary without a page reload.
- [x] Clicking Cancel discards in-progress edits and returns to read mode showing the
      pre-edit values; no field value is recorded.
- [x] "Share (copy URL)" is available and functional in read mode, and is not shown while
      editing.
- [x] Every BDD scenario above has a corresponding automated test.

---

## Architecture

### Ownership

`onepagers` owns this change end to end: the additive `x-record` link on the composed
one-pager response, and the one-pager page's new edit mode. The six supplier-owned detail
panels (`capabilitymapping`, `enterprisearchitecture`, `architecturemodeling` frontend
features) are touched only by deleting the inline section they mounted — no backend
change in those contexts, no permission or domain change anywhere. This is a frontend
relocation plus one additive backend link; spec 176's aggregate, commands, and endpoints
are unchanged and fully reused.

### Domain Model

No new aggregates, entities, or events. The facts aggregate, its commands
(`RecordFieldValue`, `ClearFieldValue`), and the `OnePagerFacts` read model from spec 176
are unchanged and continue to be the only write path. The change is confined to the
composed read's link set and to which UI surface calls the existing write endpoints.

### API Surface

- `GET /api/v1/one-pagers/{subjectType}/{subjectID}` — additive change: `_links` gains
  `x-record`, computed with the same subject-write-permission check and permission map
  the facts resource already applies (spec 176 rule 11), so a caller with the subject's
  write permission sees the same `x-record` presence on both the view and the facts
  resource. No existing field, status code, or consumer contract changes.
- `GET /api/v1/one-pagers/{subjectType}/{subjectID}/facts`,
  `PUT/DELETE /api/v1/one-pagers/{subjectType}/{subjectID}/facts/{fieldID}` — unchanged;
  the page's edit flow calls these exact endpoints, previously called only from the six
  detail panels.
- No new endpoints, no permission model change.

### Persistence

None. No migration, no schema change, no new read-model column. The composed query gains
one additional link computation using data (actor, subject type, subject ID) it already
has in scope.

### Frontend

- **Six detail panels** (`VendorDetails.tsx`, `InternalTeamDetails.tsx`,
  `AcquiredEntityDetails.tsx`, `ComponentDetails.tsx`, `CapabilityDetails.tsx`,
  `EnterpriseCapabilityDetailPanel.tsx`) — remove the inline facts section; keep the
  "One-Pager" action unchanged. `CapabilityDrawer.tsx` already shows only the action and
  needs no change.
- **One-pager page** — gains an Edit/read toggle. The Edit action is gated on `x-record`
  from the already-loaded composed view, so a read-only or shared-link viewer pays no
  extra request. Entering edit mode fetches the subject's One-Pager Configuration and
  Facts on demand — the same two resources the removed inline section already fetched —
  because the compact composed view's custom-field shape lacks the definition metadata
  (required/active/options) and per-value write links the existing form and save flow
  need. Built-in rows keep rendering read-only in place; custom rows switch to the
  existing per-type inputs, sharing the dirty-tracking and submission logic spec 176
  proved, saved through the existing save flow (which already invalidates the one-pager
  and facts caches on success). Cancel discards the in-progress form without saving.
  While editing, Save/Cancel replace the Share action; Share returns once back in read
  mode.
- The now-unused inline section component and its tests are removed along with its
  export; the per-type inputs, form-building helpers, and save/query hooks it used are
  retained and reused by the page.

### Cross-Context Integration

None beyond the existing spec 176/177 wiring. No new events, no new consumed or produced
published-language contracts.

---

## Design Decisions

1. **Advertise `x-record` on the composed view instead of fetching facts just to gate the
   Edit button** — the permission check is already computed for the facts resource, so
   hoisting the same result onto the view is cheap and keeps every read-only or
   shared-link viewer at zero extra requests. Alternative: always fetch
   `useOnePagerFacts` on page load purely to read its `x-record` link (rejected — pays an
   extra query for every viewer, including the many who will never edit, contradicting
   the constant-query-count quality attribute spec 177 established for this page).
2. **Defer the configuration + facts fetch to the moment Edit is clicked, reusing the
   exact hooks the removed inline section used** — the composed view's flattened custom
   field shape is display-optimized and omits the definition metadata and per-value
   HATEOAS links the existing form and save flow require; re-fetching the two richer
   resources on demand is simpler and safer than growing the composed response to
   duplicate that shape for the rare editor. Alternative: extend the composed view's
   custom-field DTO with full definition and per-value write links so no extra fetch is
   ever needed (rejected — bloats a read-optimized response with write-path metadata most
   requests don't need, and duplicates the configuration/facts resources' existing
   shape).
3. **Preserve the interleaved field order in edit mode** — built-in rows stay read-only in
   place; only custom rows switch to editable inputs, extending rule 5/D10's single
   ordering guarantee into edit mode. Alternative: group all editable custom fields into
   one block below the read-only built-ins while editing (rejected — reintroduces the
   pre-177 two-section mental model interleaving was designed to remove, and the sheet
   would visibly reflow between read and edit modes).
4. **Reuse the existing per-type inputs and save flow verbatim** — this is a relocation of
   a proven edit surface, not a new one; spec 176 already covers every field type and
   validation case. Alternative: build a new page-level form (rejected — duplicates
   already-proven RHF/Zod machinery for no behavioral gain).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Edit fetches configuration + facts on demand | An editor pays a fetch beyond the composed view every time they open Edit | Same cost the inline section already paid on every panel open; now paid only by users who actually edit, and only once per edit session |
| Interleaved rows swap in place during edit | The row-rendering component must reconcile the page's row chrome with the reused per-type input's own label/help-text rendering | Left to implementation; both pieces already exist and only need composing, not redesigning |
| `x-record` duplicated across two resources (view and facts) | Two link computations must stay in permission-sync | Both derive from the same subject-write-permission map introduced in spec 176; no new permission concept |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
