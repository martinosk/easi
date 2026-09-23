# 225 — Origin Entity Details: Collapsible, Re-arrangeable Groups

> **Status:** done
> **Depends on:** 223 (capability details groups, details shell), 224 (application details groups), 221 (origin entity details edited in place)
> **Roadmap alignment:** outside roadmap: presentation only; the origin entity panel that 221 unified moves onto the details shell, no domain or contract change. Respects SD6's affordance-as-link principle.

---

## Problem Statement

Acquired entities, vendors and internal teams share one detail panel since 221. It is the last flat panel: type-specific fields, notes, created, type, the related applications and the view section run together on the Architecture Canvas pane and the one-pager subject drawer. Every other node on the canvas now opens a grouped panel, so an origin entity is the one node that still reads differently.

This spec puts the origin entity panel on the details shell that 223 introduced and 224 proved on a second panel.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architecture maintainer** | The same grouped panel on every canvas node; fold away what they never edit |
| **Read-only viewer** | The same grouped panel, foldable, with no edit controls |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Origin entity details arranged in collapsible, re-arrangeable groups

  Scenario Outline: Groups and their contents
    Given <type> "<name>" has its type-specific fields, notes and one related application
    When I open its details on the Architecture Canvas
    Then the name heads the panel above every group
    And I see these groups in this default order, each with a header I can click:
      | Group        | Contents                          |
      | Description  | <type-fields>, notes              |
      | Metadata     | created, type                     |
      | Applications | the related applications          |
    And the Applications header carries the count of related applications
    And the one-pager action and the history section follow the groups

    Examples:
      | type            | name          | type-fields                              |
      | acquired entity | Nordic Cargo  | acquisition date, integration status     |
      | vendor          | SAP           | implementation partner                   |
      | internal team   | Platform Team | department, contact person               |

  Scenario: The canvas adds the view group
    Given "SAP" is placed on the current view and the membership carries a remove link
    When I open its details on the Architecture Canvas
    Then an "In this view" group holding "Remove from view" follows Applications
    And without the membership or the link there is no view group

  Scenario: Empty Applications group says so
    Given "SAP" has no related application
    When I open its details
    Then the Applications group is present with an empty state and a zero count

  Scenario: Collapse, move and remember
    Given the default order
    When I collapse Metadata and activate "Move Applications up" for one vendor
    Then Applications sits above Metadata and Metadata is collapsed
    And opening any other origin entity shows the same arrangement

  Scenario: Editing still happens in place
    Given the entity resource carries an "edit" link
    When I rename the entity in the heading and edit a field inside Description
    Then both edits behave exactly as in 221
```

---

## Business Rules & Invariants

1. **Three fixed groups plus the canvas group** — Description, Metadata, Applications, and on the canvas only, In this view. The name heads the panel above the groups and never collapses or moves.
2. **Every section lands in one group** — the type-specific fields and notes in Description; created and type in Metadata; the related applications in Applications; remove-from-view in In this view. The one-pager action and history stay below the groups.
3. **Applications never vanishes** — the group renders an empty state and a zero count when there is no related application; its header shows the count.
4. **Titles come from the group** — the related-applications list drops the heading it used to render.
5. **One remembered arrangement for all three types** — acquired entities, vendors and internal teams share one key, because they share one group set; reconciled as 223 rule 7.
6. **Shell rules apply** — collapse, move, disabled edge moves and host-slot placement follow 223 rules 4 to 7 through the shared details shell.
7. **Field affordances unchanged** — every field keeps the in-place control and link gating that 221 settled.

---

## Acceptance Criteria

- [x] The origin entity panel renders the name above the groups and the three groups in the default order with the contents in rule 2, for each of the three types; the panel test asserts the grouping per type.
- [x] The Applications group shows the count in its header and an empty state when empty; tests cover both.
- [x] The canvas host renders In this view after Applications only when the entity is on the current view with a remove link; the panel test asserts the slot placement and the renderer passes the section only then.
- [x] Collapsing and moving is reflected on another origin entity after remount; a mock-mode Playwright scenario covers two entities on the canvas (`e2e/mock/origin-entity-details-groups.spec.ts`).
- [x] All 221 in-place editing tests still pass unchanged in behaviour.

---

## Architecture

### Ownership

Frontend only, in the origin-entities feature, on the shared details shell.

### Domain Model, API Surface, Persistence

No change. Browser-local arrangement under an origin-entity key shared by the three types.

### Frontend

- The origin entity content becomes a group builder feeding the details shell, with the name field as heading and the one-pager action plus history as footer.
- The canvas renderer resolves view membership through a hook and passes the section only when the membership carries a remove link, mirroring the capability and application hosts.
- The related-applications list loses its heading and gains an empty state; the count moves to the group title.

### Cross-Context Integration

None.

---

## Design Decisions

1. **One key for three types** — the three types render the same three groups, so a vendor arrangement should carry over to a team. Alternative: _a key per type_ (rejected: three arrangements for one shape).
2. **Count in the group title** — the list used to carry the count in its label; the group header is the natural place for it now. Alternative: _count inside the panel_ (rejected: hidden when collapsed, which is when the count is most useful).
3. **Type-specific fields under Description** — they describe the entity; "Details" would collide with the panel name.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Three groups | Little to fold on a short panel | Consistency across canvas nodes is the point; headers are cheap |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (mock-mode Playwright scenario; no backend change)
- [x] API documentation updated (no API change)
- [x] User sign-off
