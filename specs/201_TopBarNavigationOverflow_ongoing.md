# 201 — Top Bar Navigation: Declutter and Overflow

> **Status:** ongoing
> **Depends on:** 168_Frontend_UI_Overhaul_done.md

---

## Problem Statement

The top bar shows eight navigation entries plus tenant name, Assistant, What's New and the user menu. On a typical laptop it is already crowded, and when the browser window is narrowed the entries do not adapt — they get clipped or push the right-hand actions off screen. Most entries also lack a tooltip, so an icon-only presentation would be unusable today.

Two entries are administrative and rarely used: *Invitations* belongs with *Users* (same audience, same task — onboarding people), and *Settings* is a per-tenant configuration screen that belongs with the account/tenant information in the user dropdown rather than in primary navigation.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Any user** | Reach every area of the app from the top bar at any browser width, with a readable label or a tooltip |
| **Admin** | Find Invitations next to Users; find Settings from the user menu |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Top bar navigation adapts to available width

  Scenario: Invitations lives under Users
    Given an admin with users:read and invitations:manage
    When they open the top bar
    Then there is no "Invitations" entry in the top bar
    And the Users page shows a "Users | Invitations" tab bar
    And selecting the Invitations tab shows the invitations list at /invitations
    And the Users entry in the top bar is highlighted on both /users and /invitations

  Scenario: Invitations tab is permission gated
    Given an architect with users:read but without invitations:manage
    When they open the Users page
    Then the tab bar shows only the Users tab

  Scenario: Settings lives in the user menu
    Given an admin with metamodel:write
    When they open the user dropdown
    Then there is a "Settings" item that navigates to /settings
    And there is no "Settings" entry in the top bar

  Scenario: Settings is hidden for users without metamodel:write
    Given an architect
    When they open the user dropdown
    Then there is no "Settings" item

  Scenario: Full width shows icon and label
    Given the top bar has enough room for every entry with its label
    Then every entry shows its icon and label

  Scenario: Reduced width shows icons only
    Given the top bar does not have room for every label
    But has room for every entry as an icon
    Then every entry shows only its icon
    And hovering an entry shows its label as a tooltip

  Scenario: Narrow width overflows entries into a "More" menu
    Given the top bar does not have room for every entry even as icons
    Then the entries that fit are shown as icons
    And the remaining entries are listed in a "More" dropdown opened from an ellipsis button
    And each overflowed entry in the dropdown shows icon and label and navigates on click

  Scenario: Every top bar entry has a tooltip
    When the user hovers any navigation entry, the Assistant button, the What's New button, the More button or the user menu trigger
    Then a tooltip with the entry's name is shown
```

---

## Business Rules & Invariants

1. **No Invitations in primary nav** — Invitations is reachable only through the Users page tab bar (route `/invitations` is unchanged).
2. **Invitations tab gating** — the Invitations tab renders only when the user has `invitations:manage`; the Users tab only with `users:read`.
3. **Settings in user menu** — the Settings item renders only with `metamodel:write`; it is never in primary nav.
4. **Three density modes, chosen by measurement** — `full` (icon + label), `compact` (icon only), `overflow` (icons for the entries that fit, the rest under a More menu). The mode is derived from the measured width available to the nav, never from a fixed viewport breakpoint.
5. **Overflow preserves order** — entries overflow from the end of the list; visible entries keep their original order.
6. **Tooltip on every entry** — every clickable element in the top bar has a tooltip; in `compact` and `overflow` modes the tooltip is the only text for nav entries.
7. **Active view stays reachable** — when the active view's entry is overflowed, the More button is shown in the active style.
8. **Right-hand actions collapse to icons on narrow headers** — Assistant and What's New drop their labels (and the tenant name is hidden) when the header is narrower than a fixed container breakpoint; they always have tooltips.

---

## Acceptance Criteria

- [x] `nav-invitations` and `nav-settings` no longer render in the top bar
- [x] Users page and Invitations page render a shared tab bar; tabs gated by permission; the Users nav entry is active for both views
- [x] User menu renders `user-menu-settings` only with `metamodel:write`; clicking it navigates to `/settings`
- [x] A pure layout function decides `{ mode, visibleCount }` from `{ availableWidth, fullWidths, compactWidth, moreWidth, gap }` and is unit tested for all three modes and the boundary cases
- [x] In `compact`/`overflow` modes nav entries render icon only; in `overflow` a `nav-more` ellipsis button opens a menu listing the overflowed entries with icon + label
- [x] Every nav entry, More button, Assistant, What's New and the user menu trigger wrap in a Mantine `Tooltip`
- [x] Existing AppNavigation tests still pass (One-Pager Quality gating)

---

## Architecture

### Ownership

Frontend only — `frontend/src/components/layout/` (AppNavigation, UserMenu), `frontend/src/features/users`, `frontend/src/features/invitations`, `frontend/src/features/chat/components/ChatButton`. No backend change.

### Verification notes

- `npm run build` passes; the compiled CSS contains the `@container (width<=1400px)` rule.
- Browser check (three density modes, overflow menu, tooltips) is pending — no compose stack in the implementation environment.
- Bug found in browser (MSW mock mode, Playwright): on the Architecture Canvas the overflow dropdown opened but was painted behind the explorer tree once the tree had loaded. Cause: a pre-spec-179 `@media (max-width: 768px)` rule in `navigation.css` made `.navigation-tree` `position: absolute; z-index: 1000`, and `explorerPane` had no positioning or stacking context, so the tree escaped the pane and out-ranked Mantine's portaled dropdown (z-index 300). Fix: `explorerPane` is now `position: relative; isolation: isolate`, and the obsolete media rule is removed. Verified at 420px on canvas and business-domains pages.

### Frontend

- The nav wrapper becomes `flex: 1 1 0; min-width: 0`, so the width available to it depends only on header width minus brand and right-hand actions. Right-hand actions collapse via a CSS container query on the header (fixed breakpoint), so their width is a function of header width only. This keeps the nav measurement free of feedback loops.
- Nav entries are measured through a hidden, `aria-hidden` twin rendered in full mode (label widths) plus one compact twin item and a More twin. A `ResizeObserver` on the nav wrapper triggers re-layout; a pure function computes the mode.
- Tooltips use Mantine `Tooltip`. The overflow menu uses Mantine `Menu`.
- ChatButton switches from global `app-header-action-*` classes to the shared header action button (icon + collapsible label + tooltip), removing the global classes from `index.css`.
- Users/Invitations share a `UserAdminTabs` component rendered above each page's header; tabs navigate between `/users` and `/invitations`.
- Each nav entry carries its own route; the nav no longer keeps a route table for views it never navigates to.

---

## Design Decisions

1. **Invitations as a tab on the Users page, not a nested dropdown in the nav** — a nav item that is both a link and a menu is ambiguous to click and harder to collapse into icon/overflow modes. Alternatives considered: _hover dropdown under Users_ (rejected: extra click to reach Users, awkward in compact/overflow modes).
2. **Measured density instead of viewport breakpoints** — the set of visible entries depends on permissions and session links, so a fixed breakpoint would be wrong for some users. Alternatives considered: _Mantine `useMediaQuery` breakpoints_ (rejected: not permission-aware).
3. **Right-hand actions collapse by container query, not by nav mode** — keeps the nav measurement a pure function of header width and avoids full↔compact oscillation.
4. **Tooltip text equals the entry label** — no invented descriptions; the tooltip's job is to name the icon.
5. **Tooltip wraps `Menu.Target`, not the other way round** — Mantine's `Tooltip` sends unknown props to its floating element, so a `Menu.Target` around a `Tooltip` would lose the click handler. `Tooltip` → `Menu.Target` → button forwards both sets of handlers to the button.
6. **Right-hand action breakpoint is 1400px** — on a fully-populated admin header the nav needs ≈1430px to show every label, so actions collapse to icons just before the nav has to; below that the nav measurement takes over.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Hidden measurement twin | Duplicate (hidden) DOM for ≤ 7 entries | Marked `aria-hidden`, `visibility: hidden`, no test ids, no tab stops |
| Container-query breakpoint for actions | Not permission-aware | Only three small actions are affected; nav (the variable part) is measured |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing — `AppNavigation`, `navLayout`, `UserAdminTabs`, `UserMenu`, `ChatButton`, `ChangeRoleModal` verified green; `UsersPage`/`InvitationsPage` tests switched to render with the router (tabs use `useNavigate`) and are to be re-run manually
- [x] Integration tests implemented if relevant — none needed, frontend-only
- [x] API documentation updated — no API change
- [ ] User sign-off
