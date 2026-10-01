# 230 — Attention Items

> **Status:** pending
> **Depends on:** 227 (completeness published), 228 (home, caches, scope), 229 (journey cache, quarter rule), 190 (invite-to-edit from a findings row)
> **Roadmap alignment:** `SD8 / H2-6` — slice D of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) (decisions D5, D7, D8); "orphaned, stale and incomplete items ranked, routed via invite-to-edit"

---

## Problem Statement

Stewardship (spec 226) records who answers for each concern in each domain, and Home (spec 228) shows each user their part of the landscape. Neither tells anyone **what in it is wrong**. Data-quality problems are scattered: unowned applications are counted tenant-wide only, stale TIME grades show as a badge inside one capability's drawer, incomplete one-pagers sit on a separate list, capabilities without a realising application are not computed anywhere. A steward accountable for "assessment" in a domain has no way to find the ungraded realisations they answer for.

This slice computes **attention items**: findings of fixed checks, one concern each, about one subject each, ranked, shown to the users whose scope covers them — the steward of the (domain, concern), the domain architect, and the user anchored to the subject — and routed to the person who can fix them through invite-to-edit. Items are never stored, acknowledged or dismissed; fixing the data is the only way one disappears (spec 184's framing, design doc D5).

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Check** | A fixed rule in code that yields attention items of one concern. Each has a key, a concern, a severity, a subject kind and a label. |
| **Attention item** | One finding of one check about one subject. Identified by (check, subject). |
| **Subject** | What the finding is about: an application, a capability, a realisation (capability × application), a journey or a stewardship. |
| **Item domains** | The domains an item belongs to: a capability's effective domains; an application's application domains; a realisation's or journey's capability's effective domains; a stewardship's domain. May be empty. |
| **Severity** | `high`, `medium` or `low`, fixed per check. |
| **Visible to** | The users who see an item without asking for a domain (rule 3). |
| **Fixer** | The user an item is routed to (rule 6); null when nobody qualifies. |
| **Application health** | An application is healthy when no attention item has it, or one of its realisations, as subject. Derived, never stored (design doc D8). |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Steward** | See the findings of the concern they answer for, in their domains, most severe first, and send each to whoever should fix it. |
| **Domain Architect** | See every finding in their domains, whatever the concern. |
| **Owner / EA owner** | See what is wrong with their own applications and capabilities. |
| **Fixer** | Receive edit access to the subject from the steward, pre-addressed. |
| **Any reader of a domain** | See one domain's full list of findings. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Attention items

  Background:
    Given "Customer Engagement" has the domain architect "Alice Smith"
    And "Mette Gram" stewards "assessment" in "Customer Engagement"

  Scenario: A steward sees their concern's findings
    Given 6 realisations in "Customer Engagement" have no TIME grade
    And 3 applications in "Customer Engagement" have no owner
    When "Mette Gram" opens Home
    Then Needs Attention shows "6 realisations not assessed"
    And does not show the unowned applications

  Scenario: The domain architect sees every concern
    When "Alice Smith" opens Home
    Then Needs Attention shows "3 applications without an owner" and "6 realisations not assessed"

  Scenario: An owner sees findings about their own application
    Given "Ole Berg" owns "CRM", whose one-pager misses 2 required fields
    And holds no stewardship
    When "Ole Berg" opens Home
    Then Needs Attention shows "1 one-pager incomplete"

  Scenario: Home groups by check, most severe first
    Given the caller sees 2 high, 5 medium and 9 low items across 4 checks
    When they open Home
    Then Needs Attention lists one row per check with its count, high checks first, then by count
    And shows at most 5 rows and the total number of items

  Scenario: View all
    When the caller selects "View all"
    Then the attention list opens with every item they see, ranked
    When they select a check row on Home instead
    Then the attention list opens filtered to that check

  Scenario: The list is ranked and paginated
    Given the caller sees 120 items
    When they open the attention list
    Then items are ordered by severity, then first domain name, then subject name, then check
    And the list pages 50 at a time

  Scenario: Filtering the list
    When the caller filters the attention list by concern "documentation" and domain "Customer Engagement"
    Then only documentation items belonging to "Customer Engagement" are shown

  Scenario: One domain for any reader
    Given the stakeholder "Eva Dahl" holds no anchor
    When they open the attention list for the domain "Customer Engagement"
    Then every item belonging to that domain is shown, whatever the concern

  Scenario: An item names its fixer
    Given the application "CRM" is owned by "Ole Berg" and its one-pager is incomplete
    Then the item names "Ole Berg" as fixer

  Scenario: The fixer falls back to the domain architect
    Given the application "Portal" is managed by the team "Web"
    And its one-pager is incomplete
    Then the item names "Alice Smith" as fixer

  Scenario: Routing a finding
    Given "Mette Gram" may grant edit access
    And a documentation item on "CRM" names "Ole Berg" as fixer
    When "Mette Gram" selects "Invite to edit" on the item
    Then the invite dialog opens for the application "CRM" with "Ole Berg"'s e-mail filled in

  Scenario: No routing without the right to grant
    Given the caller may not grant edit access
    Then no item offers "Invite to edit"

  Scenario: No routing to oneself
    Given an item's fixer is the caller
    Then the item does not offer "Invite to edit"

  Scenario: Fixing the data removes the item
    Given a realisation of "CRM" for "Invoicing" is not assessed
    When an architect records a TIME grade for it
    And the caller reloads the attention list
    Then the item is gone

  Scenario: Stale assessment
    Given a TIME grade recorded 13 months ago
    Then an "assessment stale" item exists for that realisation
    Given a TIME grade recorded 11 months ago
    Then no stale item exists for it

  Scenario: Eliminate without a journey
    Given "CRM" is graded Eliminate on "Invoicing" and "Invoicing" has no planned or in-flight journey
    Then an "Eliminate without journey" item exists for that realisation
    When a journey is planned for "Invoicing"
    Then the item is gone

  Scenario: Overdue journey
    Given the in-flight journey on "Invoicing" is overdue
    Then a "journey overdue" item exists for it

  Scenario: Overdue milestones
    Given the planned journey on "Dunning" targets a future quarter and has 2 overdue milestones, the first "Pilot"
    Then one "milestones overdue" item exists for that journey, naming 2 milestones and "Pilot"
    And no "journey overdue" item exists for it

  Scenario: Capability not realised
    Given the active leaf capability "Dunning" has no realising application
    Then a "capability not realised" item exists for it
    Given the planned leaf capability "Collections" has no realising application
    Then no item exists for it

  Scenario: Disabled steward
    Given "Mette Gram" is disabled while stewarding "assessment" in "Customer Engagement"
    Then a "steward disabled" item exists for that stewardship, visible to "Alice Smith"

  Scenario: Application health on My Work
    Given "Ole Berg" owns "CRM" with 2 attention items and "Portal" with none
    When "Ole Berg" opens Home
    Then the "CRM" card in My Work shows "2 issues" and the "Portal" card shows it is healthy

  Scenario: Nothing needs attention
    Given the caller sees no items
    When they open Home
    Then Needs Attention reads "Nothing needs your attention"

  Scenario: Opening a subject
    When the caller selects an item about a capability, a realisation or a journey
    Then the Business Domains board opens with the capability's drawer
    When they select an item about an application
    Then the application's one-pager opens
    When they select an item about a stewardship
    Then the Business Domains board opens on that domain
```

---

## Business Rules & Invariants

1. **The checks** — exactly these, fixed in code:

   | Key | Concern | Severity | Subject | Finding |
   |-----|---------|----------|---------|---------|
   | `application-unowned` | ownership | high | application | ownership state `unknown` |
   | `owner-disabled` | ownership | high | application | owned by a user who is disabled |
   | `ownership-unconfirmed` | ownership | medium | application | ownership state `nominated` |
   | `realisation-unassessed` | assessment | medium | realisation | no TIME grade |
   | `assessment-stale` | assessment | low | realisation | TIME grade recorded before the stale threshold |
   | `one-pager-incomplete` | documentation | low | capability, application | completeness `incomplete` (spec 227) |
   | `eliminate-without-journey` | planning | high | realisation | graded Eliminate, and its capability has no journey in motion (spec 184 rule 1) |
   | `journey-overdue` | planning | high | journey | journey in motion and overdue (spec 229) |
   | `milestone-overdue` | planning | medium | journey | journey in motion with at least one overdue milestone (spec 229); one item per journey |
   | `capability-unrealised` | structure | medium | capability | status Active, no child capability, no realisation |
   | `capability-without-ea-owner` | structure | low | capability | L1, and its EA owner is empty or not a user of the tenant |
   | `steward-disabled` | the stewardship's concern | high | stewardship | the steward is disabled |

2. **Stale threshold is one constant** — the 12-month threshold becomes a constant in Architecture Direction's published language, used by Architecture Direction's own stale flag and by this check.
3. **Visibility** — an item is visible to the caller when their scope kind is `tenant`; or when, for some item domain, the caller stewards (that domain, the item's concern) or architects that domain; or when the caller is anchored to the subject (EA owner or edit grantee of the capability; owner, nominated owner or edit grantee of the application; either of these for a realisation's capability or application; EA owner of a journey's capability). Scope kind `empty` sees nothing on Home.
4. **Domain view** — a caller holding `domains:read` who asks for one domain sees every item belonging to it, whatever their anchors.
5. **Ranking** — severity (high, medium, low), then the alphabetically first item domain's name (items without a domain last), then subject name, then check key, then subject id. Deterministic, keyset-paginated, 50 per page.
6. **Fixer** — the first active user found in this order: for application subjects and `one-pager-incomplete` on an application, the owner (owner kind `user`); for capability, realisation, journey and `one-pager-incomplete` on a capability, the capability's EA owner; then the domain architect of the alphabetically first item domain that has one. `steward-disabled`, `ownership-unconfirmed` and `application-unowned` go directly to the domain architect. Null when none qualifies.
7. **Routing** — an item carries `x-edit-grants` (and the fixer's e-mail) only when the caller may grant edit access (the shared `AddEditGrantsLink` rule), the fixer is a user other than the caller, and the item has a grantable artifact: the application (`component`) for application subjects, the capability for capability, realisation and journey subjects. Stewardship items are never routed; reassigning a steward is a `domains:write` act.
8. **Computed at read time, never stored, never dismissed** — no table holds items; no endpoint acknowledges, snoozes or hides one.
9. **Home summary** — one row per check among the caller's visible items: check key, concern, severity, label, count; ordered by severity, then count descending, then check key; at most 5 rows, plus the total item count. Omitted for `empty` scope.
10. **Application health** — each application card in My Work carries the number of attention items whose subject is that application or one of its realisations, counted regardless of who may see them, since the card describes the application; zero is healthy.
11. **Permissions per subject** — items about capabilities, realisations and journeys require `capabilities:read`; about applications `components:read`; assessment and planning checks also `architecture-direction:read`; a caller lacking one does not receive those items or their counts.
12. **Caches only** — every check reads the context's local caches (spec 209); completeness comes from a new cache fed by `SubjectCompletenessRecalculated`.

---

## Acceptance Criteria

- [ ] Each check has tests for a finding, a non-finding and its boundary (stale at 12 months exactly, leaf vs parent capability, Active vs Planned, L1 vs L2, legacy free-text EA owner, journey in motion vs done)
- [ ] The stale threshold is a published-language constant used by Architecture Direction's stale flag and by `assessment-stale`; Architecture Direction's existing stale tests pass unchanged
- [ ] Visibility has a test per path of rule 3, including a steward of another concern in the same domain seeing nothing and an architect seeing everything in their domain
- [ ] `GET /api/v1/attention-items` returns the caller's visible items ranked per rule 5 with keyset pagination, filters `concern`, `check`, `domainId`; with `domainId` and `domains:read` it returns the domain view of rule 4; an unknown domain answers 404
- [ ] Each item carries check, concern, severity, label, subject (kind, id, name, and capability and application for a realisation), context (missing count, assessed-at, grade, target period, overdue milestone count and first label), domains, fixer (id, name) or null, and links per rule 7
- [ ] `GET /api/v1/home` carries the attention summary of rule 9, `x-attention-items`, and application health on My Work application cards
- [ ] The completeness cache consumes `SubjectCompletenessRecalculated` and is seeded by a backfill migration from `onepagers.one_pager_subject_index`; subject deletion events remove its rows
- [ ] `InviteToEditButton` and its dialog accept a default grantee e-mail; the attention list pre-fills it from the item's fixer
- [ ] Home renders Needs Attention (rows, counts, severity marker, "View all", empty message) and a new attention list page at `/attention` with concern, check and domain filters, pagination, fixer, subject deep links and "Invite to edit"; both use Mantine primitives
- [ ] Subject deep links per the last scenario; stewardship items open `/business-domains?domain={id}`
- [ ] The attention list query key is invalidated by the edit-grant mutation effects so a routed item's affordance reflects the new grant
- [ ] Every BDD scenario has at least one corresponding test
- [ ] E2E: a steward opens Home, follows "View all", filters by concern and opens the invite dialog pre-filled
- [ ] `docs/architecture/Stewardship.md` lists the checks; `docs/backend/cross-context-events.md` lists the OnePagers subscription
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

The `stewardship` read side gains the checks, the attention query and a completeness cache. Architecture Direction gains one published-language constant (rule 2). Access Delegation's frontend `InviteToEditButton` gains a default-grantee prop. No other context changes.

### Domain Model

- **`Check`** — a value object: key, concern (spec 226's `Concern`), severity, subject kind, label. The table in rule 1 is its complete list; the check list and the concern list together are the context's vocabulary, both fixed in code (design doc D4).
- **Attention query** — per check, a query over the caches yielding (check, subject, context, item domains); a composer applies visibility (rule 3) or the domain view (rule 4), ranking (rule 5) and fixer resolution (rule 6).
- **Home summary** and **application health** — counts over the same query.

### API Surface

- `GET /api/v1/attention-items` — any authenticated user; query parameters `concern`, `check`, `domainId`, `cursor`, `limit`; response `PaginatedResponse` of items with `_links.self` and `next`; per-item `x-edit-grants` per rule 7. Invalid `concern` or `check` answers 400.
- `GET /api/v1/home` — gains `attention` (summary) and `_links.x-attention-items`; My Work application cards gain `issues`.
- Shapes, status codes and Swagger per `easi-api-standards`.

### Persistence

One new cache in the `stewardship` schema, tenant-scoped with RLS: **completeness cache** — subject type, subject id, completeness, missing count — fed by `SubjectCompletenessRecalculated` and the supplier deletion events already subscribed, seeded by a `backfill` migration from the subject index. Items themselves are not persisted.

### Frontend

- `home` feature: `NeedsAttention` section (grouped rows, "View all" gated on `x-attention-items`), `issues` badge on application cards.
- `attention` page at `/attention` (new route constant), filters bound to the URL query, cursor pagination (the One-Pager Quality page's pattern), rows with severity, label, subject link, domains, fixer, context and `InviteToEditButton`.
- `InviteToEditButton` / `InviteToEditDialog` gain an optional default grantee e-mail.

### Cross-Context Integration

| Direction | Events | Purpose |
|-----------|--------|---------|
| OnePagers → Stewardship | `SubjectCompletenessRecalculated` | Completeness cache (first consumer of spec 227) |
| Architecture Direction → Stewardship | stale threshold constant | One definition of stale |
| Stewardship → Access Delegation | none; the frontend posts to `/edit-grants` through the existing dialog | Routing |

---

## Design Decisions

1. **Checks and severities fixed in code** — a check is code; a severity changes what surfaces first and is part of the check's meaning. Alternative: tenant-configurable severities in MetaModel (rejected — invariant 5 governs the model's vocabulary, and a severity table nobody tunes is configuration without a user).
2. **Items are per (check, subject), not per (check, subject, domain)** — an application in two domains is one problem with one fix; both domains' stewards see the same item. Alternative: one item per domain (rejected — duplicates in the list of a user who stewards both).
3. **Realisation-level assessment and planning checks** — TIME is graded per realisation (spec 180), so the finding is on the pair; the item names both capability and application.
4. **Only leaf, Active capabilities can be unrealised** — a parent is realised through its children, and a Planned capability is not expected to be realised yet. Alternative: any capability without a realisation in its subtree (rejected — reports the same gap at every level of the branch).
5. **EA owner is checked on L1 only** — EA ownership is the top-level accountability for a capability area; requiring it on every L2–L4 would flood the list with one finding per capability. A legacy free-text EA owner is a finding, because nobody can be routed to it.
6. **Eliminate-without-journey uses spec 184's gap definition** — the same rule text, so when spec 184 lands the signal and the item agree on what a gap is. Alternative: wait for spec 184 and consume its signals (rejected — spec 184 is pending and its signals are computed, not published).
7. **Assessment and journey fixers are EA owners, not application owners** — recording a TIME grade and planning a journey are Architecture Direction acts, done by architects; the capability's EA owner is the architect who answers for it (design doc D7 refined per check).
8. **Edit grant expiry is not a check** — moved to My Work in spec 228 (its decision 7).
9. **Application health counts all items, not only visible ones** — the card describes the application's state, which should not depend on who looks.
10. **Stewardship items are never routed** — the fix is reassigning the steward, which requires `domains:write`, not an edit grant on an artifact.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Computed per request over all caches | Cost grows with tenant size and check count | Each check is one indexed query; the list is paginated; tenants hold thousands of rows, not millions |
| Fixed severities | A tenant cannot re-prioritise | Severity order is visible and changes by spec, with the reason recorded |
| Gap rule shared with spec 184 by text, not by code | Two implementations until spec 184 decides where its engine lives | The rule is one clause; both reference spec 184 rule 1 |
| Keyset over a computed ordering | The cursor encodes severity, domain name, subject name, check and id | The One-Pager Quality list's cursor shape; renames between pages can shift a row, as there |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
