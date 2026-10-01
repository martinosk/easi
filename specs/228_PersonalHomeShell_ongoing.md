# 228 — Personal Home Shell

> **Status:** ongoing — implemented 2026-10-01, awaiting user sign-off
> **Depends on:** 226 (stewardship context, stewards and caches), 214 (application ownership), 200 (EA owner as user id)
> **Roadmap alignment:** `SD8 / H2-6` (stewardship read side) — slice C of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) (decisions D1, D3, D9, D11)

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
| **Anchor** | One link between the caller and the landscape: a current stewardship they hold, a domain they architect, a capability whose EA owner they are, an application they own or are nominated to own, an active edit grant they hold on a capability or application. |
| **Current stewardship** | A stewardship that has been assigned and not released. |
| **Anchored application** | An application the caller owns, is nominated to own, or holds an active edit grant on. |
| **Scope domains** | Domains in which the caller holds any current stewardship or which they architect. |
| **Portfolio** | The capabilities and applications inside the caller's scope (rule 3), counted on the tiles. |
| **Realisation** | A direct realisation (`SystemLinkedToCapability`) of a capability by an application. Inherited realisations are never counted. |
| **Effective domains** | The domains a capability belongs to: those its L1 ancestor (or itself, if L1) is assigned to. A capability may belong to several. |
| **Application domains** | The effective domains of every capability an application directly realises. |
| **Scope kind** | `personal` (the caller has at least one anchor), `tenant` (no anchor and the caller holds `domains:write`), `empty` (no anchor, no `domains:write`). |
| **My Work** | The capabilities and applications the caller personally answers for or may edit: EA-owned, owned, nominated, edit-granted. |
| **Dominant grade** | The TIME grade recorded most often across a subject's assessed realisations; a tie goes to the most severe grade: Eliminate, then Migrate, Tolerate, Invest. Not assessed realisations are ignored. |

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
    Given the domain "Customer Engagement" contains the L1 "Customer Management" and its 11 descendants, 12 capabilities in all, directly realised by 9 applications
    And the domain "Finance" contains 30 capabilities directly realised by 20 applications
    And no application realises capabilities in both domains, unless a scenario says otherwise

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
    And the Applications tile counts the applications directly realising "Invoicing"
    And the Domains tile shows 1 and names "Finance"
    And My Work lists "Invoicing" with its level

  Scenario: An application owner's application is in their portfolio and My Work
    Given "Mette Gram" owns the application "CRM" and holds no other anchor
    And "CRM" realises one capability in "Customer Engagement" and one in "Finance"
    When they open Home
    Then My Work lists "CRM" as owned by them
    And the Capabilities tile shows 0
    And the Applications tile shows 1
    And the Domains tile shows 2 and names "Customer Engagement" and "Finance"
    And the TIME tile counts both realisations of "CRM"

  Scenario: A steward's TIME tile stays inside their domain
    Given "Mette Gram" stewards "assessment" in "Customer Engagement" and holds no other anchor
    And "CRM" realises one capability in "Customer Engagement" and one in "Finance"
    When they open Home
    Then the TIME tile counts the realisation of the "Customer Engagement" capability
    And does not count the realisation of the "Finance" capability

  Scenario: A nominated owner sees the application as nominated
    Given "Mette Gram" is nominated as owner of "Portal" and not yet confirmed
    When they open Home
    Then My Work lists "Portal" marked as nominated

  Scenario: A team-owned application is no user's
    Given "Billing Engine" is managed by the team "Payments"
    When a user who holds no anchor opens Home
    Then "Billing Engine" is not in My Work

  Scenario: An edit grant brings the artifact into scope
    Given "Ole Berg" holds an active edit grant on the application "CRM" expiring on 2026-10-21
    When they open Home
    Then My Work lists "CRM" with "Edit access until 21 Oct 2026"
    And the Applications tile counts "CRM"

  Scenario: An expired grant is not an anchor
    Given the only edit grant of "Ole Berg" expired yesterday
    When they open Home
    Then their scope kind is not personal
    And My Work does not list the granted artifact

  Scenario: A revoked grant is not an anchor
    Given the only edit grant of "Ole Berg" was revoked
    When they open Home
    Then their scope kind is not personal
    And My Work does not list the granted artifact

  Scenario: A grant on a deleted artifact is not an anchor
    Given "Ole Berg" holds an active edit grant on the application "Legacy Portal"
    And "Legacy Portal" has been deleted
    When they open Home
    Then "Legacy Portal" is not in My Work and is not counted

  Scenario: TIME distribution over the portfolio
    Given the portfolio's applications realise its capabilities through 40 realisations
    And 20 are graded Invest, 8 Tolerate, 4 Migrate, 2 Eliminate and 6 are not assessed
    When the user opens Home
    Then the TIME tile shows Invest 50%, Tolerate 20%, Migrate 10%, Eliminate 5% and Not assessed 15%

  Scenario: TIME percentages always sum to 100
    Given the portfolio has 3 realisations graded Invest, Tolerate and Migrate
    When the user opens Home
    Then the TIME tile shows percentages that sum to exactly 100

  Scenario: A portfolio without realisations
    Given the caller's portfolio contains capabilities but no realisation
    When they open Home
    Then the TIME tile states that there are no realisations to assess

  Scenario: My Work shows the dominant grade
    Given "CRM" is graded Invest on two capabilities and Eliminate on one
    Then its My Work card shows "I"
    Given "Portal" is graded Invest on one capability and Migrate on one
    Then its My Work card shows "M"
    Given "Billing" has no assessed realisation
    Then its My Work card shows no grade

  Scenario: My Work opens the subject
    When the user selects a capability card in My Work
    Then the capability's one-pager opens
    When the user selects an application card
    Then the application's one-pager opens

  Scenario: Many items in My Work
    Given the user answers for 20 subjects
    When they open Home
    Then My Work shows 12 cards and a "Show all 20" control
    When they select "Show all 20"
    Then all 20 cards are shown

  Scenario: Sections follow read permissions
    Given a caller with a personal scope who lacks "architecture-direction:read"
    When they open Home
    Then no TIME tile is shown
    And no My Work card shows a grade

  Scenario: An architect without anchors sees the tenant
    Given the architect "Per Lund" holds no anchor
    When they open Home
    Then the scope reads that no personal scope is set and the whole portfolio is shown
    And the tiles count every capability, application, domain and realisation of the tenant
    And My Work states that nothing is assigned to them personally

  Scenario: An architect with an anchor sees only their scope
    Given the architect "Per Lund" is EA owner of "Invoicing" and holds no other anchor
    When they open Home
    Then their scope kind is personal
    And the tiles count only their portfolio

  Scenario: A stakeholder without anchors sees how to get scope
    Given the stakeholder "Eva Dahl" holds no anchor
    When they open Home
    Then no tiles are shown
    And the page explains that Home fills when they are made a steward, own an application, or are granted edit access to a capability or application

  Scenario: Home follows changes to anchors
    Given "Mette Gram" stewards "assessment" in "Customer Engagement"
    When an architect assigns "Mette Gram" to "ownership" in "Finance"
    And they reload Home
    Then their scope and tiles cover both domains

  Scenario: Releasing the last stewardship ends the personal scope
    Given "Mette Gram" stewards only "assessment" in "Customer Engagement" and holds no other anchor
    When an architect releases that stewardship
    And they reload Home
    Then their scope kind is not personal

  Scenario: Home is the landing page
    When a signed-in user opens EASI at "/"
    Then Home is shown and the navigation marks "Home" as active
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

1. **Anchors, exactly** — the caller's anchors are:
   - current stewardships whose steward is the caller;
   - domains whose domain architect is the caller;
   - capabilities whose `eaOwner` equals the caller's user id (legacy free-text values, spec 200, match nobody);
   - applications in ownership state `owned` or `nominated` with owner kind `user` and owner id the caller;
   - edit grants whose grantee e-mail equals the lower-cased e-mail of the caller's session (the comparison Access Delegation uses), whose artifact type is `capability` or `component`, whose expiry is in the future, and whose artifact still exists in the capability or application cache.

   Nothing else is an anchor; role is never one. An actor without an e-mail (agent tokens) has no grant anchors.
2. **Scope kind** — `personal` when the caller has at least one anchor; otherwise `tenant` when the caller holds `domains:write`, otherwise `empty`. A caller with `domains:write` and one anchor sees only their personal scope (design doc D1, D9).
3. **Portfolio** — for `personal`:
   - capabilities = those whose effective domains intersect the scope domains, ∪ EA-owned capabilities, ∪ edit-granted capabilities;
   - applications = those directly realising a portfolio capability, ∪ anchored applications.

   For `tenant`: every capability and application. For `empty`: nothing.
4. **Concern does not narrow the portfolio** — holding any stewardship in a domain puts the whole domain in the portfolio; concerns narrow attention items only (slice D).
5. **Tiles**
   - **Capabilities:** total and count per status (Active, Planned, Deprecated).
   - **Applications:** total.
   - **Domains:** count and up to five names in alphabetical order, with "and n more" beyond. For `personal`: the scope domains ∪ the effective domains of EA-owned and edit-granted capabilities ∪ the application domains of anchored applications. For `tenant`: every domain.
   - **TIME:** over every realisation whose capability is in the portfolio or whose application is an anchored application. Shows the share per grade and of not assessed, of all such realisations, in whole percent by the largest-remainder method so the shares sum to exactly 100. An assessment counts only while its realisation exists; stale grades count as their grade. With no such realisation the tile carries a total of 0 and no shares.
6. **My Work** — EA-owned capabilities, owned and nominated applications, edit-granted capabilities and applications, each once (an owned application that is also granted appears as owned).
   - Each card carries subject type, id, name, capability level (capabilities), relation (`ea-owner`, `owner`, `nominated`, `edit-grant`), grant expiry date (edit grants) and dominant grade (none when no realisation of the subject is assessed).
   - Ordered by relation (ea-owner, owner, nominated, edit-grant), then name.
   - Every item is returned with the total. The frontend shows the first 12 and a control to show the rest.
7. **Computed at read time** — tiles, scope and My Work are computed from the context's caches per request; nothing about a user's home is stored.
8. **Sections follow read permissions** — portfolio membership is computed regardless of the caller's permissions; permissions decide only what the response carries.
   - Without `capabilities:read`: no Capabilities tile and no capability cards.
   - Without `components:read`: no Applications tile and no application cards.
   - Without `architecture-direction:read`: no TIME tile and no dominant grade on any card.
   - Without `domains:read`: the Domains tile and the scope carry counts but no domain names.

   The response is `200` for every authenticated caller; permissions never yield `403` here. The frontend renders what the response contains and never inspects the role or permissions.
9. **The caller only** — `GET /api/v1/home` takes no user id, e-mail or other identity parameter. Every anchor is resolved from the request's actor (user id and session e-mail). Slices D and E keep this invariant.
10. **Tenant-scoped** — every cache row carries the tenant and is isolated by row-level security; every query, including the recursive effective-domains walk, filters on the tenant at every step.
11. **Caches only** — every fact is read from a local, event-fed, backfilled cache in the `stewardship` schema; no request reads another context (spec 209).
12. **Read side apart from write side** — the new caches and the Home queries live in their own package. Stewardship's command handlers and reactors read only the domain and user caches, and never import that package.
13. **`/` is Home, `/canvas` is the canvas**
    - The canvas route renders the canvas.
    - `/` with a `view` parameter redirects to `/canvas` with its query and hash preserved. The redirect target is always the path `/canvas` plus the original query string and hash; no part of the target is taken from a query value.
    - Moving between Home and any other page keeps the application shell mounted; an open assistant conversation survives it.
    - A `view` link opened while the canvas is already initialised opens that view beside the open ones; an empty `view` value is removed from the address and ignored.
    - The view share link is generated on `/canvas`.
    - The catch-all route leads to Home.

---

## Acceptance Criteria

**Backend**
- [x] `GET /api/v1/home` returns the response shape below per rules 1–9, with `Cache-Control: private, no-store` and `Vary: Cookie`
- [x] Anchor resolution has one test per anchor kind and per exclusion: legacy free-text EA owner, team ownership, expired grant, grant expiring exactly now, revoked grant, grant on an artifact other than a capability or component, grant on a deleted capability or application, released stewardship, actor without e-mail
- [x] Grants match the session e-mail case-insensitively, tested with mixed-case data
- [x] Portfolio counts, Domains tile and TIME tile are correct for:
  - a capability in several domains;
  - an application realising capabilities in and out of scope (steward and application owner);
  - an L2 whose L1 is reparented into another domain;
  - an application owner with no other anchor;
  - a holder of only an edit grant on an application
- [x] Inherited realisations are not counted
- [x] The Domains tile names five domains alphabetically and "and 1 more" for six
- [x] Scope kind is `tenant` for a caller with `domains:write` and no anchor, `empty` for one without, and `personal` for a caller with `domains:write` and one anchor; an `empty` response carries no portfolio and no My Work
- [x] TIME shares include not assessed, ignore assessments of deleted realisations, count stale grades, and sum to exactly 100 by largest remainder (tested with 1/1/1); a portfolio without realisations yields a total of 0 and no shares
- [x] Dominant grade picks the most frequent grade, ignores not assessed realisations, and breaks every pairwise tie and a three-way tie toward the most severe grade
- [x] My Work dedupes an EA-owned capability that is also granted and an owned application that is also granted, excludes applications nominated to a team, orders by relation then name, and returns every item with the total
- [x] Each permission in rule 8 has a handler test asserting exactly what is omitted when the caller lacks it, including the dominant grade and the card links
- [x] Query parameters on `GET /api/v1/home` are ignored; one caller's request never returns another user's anchors
- [x] A tenant-isolation integration test: tenant B holds data matching tenant A's anchors, and tenant A's home excludes it
- [x] Cache projectors handle every subscribed event, including deletions and `EditGrantExpired`; capability cache status defaults to Active on `CapabilityCreated`
- [x] Backfill migrations seed every cache from existing data: realisations with `origin = 'Direct'` only, edit grants with `status = 'active'` only, `tenant_id` copied from the source row
- [x] An integration test runs the backfills against a seeded tenant, including an inherited realisation, and asserts the same cache rows the projectors produce for the same data
- [x] Command handlers and reactors do not import the Home read-side package

**Frontend**
- [x] The home page renders scope, tiles and My Work with Mantine primitives, a loader while loading, an error alert with retry when no home could be loaded (a failed background refetch keeps the rendered home), the empty-scope guidance for `empty`, the "nothing assigned to you personally" state for an empty My Work, and omits any section absent from the response
- [x] Cards are links, present only when the card carries its `x-one-pager` link: capability cards open `/one-pagers/capability/{id}`, application cards open `/one-pagers/application/{id}`
- [x] My Work shows 12 cards and a "Show all n" control when more are returned
- [x] Grades are shown as the existing TIME badges with the full grade name as accessible label; the TIME bar has a visible percentage legend
- [x] Grant expiry renders as `21 Oct 2026` from the backend's date
- [x] `useHome` sets `staleTime: 0`, `refetchOnMount: 'always'` and `refetchOnWindowFocus: true`
- [x] Canvas initialisation (`useAppInitialization`) runs only when the canvas view is shown; landing on Home waits on no views query, shows no "Data loaded" toast and creates no default view
- [x] The route table is extracted from `main.tsx` into a component that tests can render. Router-level tests cover `/` → Home, `/canvas` → canvas, `/?view=v1&x=1` → `/canvas?view=v1&x=1`, `/?view=//evil.example` → `/canvas?view=//evil.example` (stays on `/canvas`, same origin), and an unknown path → Home
- [x] The navigation gains a "Home" entry, active on `/`; the canvas entry targets `/canvas`; the logo links to Home with the accessible name "EASI home"
- [x] Deep-link parameters are read only on the routes they are registered for; `view` is registered for `/canvas` only, and the view share link is generated on `/canvas`
- [x] Without domain names, stewardships of one concern in several domains read as one scope statement with the number of domains
- [x] A router-level test proves the application shell is not remounted between Home and the canvas
- [x] A canvas reopened while its first default-view creation is in flight starts no second creation
- [x] Every call site that navigates to `ROUTES.HOME` meaning the canvas is moved to `ROUTES.CANVAS`. Known call sites:
  - the canvas entry in `AppNavigation.tsx`;
  - the `/canvas` → `/` redirect in `main.tsx`, which is removed;
  - `OnePagerActionButton` (checked: it never navigated to `/`);
  - the artifact links on *My edit access*, which used the API href as a route and fell through to the catch-all: they now open the artifact's own page (one-pager, business domain, or the canvas with that view) and the canvas for any other artifact type
- [x] Tests updated:
  - `lib/deepLinks/generators.test.ts`;
  - `components/layout/AppNavigation.test.tsx`;
  - `OnePagerActionButton.test.tsx`
- [x] MSW handlers for `GET /api/v1/home`, with a fixture, serve Vitest, the dev mock API and the mock Playwright project
- [x] E2E:
  - `e2e/helpers.ts` `openApp` and `e2e/mock/edge-details.spec.ts` open `/canvas`;
  - a steward opens `/` and sees their domain's tiles, with stewardship data created through the API and removed after the test;
  - the canvas opens from the navigation;
  - `/?view=` opens the canvas with that view

**Docs and architecture**
- [x] The dependency graph shows `stewardship` depending on `capabilitymapping`, `auth`, `architecturemodeling`, `architecturedirection` and `accessdelegation` only, and remains acyclic (`TestContextDependencyGraphIsAcyclic`)
- [x] `docs/architecture/Stewardship.md`:
  - Purpose and Ubiquitous Language are rewritten for a context of accountability and personal scope (anchor, scope, portfolio, My Work, dominant grade);
  - inbound events, caches and relationship types are listed;
  - an open question records that My Work moves to its own context if it grows beyond anchors (notifications, recents), and that the dominant-grade rule moves to Architecture Direction once a second surface shows one grade per subject
- [x] `docs/architecture/README.md` shows the new edges Stewardship → Architecture Modeling, Architecture Direction and Access Delegation, with their relationship types
- [x] `docs/backend/cross-context-events.md` lists the new subscriptions (checked by the existing catalogue and registry tests); `docs/architecture/components.csv` is regenerated
- [x] Every BDD scenario has at least one corresponding test (see Test Strategy)
- [x] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

The `stewardship` context gains its read side. It newly subscribes to Architecture Modeling, Architecture Direction and Access Delegation and to more Capability Mapping events, all through published languages. No supplier changes. The frontend gains a `home` feature and a routing change.

### Domain Model

No aggregate changes. The read side lives in its own package tree (rule 12): `stewardship/application/home` (scope resolution, composition, redaction, cache projectors) and `stewardship/application/home/readmodels` (cache writers and queries; SQL may live only under a `readmodels` directory, so `architecture_sql_test.go` approves `*/application/*/readmodels/*.go`). It adds:

- **Caller facts** — what the caches hold about the caller: their current stewardships, architected domains, EA-owned capabilities, applications whose user owner they are (with the ownership state), and their unexpired edit grants on capabilities and applications that still exist. Matching the caller, grant expiry and artifact existence are evaluated by the query.
- **Scope resolver** — caller facts → anchors → scope kind, scope domains, anchored subjects (rules 1–2). It decides which ownership states anchor (`owned`, `nominated`) and the relation every anchored subject carries; the query returns the state and never the relation.
- **Subject** — a capability or an application, identified by its type and id; anchors, My Work cards and grade counts are keyed by it.
- **Portfolio query** — scope → tiles (rules 3–5). Effective domains are derived at query time from the capability cache's parent chain and the L1 domain-assignment cache, by a recursive walk that filters on the tenant at every step and stops after 8 levels.
- **My Work query** — anchored subjects → cards with dominant grade (rule 6).

### API Surface

`GET /api/v1/home` — the caller's home; any authenticated user; `200` always (rule 8); no parameters (rule 9). Swagger per `easi-api-standards`. Slices D and E add sections and links to this resource.

```json
{
  "scope": {
    "kind": "personal",
    "stewardships": [
      { "domain": { "id": "d1", "name": "Customer Engagement" }, "concern": "assessment", "concernLabel": "Assessment" }
    ],
    "architectedDomains": [ { "id": "d2", "name": "Finance" } ],
    "architectedDomainCount": 1,
    "eaOwnedCapabilities": 1,
    "ownedApplications": 2,
    "editGrants": 1
  },
  "portfolio": {
    "capabilities": { "total": 12, "byStatus": { "active": 10, "planned": 1, "deprecated": 1 } },
    "applications": { "total": 9 },
    "domains": { "total": 1, "names": ["Customer Engagement"] },
    "time": {
      "total": 40,
      "shares": { "invest": 50, "tolerate": 20, "migrate": 10, "eliminate": 5, "notAssessed": 15 }
    }
  },
  "myWork": {
    "total": 2,
    "items": [
      {
        "subjectType": "capability", "id": "c1", "name": "Invoicing", "level": "L2",
        "relation": "ea-owner", "dominantGrade": "invest",
        "_links": { "x-one-pager": { "href": "/api/v1/one-pagers/capability/c1", "method": "GET" } }
      },
      {
        "subjectType": "application", "id": "a1", "name": "CRM",
        "relation": "edit-grant", "grantExpiresOn": "2026-10-21", "dominantGrade": "eliminate",
        "_links": { "x-one-pager": { "href": "/api/v1/one-pagers/application/a1", "method": "GET" } }
      }
    ]
  },
  "_links": { "self": { "href": "/api/v1/home", "method": "GET" } }
}
```

- `scope.kind` is always present. For `tenant` the anchor fields are empty or zero; for `empty`, `portfolio` and `myWork` are omitted.
- Each tile under `portfolio` and each field gated by rule 8 is omitted on its own. Without `domains:read`, the domain objects in `scope` (`stewardships[].domain`, `architectedDomains`) and `portfolio.domains.names` are omitted, and the counts remain: `architectedDomainCount` is always present for that reason.
- `portfolio.time.shares` is omitted when `total` is 0.
- `myWork.total` counts the items after rule 8's omissions; for `tenant`, `myWork` is `{ "total": 0, "items": [] }`.
- `grantExpiresOn` is a calendar date in UTC; `dominantGrade` is omitted when none.
- `x-one-pager` is present when the caller holds the read permission of the subject.

### Persistence

New tables in the `stewardship` schema. Each has the primary key `(tenant_id, <id>)`, `ENABLE ROW LEVEL SECURITY`, the tenant isolation policy for `easi_app` with `USING` and `WITH CHECK`, and explicit grants, following migration 166. Each is seeded by a `backfill` migration.

| Cache | Content | Fed by |
|-------|---------|--------|
| Capability cache | id, name, level, parent id, status (Active on creation), EA owner value; index `(tenant_id, parent_id)` | `CapabilityCreated`, `CapabilityUpdated`, `CapabilityMetadataUpdated`, `CapabilityParentChanged`, `CapabilityLevelChanged`, `CapabilityDeleted` |
| Domain assignment cache | (L1 capability id, domain id) | `CapabilityAssignedToDomain`, `CapabilityUnassignedFromDomain`, `CapabilityDeleted`, `BusinessDomainDeleted` (idempotent; Capability Mapping also unassigns on domain deletion) |
| Application cache | id, name, ownership state, owner kind, owner id | `ApplicationComponentCreated/Updated/Deleted`, `ApplicationOwnerNominated`, `ApplicationOwnershipConfirmed`, `ApplicationOwnerAssigned`, `ApplicationOwnershipCleared` |
| Realisation cache | realisation id, capability id, component id; direct realisations only | `SystemLinkedToCapability`, `SystemRealizationDeleted`, `CapabilityDeleted`, `ApplicationComponentDeleted` |
| TIME cache | (capability id, component id), grade, assessed at — the pair is unique, as in Architecture Direction, and `TimeAssessmentRemoved` carries no realisation id | `TimeAssessmentRecorded`, `TimeAssessmentRemoved` |
| Edit grant cache | grant id, artifact type, artifact id, grantee e-mail, expires at — no grantor or reason; index `(tenant_id, grantee_email)` | `EditGrantActivated`, `EditGrantRevoked`, `EditGrantExpired` |

Expiry is evaluated against `expires_at` at read time: `EditGrantExpired` is not raised today, and Access Delegation itself enforces expiry on read. The existing domain cache (domain names, domain architect) and user cache from spec 226 are reused unchanged.

The edit grant cache holds grantee e-mails, which are personal data. The canvas lists it with the user cache as stores any future user-erasure flow must purge. Home query errors log the actor id, never the e-mail.

### Frontend

- **Feature `home`** under `/frontend/src/features/`:
  - API client and `homeQueryKeys` (`all`, `detail`) in `queryKeys.ts`;
  - `useHome` with the refetch options in the acceptance criteria;
  - `HomePage`, built from Mantine primitives per `easi-frontend-styling`: scope header, tile row and My Work grid as `SimpleGrid` with breakpoint columns, empty-scope guidance as `Title` + `Text`;
  - the page scrolls inside the main region through a CSS module.
- **Cards** are `Card component={Link}` with relative, basename-aware paths. The frontend maps the presence of `x-one-pager` and the subject type to `/one-pagers/{type}/{id}`; a card without the link is not clickable.
- **Routing:**
  - The route table moves out of `main.tsx` into an `AppRoutes` component.
  - Every application route sits under one pathless layout route that returns `<Navigate to={{ pathname: ROUTES.CANVAS, search, hash }} replace />` for `/` with a `view` parameter and its outlet otherwise. `/` renders the application shell directly, as every other route does, so the shell is never remounted and the redirect still runs before canvas initialisation can read and clear `view`.
  - `ROUTES.CANVAS` renders the canvas; `AppView` and `mainViews` gain `home`.
  - `generateViewShareUrl` targets `/canvas`, and the `view` deep-link parameter is registered for `/canvas` only.
  - The navigation gains a Home entry, and the canvas entry targets `ROUTES.CANVAS`.
  - The brand logo becomes a router link to `ROUTES.HOME`.
- **Canvas initialisation** moves from `App` into the canvas view.
- **Freshness:** Home's query is invalidated by no mutation; with `staleTime: 0` it is refetched on mount and on window focus, since its inputs change in other contexts' pages.

### Cross-Context Integration

| Direction | Events | Purpose |
|-----------|--------|---------|
| Capability Mapping → Stewardship | capability lifecycle, metadata, parent and level events; domain assignment events; direct realisation events | Capability, domain assignment and realisation caches |
| Architecture Modeling → Stewardship | application lifecycle and ownership events | Application cache |
| Architecture Direction → Stewardship | `TimeAssessmentRecorded`, `TimeAssessmentRemoved` | TIME cache |
| Access Delegation → Stewardship | `EditGrantActivated`, `EditGrantRevoked`, `EditGrantExpired` | Edit grant cache |
| Stewardship → Stewardship | `StewardAssigned`, `StewardReleased` | Already projected into the stewardships read model (spec 226) |

---

## Test Strategy

| Tier | Covers |
|------|--------|
| Go unit (scope resolver, pure calculators) | Rule 1 relation per kind of caller fact and the ownership states that anchor nothing, rule 2 scope kind, dominant grade, TIME largest remainder, My Work dedupe, ordering and total, refusal of a grade or capability status the read side does not know |
| Go projector unit (real projectors, fake store) | Every subscribed event, including deletions and `EditGrantExpired` |
| Go integration (`_integration_test.go`, read model and handler) | Rule 1 exclusions the query evaluates (legacy free-text EA owner, team ownership, expired grant, grant expiring exactly now, revoked grant, other artifact type, deleted artifact, released stewardship, actor without e-mail), portfolio, effective domains and reparenting, TIME join on existing realisations, tenant and empty kinds, rule 8 omissions, rule 9, case-insensitive grant match, RLS and tenant isolation, backfills against projector output |
| Vitest + MSW | Section rendering and omission, loading, error and empty states, card links, "Show all", grant expiry text, navigation, logo, router table and redirects, share URL |
| Playwright | Steward home (stewardship created through the API as the bypass identity, removed after), navigation to `/canvas`, legacy `/?view=` |

The E2E stack runs as a single bypass identity, so anchor-less roles and permission omissions are covered by Go integration and Vitest tests, not by Playwright. The bypass user exists only in its own tenant (`acme`) while header-less bypass requests run in `default`, so the steward E2E sends `X-Tenant-ID: acme` for its API setup and from the browser. Expiry tests pass the evaluation instant explicitly, so "expiring exactly now" is deterministic.

---

## Design Decisions

1. **Scope from anchors; role only for the anchor-less** — design doc D1 and D9. `domains:write` is the tenant-fallback criterion because it is the permission that assigns stewards: whoever may distribute accountability may see the whole. Alternative: role check `admin|architect` (rejected — the backend works in permissions, and the two coincide today).
2. **A stewardship brings in the whole domain, whatever the concern** — the portfolio answers "what is my part of the landscape"; the concern answers "what am I accountable for fixing in it" and belongs to attention items. Alternative: per-concern portfolios (rejected — an assessment steward would see TIME counts but no application count for the same domain).
3. **EA ownership anchors the capability only, not its subtree** — `eaOwner` is set per capability; inheriting it would invent a rule spec 200 does not have. Alternative: subtree inheritance (rejected — a new domain rule belongs in Capability Mapping, not in a read side).
4. **Applications belong to domains only through the capabilities they realise** — there is no direct application–domain link. An application realising capabilities in two domains is in both, and an anchored application brings those domains onto the Domains tile. Its capabilities are not added to the Capabilities tile: they are not the caller's to answer for. Alternative: a primary domain per application (rejected — no such fact exists).
5. **Nominated owners see the application** — a nomination is the user's claim awaiting confirmation; hiding it until confirmed hides work they have taken on. It is marked so the difference is visible.
6. **Edit grants anchor capabilities and applications only** — the portfolio counts capabilities and applications; a grant on a view, domain, vendor, team or acquired entity has nothing to count. Alternative: grants on domains bring the domain into scope (rejected — permission to edit is not accountability for a domain's landscape). The existing *My edit access* page still lists every grant.
7. **Grant expiry replaces the design doc's "edit grant expiring" attention item** — expiry is personal, not a concern of a domain, so it is shown on the My Work card rather than raised as a finding for a steward.
8. **One cache per supplier fact, all in `stewardship`** — the context's read side is one language (design doc D3); the Architecture Direction realisation cache is the precedent. Alternative: reuse suppliers' read models over HTTP (rejected — spec 209).
9. **Not assessed is part of the TIME distribution** — a bar of only graded realisations hides how much is ungraded, which is the assessment concern's whole point. Alternative: the mockup's four-grade bar (rejected — overstates assessment coverage).
10. **Old `/?view=` links redirect** — share links were generated on `/` since spec 113; breaking them silently would strand every link in chat and documents.
11. **TIME counts portfolio capabilities and anchored applications** — a steward's bar describes their domain, so an application that is in the portfolio only because it realises a portfolio capability does not bring its realisations elsewhere. An application the caller owns, is nominated for or may edit is theirs as a whole, so all its realisations count. Alternatives: capability or application in the portfolio (rejected — a steward's bar fills with other domains); capability in the portfolio only (rejected — an application owner's bar would be empty).
12. **My Work returns every item** — a person answers for tens of subjects, not thousands; returning all with the total avoids pagination machinery on a composed resource and leaves no item unreachable. The frontend shows 12 first. Alternatives: cursor pagination, a cap with no way to the rest, a separate My Work page (rejected — machinery or a dead end for no gain at this size).
13. **Capability cards open the one-pager** — the Business Domains board can open only capabilities assigned to a domain, and picks one domain arbitrarily for a capability in several; EA-owned and granted capabilities may be in none. The one-pager works for every capability and matches application cards.
14. **Read side in its own package** — invariant 2 forbids analysis from creating dependencies between write sides. The dependency test cannot tell a query's dependency from a command's, so the package boundary keeps Stewardship's commands on the domain and user caches.
15. **One slice** — Home has no value until it is at `/`, and scope, tiles and My Work share the same caches and response. Splitting would ship a placeholder page or a routing change without content.
16. **Canvas initialisation belongs to the canvas** — with Home as the landing page, waiting on views, creating a default view and announcing that canvas data loaded are canvas concerns that would slow and confuse the landing.
17. **The query reads facts, the resolver decides anchors** — matching the caller, expiry and existence are set lookups and stay in SQL; which ownership state is an anchor and what relation a subject carries are rules of this context's language and live in Go, where they are unit-tested and named once. Alternatives: every rule in SQL (rejected — the relation and subject-type vocabulary was spelled in both SQL and Go, held together only by integration tests); every rule in Go (rejected — expiry and existence would ship every expired grant to the application to be thrown away).
18. **A grade or capability status the read side does not know fails the request** — counting only the values it knows would make a tile's total silently smaller than the portfolio. A supplier adding a value changes this context in the same release. Alternative: drop the unknown value (rejected — a wrong number with no error).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Three new suppliers (five in all) | The context's dependency fan-in grows to five published languages | All are leaves toward `stewardship`; nothing imports it, so no cycle is possible |
| Effective domains derived at query time | A recursive walk per request, over the whole tenant for `tenant` scope | Capability trees are shallow; an indexed parent column, tenant filter at every step and a depth cap of 8 |
| Six caches mirror supplier facts | Storage duplication and backfill cost | The spec 209 pattern; every cache is small and keyed by the supplier's id |
| No mutation invalidates Home; `staleTime: 0` | Home refetches on every mount and focus | One indexed read per user; inputs are edited on other pages and in other contexts |
| My Work returns every item | A larger response for heavy owners | Tens of items per person; revisit if a tenant shows otherwise |
| Home lives in `stewardship` | The context's language widens from accountability to personal scope | Canvas rewritten; open question to split My Work out if it grows |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
