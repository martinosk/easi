# 224 — Application Details: Collapsible, Re-arrangeable Groups

> **Status:** done
> **Depends on:** 223 (capability details groups, details shell), 219 (application details edited in place), 214 (application ownership), 215 (application hosting), 216 (application composition)
> **Roadmap alignment:** outside roadmap: presentation only; the application panel that 219 unified is regrouped on the details shell that 223 introduced, no domain or contract change. Respects SD6's affordance-as-link principle.

---

## Problem Statement

Spec 219 gave applications one detail panel on every surface, and 214 to 216 have since added ownership, hosting and composition to it. The panel is now a flat run of twelve sections: description, ownership, hosting, composition, experts, created, type, realised capabilities, origins, fit scores, the view section and the one-pager action. On the Business Domains application drawer and the Architecture Canvas pane the reader scrolls through all of it to reach the one section they came for.

Spec 223 settled how a grouped panel behaves for capabilities and extracted the details shell so a second panel is a group builder and a storage key. This spec puts the application panel on that shell.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Application steward** | Reach ownership and hosting without scrolling past everything else; fold away what they never edit |
| **Domain owner** | Open an application from a capability's realising list and see what it realises and how it fits first |
| **Read-only viewer** | The same grouped panel, foldable, with no edit controls |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Application details arranged in collapsible, re-arrangeable groups

  Scenario: Groups and their contents
    Given application "Phoenix" has a description, an owner, a hosting classification, parts, experts, two realised capabilities, one origin and strategic pillars are configured
    When I open its details on the Business Domains application drawer
    Then the application name heads the panel above every group
    And I see these groups in this default order, each with a header I can click:
      | Group                 | Contents                                              |
      | Description           | description, type                                     |
      | Ownership             | ownership, hosting                                    |
      | Composition           | part-of line, parts list, or "Standalone"             |
      | Metadata              | experts, created                                      |
      | Realises capabilities | the realised capabilities, direct then inherited      |
      | Origins               | the origin entities                                   |
      | Fit scores            | one row per strategic pillar                          |
    And the one-pager action and the history section follow the groups

  Scenario: The canvas shows the same groups
    Given "Phoenix" is placed on the current view
    When I open its details on the Architecture Canvas
    Then I see the same groups in the same order
    And an "In this view" group, holding the custom colour control and "Remove from view", follows Fit scores

  Scenario: Empty groups say so instead of disappearing
    Given "Phoenix" realises no capability, has no origin and no strategic pillar is configured
    When I open its details
    Then Realises capabilities, Origins and Fit scores are present with a short empty state each

  Scenario: Collapse, move and remember
    Given the default order
    When I collapse Metadata and activate "Move Fit scores up" on the drawer
    Then Fit scores sits above Origins and Metadata is collapsed
    And the Architecture Canvas pane shows the same arrangement for any application

  Scenario: Editing still happens in place
    Given the application resource carries an "edit" link
    When I rename the application in the heading and edit the description inside Description
    Then both edits behave exactly as in 219
```

---

## Business Rules & Invariants

1. **Seven fixed groups plus the canvas group** — Description, Ownership, Composition, Metadata, Realises capabilities, Origins, Fit scores, and on the canvas only, In this view. The name heads the panel above the groups and never collapses or moves.
2. **Every section lands in one group** — description and type in Description; ownership and hosting in Ownership; composition in Composition; experts and created in Metadata; the realised capabilities list, the origins list and the fit scores in their own groups; custom colour and remove-from-view in In this view. The one-pager action and history stay below the groups.
3. **Groups never vanish on data** — a group whose list is empty renders its empty state; it is not omitted. Only In this view depends on the host.
4. **Titles come from the group** — sections inside a group do not repeat the group title. The composition, origins, realised-capabilities and fit-score sections drop the headings they used to render.
5. **One remembered arrangement per entity type** — application groups are stored under their own key, separate from the capability key, shared by every surface, reconciled as 223 rule 7.
6. **Shell rules apply** — collapse, move, disabled edge moves, and host-slot placement follow 223 rules 4 to 7 through the shared details shell.
7. **Field affordances unchanged** — every field keeps the in-place control and link gating that 219, 214, 215 and 216 settled.

---

## Acceptance Criteria

- [x] The application panel renders the name above the groups and the seven groups in the default order with the contents in rule 2; the panel test asserts the grouping.
- [x] Realises capabilities, Origins and Fit scores render an empty state instead of disappearing; tests cover each.
- [x] The composition, origins, realised-capabilities and fit-score sections no longer render their own headings.
- [x] The canvas host renders In this view after Fit scores only when the application is on the current view with an actionable link; the canvas test asserts this.
- [x] Collapsing and moving on one surface is reflected on the other after remount; a mock-mode Playwright scenario covers the drawer and the canvas (`e2e/mock/application-details-groups.spec.ts`).
- [x] All 219, 214, 215 and 216 in-place editing tests still pass unchanged in behaviour.

---

## Architecture

### Ownership

Frontend only, in the applications feature, on the shared details shell.

### Domain Model, API Surface, Persistence

No change. Browser-local arrangement under an application-specific key.

### Frontend

- The application content becomes a group builder feeding the details shell, with the name field as heading and the one-pager action plus history as footer.
- The canvas host resolves view membership through a hook and passes the section only when it has actionable links, mirroring the capability host.
- The origins and fit-score sections lose their internal headings and gain empty states; the realised-capabilities list gains an empty state.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Groups stay put when empty** — an Origins group that appears only once data loads would shift the layout under the reader; an empty state is calmer and tells the reader the feature exists. Alternative: _omit empty groups as 223 does for host slots_ (rejected: host slots are structural, empty lists are data).
2. **Composition and Origins are separate groups** — they answer different questions (what it consists of, where it came from) and each has its own affordances. Alternative: _one "Lineage" group_ (rejected: invented word for two unrelated concerns).
3. **Type sits in Description** — it is a static classification with a reference link, read together with the description. Alternative: _Metadata_ (rejected: Metadata holds people and dates).
4. **Own storage key** — application and capability group sets differ, so one key would reconcile against the wrong id set.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Seven groups | More headers than the capability panel | Every one collapses and moves; users fold what they do not use |
| Empty states for lists | A little more vertical space on sparse applications | One line each |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (mock-mode Playwright scenario; no backend change)
- [x] API documentation updated (no API change)
- [x] User sign-off
