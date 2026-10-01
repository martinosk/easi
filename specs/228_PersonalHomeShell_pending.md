# 228 — Personal Home Shell

> **Status:** pending
> **Depends on:** 226 (stewardship context, stewards and caches), 214 (application ownership), 200 (EA owner as user id)
> **Roadmap alignment:** `SD8 / H2-6` (stewardship read side), `H3-9` (portfolio counts on the home only; KPI definitions stay with H3-9) — slice C of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) (decisions D1, D9, D11)

---

## Problem Statement

EASI opens on the Architecture Canvas, which says nothing about what, of everything in the repository, belongs to the person who opened it. Since spec 226 a user can be a steward of a concern in a domain, but nothing shows it to them; an application owner, an EA owner of a capability or the holder of an edit grant has to know where to look.

This slice makes `/` a **Home** composed for the signed-in user: which parts of the landscape are theirs (**scope**), how big and how assessed that part is (**portfolio** tiles and the TIME distribution), and the capabilities and applications they personally answer for (**My Work**). Scope is computed from the user's anchors — stewardships, domain architect, EA ownership, application ownership and edit grants — never from their role; role decides only what an anchor-less user sees (design doc D1, D9). The canvas moves to `/canvas`.

The home's findings (*Needs Attention*, slice D) and journeys (*Journey Health*, slice E) are separate specs; this slice builds the caches, the scope resolution and the page they extend.

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Home** | The page at `/`, composed by the backend for the caller. |
| **Anchor** | One link between the caller and the landscape: a stewardship they hold, a domain they architect, a capability whose EA owner they are, an application they own or are nominated to own, an active edit grant they hold on a capability or application. |
| **Scope domains** | Domains in which the caller holds any stewardship or which they architect. |
| **Portfolio** | The capabilities and applications inside the caller's scope (rule 3), counted on the tiles. |
| **Effective domains** | The domains a capability belongs to: those its L1 ancestor (or itself, if L1) is assigned to. A capability may belong to several. |
| **Application domains** | The effective domains of every capability an application realises. |
| **Scope kind** | `personal` (the caller has at least one anchor), `tenant` (no anchor and the caller holds `domains:write`), `empty` (no anchor, no `domains:write`). |
| **My Work** | The capabilities and applications the caller personally answers for or may edit: EA-owned, owned, nominated, edit-granted. |
| **Dominant grade** | The TIME grade recorded most often across a subject's realisations; ties go to the later letter of Eliminate > Migrate > Tolerate > Invest. |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Steward** (any role) | Open EASI and see the domains they answer for and how that part of the landscape stands. |
| **Application owner / EA owner** (any role) | See the applications and capabilities that are theirs, with their TIME standing, one click from each. |
| **Admin / Architect without anchors** | See the whole tenant's portfolio rather than an empty page. |
| **Stakeholder without anchors** | Be told why the page is empty and how to get scope. |
| **Canvas user** | Reach the canvas and every existing canvas link as before. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Personal home

  Background:
    Given the domain "Customer Engagement" contains the L1 "Customer Management" with 12 capabilities in its subtree, realised by 9 applications
    And the domain "Finance" contains 30 capabilities realised by 20 applications

  Scenario: A steward's home covers their domains
    Given "Mette Gram" stewards "assessment" in "Customer Engagement" and holds no other anchor
    When they open Home
    Then the scope reads that they steward assessment in Customer Engagement
    And the Capabilities tile shows 12 with the count per status Active, Planned and Deprecated
    And the Applications tile shows 9
    And the Domains tile shows 1 and names "Customer Engagement"

  Scenario: The domain architect's home covers their domains
    Given "Alice Smith" is the domain architect of "Finance" and holds no other anchor
    When they open Home
    Then the Capabilities tile shows 30 and the Applications tile shows 20

  Scenario: An EA owner's capability is in their portfolio and My Work
    Given "Jonas Holm" is EA owner of "Invoicing" in "Finance" and holds no other anchor
    When they open Home
    Then the Capabilities tile shows 1
    And the Applications tile counts the applications realising "Invoicing"
    And My Work lists "Invoicing" with its level

  Scenario: An application owner's application is in their portfolio and My Work
    Given "Mette Gram" owns the application "CRM"
    When they open Home
    Then My Work lists "CRM" as owned by them
    And the Applications tile counts "CRM"

  Scenario: A nominated owner sees the application as nominated
    Given "Mette Gram" is nominated as owner of "Portal" and not yet confirmed
    When they open Home
    Then My Work lists "Portal" marked as nominated

  Scenario: A team-owned application is not the user's
    Given "Billing Engine" is managed by the team "Payments"
    When a member of that team who holds no anchor opens Home
    Then "Billing Engine" is not in My Work

  Scenario: An edit grant brings the artifact into scope
    Given "Ole Berg" holds an active edit grant on the application "CRM" expiring on 2026-10-21
    When they open Home
    Then My Work lists "CRM" with "Edit access until 21 Oct 2026"
    And the Applications tile counts "CRM"

  Scenario: An expired or revoked grant is not an anchor
    Given the only edit grant of "Ole Berg" expired yesterday
    When they open Home
    Then their scope is not personal

  Scenario: TIME distribution over the portfolio
    Given the portfolio's applications realise its capabilities through 40 realisations
    And 20 are graded Invest, 8 Tolerate, 4 Migrate, 2 Eliminate and 6 are not assessed
    When the user opens Home
    Then the TIME tile shows Invest 50%, Tolerate 20%, Migrate 10%, Eliminate 5% and Not assessed 15%

  Scenario: My Work shows the dominant grade
    Given "CRM" is graded Invest on two capabilities and Eliminate on one
    Then its My Work card shows "I"
    Given "Portal" is graded Invest on one capability and Migrate on one
    Then its My Work card shows "M"

  Scenario: My Work opens the subject
    When the user selects a capability card in My Work
    Then the Business Domains board opens with that capability's drawer
    When the user selects an application card
    Then the application's one-pager opens

  Scenario: Many items in My Work
    Given the user answers for 20 subjects
    When they open Home
    Then My Work shows 12 cards and states "20 in total"

  Scenario: An architect without anchors sees the tenant
    Given the architect "Per Lund" holds no anchor
    When they open Home
    Then the scope reads that no personal scope is set and the whole portfolio is shown
    And the tiles count every capability, application, domain and realisation of the tenant
    And My Work is empty

  Scenario: A stakeholder without anchors sees how to get scope
    Given the stakeholder "Eva Dahl" holds no anchor
    When they open Home
    Then no tiles are shown
    And the page explains that Home fills when they are made a steward, own an application or are granted edit access

  Scenario: Home follows changes to anchors
    Given "Mette Gram" stewards "assessment" in "Customer Engagement"
    When an architect assigns "Mette Gram" to "ownership" in "Finance"
    And they reload Home
    Then their scope and tiles cover both domains

  Scenario: Home is the landing page
    When a signed-in user opens EASI at "/"
    Then Home is shown
    When they select "Architecture Canvas" in the navigation
    Then the canvas opens at "/canvas"

  Scenario: Existing view links still open the canvas
    Given a view share link "/?view=v1" copied before this change
    When a user opens it
    Then the canvas opens at "/canvas?view=v1" with that view
    And a view share link copied after this change reads "/canvas?view=v1"

  Scenario: The logo returns to Home
    Given a user on any page
    When they select the EASI logo
    Then Home opens
```

---

## Business Rules & Invariants

1. **Anchors, exactly** — the caller's anchors are: live stewardships whose steward is the caller; domains whose domain architect is the caller; capabilities whose `eaOwner` equals the caller's user id (legacy free-text values, spec 200, match nobody); applications in ownership state `owned` or `nominated` with owner kind `user` and owner id the caller; edit grants whose grantee email equals the caller's email, whose artifact type is `capability` or `component`, and whose expiry is in the future. Nothing else is an anchor; role is never one.
2. **Scope kind** — `personal` when the caller has at least one anchor; otherwise `tenant` when the caller holds `domains:write`, otherwise `empty`.
3. **Portfolio** — for `personal`: capabilities = those whose effective domains intersect the scope domains, ∪ EA-owned capabilities, ∪ edit-granted capabilities; applications = those realising a portfolio capability, ∪ owned or nominated applications, ∪ edit-granted applications. For `tenant`: every capability and application. For `empty`: nothing.
4. **Concern does not narrow the portfolio** — holding any stewardship in a domain puts the whole domain in the portfolio; concerns narrow attention items only (slice D).
5. **Tiles** — Capabilities: total and count per status (Active, Planned, Deprecated). Applications: total. Domains: count and up to five names in alphabetical order with "and n more" beyond; for `personal` the scope domains plus the effective domains of EA-owned and edit-granted capabilities; for `tenant` every domain. TIME: over realisations whose capability or application is in the portfolio, the share per grade and of not assessed, of all such realisations, rounded to whole percent; an assessment counts only while its realisation exists; stale grades count as their grade.
6. **My Work** — EA-owned capabilities, owned and nominated applications, edit-granted capabilities and applications, each once (an owned application that is also granted appears as owned). Each card: subject type, id, name, capability level (capabilities), relation (`ea-owner`, `owner`, `nominated`, `edit-grant`), grant expiry (edit grants) and dominant grade (none when no realisation of the subject is assessed). Ordered by relation (ea-owner, owner, nominated, edit-grant), then name; the first 12 are returned with the total.
7. **Computed at read time** — tiles, scope and My Work are computed from the context's caches per request; nothing about a user's home is stored.
8. **Sections follow read permissions** — the backend omits the capability data from tiles and My Work unless the caller holds `capabilities:read`, application data unless `components:read`, and the TIME tile unless `architecture-direction:read`. The frontend renders the sections the response contains and never inspects the role.
9. **Tenant-scoped** — every cache row carries the tenant and is isolated by row-level security.
10. **Caches only** — every fact is read from a local, event-fed, backfilled cache in the `stewardship` schema; no request reads another context (spec 209).
11. **`/` is Home, `/canvas` is the canvas** — the canvas route renders the canvas; `/` with a `view` parameter redirects to `/canvas` with its query preserved; the view share link is generated on `/canvas`; the catch-all route leads to Home.

---

## Acceptance Criteria

- [ ] `GET /api/v1/home` returns scope (kind, stewardships with domain and concern, architected domains, counts of EA-owned capabilities, owned or nominated applications and edit grants), portfolio tiles and My Work per rules 1–8, with `self` link
- [ ] Anchor resolution has one test per anchor kind and per exclusion (legacy free-text EA owner, team ownership, expired grant, revoked grant, grant on a non-capability, non-component artifact)
- [ ] Portfolio counts are correct for a capability in several domains, an application realising capabilities in and out of scope, and an L2 whose L1 is reparented into another domain
- [ ] Scope kind is `tenant` for a caller with `domains:write` and no anchor, `empty` for one without; an `empty` response carries no tiles and no My Work
- [ ] TIME percentages include not assessed, ignore assessments of deleted realisations and sum to 100 ± rounding
- [ ] Dominant grade follows the tie rule, with a test per tie
- [ ] Cache projectors handle every subscribed event and their deletions; backfill migrations seed every cache from existing data; an integration test runs the backfills against a seeded tenant
- [ ] The user cache stores e-mail (backfilled), and edit grants join on it case-insensitively
- [ ] The home page renders scope, tiles and My Work with Mantine primitives, the empty-scope guidance for `empty`, and omits any section absent from the response
- [ ] Capability cards open `/business-domains?capability={id}`; application cards open `/one-pagers/application/{id}`
- [ ] `/` renders Home, `/canvas` renders the canvas, `/?view=v1` redirects to `/canvas?view=v1`, the navigation's canvas entry targets `/canvas`, the view share link is generated on `/canvas`, and the logo links to Home
- [ ] Every call site that navigates to `ROUTES.HOME` meaning the canvas is moved to `ROUTES.CANVAS`
- [ ] E2E: a steward opens `/` and sees their domain's tiles; the canvas opens from the navigation
- [ ] The dependency graph shows `stewardship` depending on `capabilitymapping`, `auth`, `architecturemodeling`, `architecturedirection` and `accessdelegation` only, and remains acyclic
- [ ] `docs/architecture/Stewardship.md`, `docs/backend/cross-context-events.md` and `docs/architecture/components.csv` reflect the new subscriptions and caches
- [ ] Every BDD scenario has at least one corresponding test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

The `stewardship` context gains its read side. It newly subscribes to Architecture Modeling, Architecture Direction and Access Delegation and to more Capability Mapping and Auth events, all through published languages. No supplier changes. The frontend gains a `home` feature and a routing change.

### Domain Model

No aggregate changes. The read side adds:

- **Scope resolver** — anchors → scope kind, scope domains, anchored subjects (rules 1–2).
- **Portfolio query** — scope → tiles (rules 3–5), effective domains derived at query time from the capability cache's parent chain and the L1 domain-assignment cache.
- **My Work query** — anchored subjects → cards with dominant grade (rule 6).

### API Surface

- `GET /api/v1/home` — the caller's home; any authenticated user. Response sections `scope`, `portfolio`, `myWork`, each omitted per rule 8 (`portfolio` and `myWork` also omitted for `empty`). `_links.self`. Slices D and E add sections and links to this resource.
- Shapes, status codes and Swagger per `easi-api-standards`.

### Persistence

New tables in the `stewardship` schema, each tenant-scoped with RLS and seeded by a `backfill` migration:

| Cache | Content | Fed by |
|-------|---------|--------|
| Capability cache | id, name, level, parent id, status, EA owner value | `CapabilityCreated`, `CapabilityUpdated`, `CapabilityMetadataUpdated`, `CapabilityParentChanged`, `CapabilityLevelChanged`, `CapabilityDeleted` |
| Domain assignment cache | (L1 capability id, domain id) | `CapabilityAssignedToDomain`, `CapabilityUnassignedFromDomain`, `CapabilityDeleted`, `BusinessDomainDeleted` |
| Application cache | id, name, ownership state, owner kind, owner id | `ApplicationComponentCreated/Updated/Deleted`, `ApplicationOwnerNominated`, `ApplicationOwnershipConfirmed`, `ApplicationOwnerAssigned`, `ApplicationOwnershipCleared` |
| Realisation cache | realisation id, capability id, component id | `SystemLinkedToCapability`, `SystemRealizationDeleted`, `CapabilityDeleted`, `ApplicationComponentDeleted` |
| TIME cache | (capability id, component id), grade, assessed at | `TimeAssessmentRecorded`, `TimeAssessmentRemoved` |
| Edit grant cache | grant id, artifact type, artifact id, grantee e-mail, expires at | `EditGrantActivated`, `EditGrantRevoked`, `EditGrantExpired` |
| User cache (existing) | + e-mail | `UserCreated` |

Expiry is evaluated against `expires_at` at read time: `EditGrantExpired` is not raised today, and Access Delegation itself enforces expiry on read.

### Frontend

- Feature `home` under `/frontend/src/features/`: API client, query key, `useHome`, `HomePage` with scope header, tile row, My Work grid and empty-scope guidance, built from Mantine primitives per `easi-frontend-styling`.
- Routing: `ROUTES.HOME` renders Home; `ROUTES.CANVAS` renders the canvas instead of redirecting; a `view` parameter on `/` redirects to `/canvas`; `generateViewShareUrl` targets `/canvas`; the canvas navigation entry targets `ROUTES.CANVAS`; the brand logo links to `ROUTES.HOME`; `AppView` gains `home`.
- A capability deep-link generator (`/business-domains?capability={id}`) joins the existing generators.
- Home's query is invalidated by no mutation: it is fetched on mount and on window focus, since its inputs change in other contexts' pages.

### Cross-Context Integration

| Direction | Events | Purpose |
|-----------|--------|---------|
| Capability Mapping → Stewardship | capability lifecycle, metadata, parent and level events; domain assignment events; realisation events | Capability, domain assignment and realisation caches |
| Architecture Modeling → Stewardship | application lifecycle and ownership events | Application cache |
| Architecture Direction → Stewardship | `TimeAssessmentRecorded`, `TimeAssessmentRemoved` | TIME cache |
| Access Delegation → Stewardship | `EditGrantActivated`, `EditGrantRevoked`, `EditGrantExpired` | Edit grant cache |
| Auth → Stewardship | `UserCreated` (e-mail) | User cache e-mail |
| Stewardship → Stewardship | `StewardAssigned`, `StewardReleased` | Already projected into the stewardships read model (spec 226) |

---

## Design Decisions

1. **Scope from anchors; role only for the anchor-less** — design doc D1 and D9. `domains:write` is the tenant-fallback criterion because it is the permission that assigns stewards: whoever may distribute accountability may see the whole. Alternative: role check `admin|architect` (rejected — the backend works in permissions, and the two coincide today).
2. **A stewardship brings in the whole domain, whatever the concern** — the portfolio answers "what is my part of the landscape"; the concern answers "what am I accountable for fixing in it" and belongs to attention items. Alternative: per-concern portfolios (rejected — an assessment steward would see TIME counts but no application count for the same domain).
3. **EA ownership anchors the capability only, not its subtree** — `eaOwner` is set per capability; inheriting it would invent a rule spec 200 does not have. Alternative: subtree inheritance (rejected — a new domain rule belongs in Capability Mapping, not in a read side).
4. **Applications belong to domains through the capabilities they realise** — there is no direct application–domain link. An application realising capabilities in two domains is in both. Alternative: a primary domain per application (rejected — no such fact exists).
5. **Nominated owners see the application** — a nomination is the user's claim awaiting confirmation; hiding it until confirmed hides work they have taken on. It is marked so the difference is visible.
6. **Edit grants anchor capabilities and applications only** — the portfolio counts capabilities and applications; a grant on a view, domain, vendor, team or acquired entity has nothing to count. Alternative: grants on domains bring the domain into scope (rejected — permission to edit is not accountability for a domain's landscape). The existing *My edit access* page still lists every grant.
7. **Grant expiry replaces the design doc's "edit grant expiring" attention item** — expiry is personal, not a concern of a domain, so it is shown on the My Work card rather than raised as a finding for a steward.
8. **One cache per supplier fact, all in `stewardship`** — the context's read side is one language (design doc D3); the Architecture Direction realisation cache is the precedent. Alternative: reuse suppliers' read models over HTTP (rejected — spec 209).
9. **Not assessed is part of the TIME distribution** — a bar of only graded realisations hides how much is ungraded, which is the assessment concern's whole point. Alternative: the mockup's four-grade bar (rejected — overstates assessment coverage).
10. **Old `/?view=` links redirect** — share links were generated on `/` since spec 113; breaking them silently would strand every link in chat and documents.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Five new supplier subscriptions | The context's dependency fan-in grows to five published languages | All are leaves toward `stewardship`; nothing imports it, so no cycle is possible |
| Effective domains derived at query time | A recursive walk per request | Capability trees are shallow (four levels); a recursive CTE over an indexed parent column |
| Six caches mirror supplier facts | Storage duplication and backfill cost | The spec 209 pattern; every cache is small and keyed by the supplier's id |
| No mutation invalidates Home | Home can be stale within a session until refocus | Home's inputs are edited on other pages; refetch on mount and focus covers returning to it |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
