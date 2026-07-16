# 191 — One-Pager: Open Subject Drawer

> **Status:** ongoing
> **Depends on:** 177 (one-pager view), 187 (edit one-pager on page), 188 (relations as built-ins)

---

## Problem Statement

Built-in fields (description, experts, maturity, …) are deliberately read-only on the one-pager: per 187 rule 4 they are edited on the entity itself. But there is no path from the one-pager to that entity. Subjects have no deep-linkable route — their detail panels only open through transient client-side selection (canvas node clicks, board selection) — so a user who spots a missing required built-in on the one-pager (often arriving from the Quality list, 189) must abandon the page and hunt for the subject in some view that happens to contain it.

The composed one-pager response already carries an unconditional `x-subject` navigation link that the frontend ignores. This spec turns that link into a working affordance: an "Open capability/application/…" button on the one-pager header that opens the subject's existing detail panel in a drawer on the page, where description, experts, and the other built-ins can be edited through the affordances that already exist there.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | Fix a missing description or expert the moment the one-pager reveals it, without hunting for the subject in a view |
| **Invited editor (190 grantee)** | Edit the subject they were invited to complete, starting from the one-pager they were pointed at |
| **Read-only viewer** | Inspect the subject's details from the one-pager without being offered edit actions they cannot perform |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Open the subject's detail drawer from its one-pager

  Scenario: Open the subject drawer
    Given I am viewing the one-pager of capability "Order Management"
    And the composed one-pager response carries an "x-subject" link
    When I click the "Open capability" button in the header
    Then a drawer opens on the one-pager page showing the capability's detail panel

  Scenario: Button label follows the subject type
    Given I am viewing the one-pager of an application component
    Then the header button is labeled "Open application"
    And each of the six subject types labels the button with its own type label

  Scenario: Edit affordances inside the drawer self-gate
    Given I hold the subject's write permission or an active edit grant for it
    When I open the subject drawer
    Then the detail panel shows its edit affordances (Edit, Add Expert, …) exactly as it does elsewhere in the app

  Scenario: Read-only viewer sees a read-only drawer
    Given I can view the one-pager but the subject's own response carries no edit links
    When I open the subject drawer
    Then the detail panel renders read-only, with no edit affordances

  Scenario: Edits made in the drawer refresh the one-pager
    Given the one-pager shows "Description" as missing
    When I edit the capability's description from the drawer and save
    Then the one-pager shows the new description and updated completeness without a full page reload
```

---

## Business Rules & Invariants

1. **Affordance driven by `x-subject`** — the header button renders exactly when the composed one-pager response carries the `x-subject` link. No other condition; no client-side permission inspection.
2. **No permission logic in onepagers** — the onepagers context emits no new edit-gated link and gains no new subject-type→permission map. Every edit affordance inside the drawer comes from the subject resource's own HATEOAS links (`edit`, `x-add-expert`, `x-remove`, …), evaluated by the owning bounded context.
3. **The drawer hosts the existing detail surface** — the same detail panel component the subject uses elsewhere (per subject type), with behavior parity: the same edit dialogs, experts management, and link gating.
4. **One-pager freshness after subject edits** — subject mutations reachable from the drawer invalidate the composed one-pager query and the one-pager-quality list caches, so built-in values and completeness refresh on the page behind the drawer.
5. **Read-mode affordance** — the button belongs to the one-pager's read-mode header actions; it is not offered while facts edit mode is active.

---

## Acceptance Criteria

- [x] The one-pager read-mode header shows an "Open {subject type label}" button for all six subject types when `x-subject` is present, and hides it when absent.
- [x] Clicking the button opens a drawer on the one-pager page hosting the subject's existing detail panel for that subject type.
- [x] With subject write permission (or an edit grant), the drawer exposes the panel's usual edit affordances; without them, the panel is read-only — verified by link-presence-driven tests, not role checks.
- [x] Editing a built-in value (e.g. description, experts) from the drawer updates the one-pager's field values and completeness indicator via cache invalidation, without a manual reload.
- [x] Subject mutations reachable from the drawer also invalidate the one-pager-quality list caches.
- [x] No new client-side permission logic is introduced; the button and drawer contents gate exclusively on HATEOAS link presence.

---

## Architecture

### Ownership

Frontend-only change owned by the one-pagers feature, which composes (does not modify) the detail panel components owned by the subject features (capabilities, components, origin-entities, enterprise-architecture). The onepagers backend context is untouched.

### Domain Model

None. No new aggregates, events, or read models.

### API Surface

None. The existing unconditional `x-subject` link (GET, subject detail resource) is the sole contract this feature consumes; it remains a navigation rel. Edit capability is communicated by the subject resources' own links, unchanged.

### Persistence

None.

### Frontend

- New "Open {subject}" button in the one-pager header's read-mode actions, gated on `hasLink(view, 'x-subject')`, labeled via the existing subject-type label mapping.
- A drawer component in the one-pagers feature that resolves the subject type to its existing detail panel component (reference patterns: the Business Domains board's capability drawer for hosting a panel in a drawer; the canvas detail-content renderer for the per-type component map). Panels keep their own data fetching by subject id.
- Cache invalidation: the mutation effects of subject features whose mutations are reachable from the drawer (capability, component, acquired entity, vendor, internal team updates and expert add/remove) are extended to invalidate the composed one-pager query keys and the one-pager-quality list keys. Enterprise capabilities have no update/expert mutations; their drawer-reachable mutations are the direction captures hosted inside the panel, which drive the included-capabilities built-in — the direction composition effects invalidate the enterprise-capability one-pager and quality list keys instead.
- The by-id panel hosts (`CapabilityDetailsPanel`, `ComponentDetailsPanel`) live in their owning subject features and reuse the existing panel content components without view coupling: no current-view lookup, so view-scoped affordances (remove from view, color pickers) never render in the drawer.

### Cross-Context Integration

None server-side. Client-side, the one-pagers feature imports detail panel components and mutation-effect keys across feature boundaries — matching the established cross-feature invalidation pattern.

---

## Design Decisions

1. **In-page drawer over a deep-linkable subject page** — reuses the existing panels where the user already is; smallest coherent slice; no new routes. Alternative: a standalone `/subjects/:type/:id` page (rejected for this slice — materially larger, though it would also have given My Edit Access and 190 invitees a landing URL; it deserves its own spec if wanted).
2. **Always-visible "Open …" button gated only on `x-subject`; edit affordances self-gate inside** — keeps all edit-permission truth in the owning contexts, which already gate `edit` via write-permission-or-edit-grant. Alternative: a backend-gated `x-edit-subject` link on the composed response (rejected — duplicates the owning context's gating inside onepagers, which already maintains four separate subject-type→permission maps).
3. **Reuse the existing detail panel components** — one source of truth for subject editing UI and its HATEOAS gating. Alternative: a purpose-built mini-editor for built-ins on the one-pager (rejected — duplicates edit dialogs and expert management, and drifts from the panels).
4. **Built-ins stay read-only on the one-pager itself** — 187 rule 4 stands. Alternative: in-place built-in editing in one-pager edit mode (rejected — reverses a fresh, deliberate decision and requires per-field write affordances across six subject types).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Always-visible "Open …" button | Read-only viewers open a drawer they cannot edit in | Label says "Open", not "Edit"; the panel renders the same read-only view it does elsewhere |
| Reusing detail panels in a new host | Panels expect page-level wiring (selection state, edit callbacks) | Follow the Business Domains drawer precedent; adapt wiring per panel inside the drawer component |
| Drawer instead of a subject route | The app still has no deep link to a subject's edit surface | Out of this slice by decision 1; the one-pager URL remains the shareable entry point |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (none relevant — frontend-only; no API or persistence changes)
- [x] API documentation updated (no API surface changes)
- [ ] User sign-off
