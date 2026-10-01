# Design: Personal Home

> Status: approved 2026-09-28 — Phase-1 design document; slices in specs 226–231
> Author: agent + maosk
> Date: 2026-09-28
> Roadmap: moves H2-6 (stewardship read side) and H3-9 (dashboards & KPIs read side), decision SD8 in [`../architecture/ROADMAP.md`](../architecture/ROADMAP.md). Mockup: [`/mockups/easi-home-dashboard-laptop-1600x803-2x.png`](../../mockups/easi-home-dashboard-laptop-1600x803-2x.png).
> Reviewed 2026-09-28 (user decisions): accountability is per domain **and per concern**, not one domain owner (D2); the mockup's capability approval lifecycle and recent views are dropped; application criticality is a separate spec outside this family; `/` becomes Home and the canvas moves to `/canvas`.
> Refined 2026-09-29 while specifying slices B–F (pending user approval): the completeness event is `SubjectCompletenessRecalculated` (spec 135 naming for projection-published events); the overdue rule lives in the shared kernel and Architecture Direction serves it to the timeline (refines D10); journeys (E) precede attention items (D) because the planning checks read the journey cache; edit-grant expiry moves from an attention item to My Work; a slice F moves the timeline onto the served overdue state.

## Problem Statement

EASI opens on the Architecture Canvas. Nothing on that page tells a user what, of everything in the repository, is theirs, what is wrong with it, or what is moving. An architect finds data-quality problems by opening the One-Pager Quality list or by stumbling on them in a domain board; a stakeholder who is accountable for part of a domain's system landscape has no page at all that answers "are the applications I answer for owned, assessed and healthy?"

The home page should answer three questions for the person who opened it: **what is my portfolio, what in it needs my attention, and what is in motion**. Which items appear depends on who the user is and what they are accountable for, not only on their role. Accountability within a domain is divided by concern: one person may answer for owners being assigned, another for TIME assessments being current, a third — once such a feature exists — for budgets being set. The mockup shows the shape — portfolio tiles, a ranked *Needs Attention* list, a *My Work* panel, and a journey strip — and this document maps that shape onto what EASI actually knows.

## Research Summary

- **No home page exists.** `ROUTES.HOME = '/'` renders the canvas; `/canvas` redirects to `/` ([`/frontend/src/main.tsx`](../../frontend/src/main.tsx)). Navigation entries are gated by permission or by session HATEOAS links (`x-one-pager-quality`), never by role inspection.
- **Three roles, static permissions:** admin, architect, stakeholder (read-only plus `assistant:use`). Users have no team, no department and no domain membership ([`/backend/internal/auth/domain/valueobjects/role.go`](../../backend/internal/auth/domain/valueobjects/role.go)). The frontend must not branch on role (spec 199 rule 4); visibility is driven by links in the current-session response.
- **Every link between a person and an artifact today:**

  | Anchor | Where | Who can hold it |
  |--------|-------|-----------------|
  | Domain architect | `BusinessDomain.domainArchitectID` (Capability Mapping) | admin or architect |
  | EA owner | `Capability.eaOwner`, user id since spec 200 | admin or architect |
  | Application owner | `ApplicationComponent` ownership state machine, spec 214: `owned` references a user, `managed` references an `InternalTeam` | any user; teams are not linked to users |
  | Nominated owner | ownership state `nominated`, awaiting an architect's confirmation | any user |
  | Edit grant | Access Delegation, 30-day write grant per artifact, spec 126 | any email, auto-invited as stakeholder |

  There is no accountability field of any kind on a domain. A stakeholder can be reached only through an application they own or an edit grant they hold.
- **Quality signals live in three homes, none aggregated:**

  | Signal | Owner | Served as | Publishes events? |
  |--------|-------|-----------|-------------------|
  | One-pager completeness (required fields filled) | OnePagers, materialised `one_pager_subject_index` | per-row list `GET /one-pager-quality`, per-type map | **No** — OnePagers has no `publishedlanguage` package |
  | Application ownership state, hosting | Architecture Modeling | live `GET /components/ownership-statistics`, tenant-wide counts only | Yes (`ApplicationOwnerNominated` … `ApplicationOwnershipCleared`) |
  | TIME grade per realisation, stale after 12 months | Architecture Direction | per-component rollups only; no tenant- or domain-wide count per grade | Yes (`TimeAssessmentRecorded/Removed`) |
  | Journey status, milestones, target periods | Architecture Direction | `GET /capability-journeys`; "overdue" is computed in the frontend timeline model (spec 197 rule 8), never persisted | Yes (`JourneyPlanned` … `MilestonesReordered`) |
  | Capabilities with no realising application | Capability Mapping | no query; derivable from `SystemLinkedToCapability` / `SystemRealizationDeleted` | Yes |
  | Landscape signals (Eliminate with no journey, standard assessed Tolerate, …) | Architecture Direction, spec 184 | pending, not built | — |

- **No criticality attribute exists on applications.** The application record carries name, description, ownership, hosting, composition and experts. Criticality is specified separately from this family; until it lands the home treats every application in scope alike.
- **No capability approval lifecycle exists.** Capability status is Active / Planned / Deprecated. The mockup's Draft / Proposed / Accepted tile and "L3s pending confirmation" are illustrative and dropped.
- **Access Delegation is the precedent for a small write-side context about people and artifacts:** one aggregate (`EditGrant`), value objects for artifact reference and grantee, published events, artifact-name caches fed by supplier events, HATEOAS affordances on the artifact.
- **Read-side precedent:** the one-pager subject index (projector over supplier events, RLS'd table, backfill migration, keyset-paginated query, session-link-gated nav entry) is the closest template for a ranked attention list; the Architecture Direction realisation cache is the minimal supplier-event cache. Both follow spec 209: projector wired inside the consuming context, no cross-schema SQL at runtime, backfill migrations only.
- **Coverage assessment** (29 Aug, revised 1 Sep) rates G3 *data quality & stewardship* and C6 *reporting, dashboards & KPIs* partial, with the explicit guidance "build a stewardship surface as a read side over published events, not by re-coupling contexts". Spec 119's portfolio dashboard was superseded before it shipped.

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| Home | The page at `/`: the user's portfolio, attention items and journeys in motion, composed for the signed-in user. |
| Concern | One kind of thing a domain's landscape can be wrong about, and therefore one thing a person can be accountable for: **ownership** (every application has an owner), **assessment** (every realisation has a current TIME grade), **documentation** (every one-pager is complete), **planning** (Eliminate grades have journeys, journeys keep to their milestones), **structure** (every capability has a realising application and an EA owner). The list is fixed in code and grows only when a new check exists — a budget concern arrives with a budget feature. |
| Stewardship | The fact that one user is accountable for one concern in one domain. At most one steward per (domain, concern); one user may hold many. Any role. |
| Steward | The user holding a stewardship. Distinct from the domain architect, who curates the model, and from an application owner, who fixes their own application. |
| Scope | What the home is computed over for the caller: the (domain, concern) pairs they steward, every concern in the domains they architect, the capabilities they EA-own, the applications they own, and the artifacts they hold an edit grant on. Never a role. |
| Portfolio | The subjects inside the caller's scope, counted on the home's tiles. |
| Attention item | A computed finding of one concern about one subject that someone should act on. Ranked, live, never acknowledged or dismissed — fixing the data is the only way it disappears (spec 184's framing). |
| Fixer | Whoever an attention item can be routed to for fixing: the subject's owner if known, otherwise the domain architect. Routing reuses invite-to-edit (spec 190). The steward sees the item; the fixer receives the grant. |
| Application health | Derived, never stored: owned (state `owned` or `managed`), every realisation assessed and not stale, one-pager complete, no Eliminate grade without an active journey. |
| Journey in motion | A journey in scope with status planned or in-flight; overdue when a milestone's target period is earlier than the current quarter and not done. |

## Proposed Approach

### Scope first, role second

Personalisation is a function of **scope**, not role. Role decides which sections a user may see at all — and is already expressed as session links and permissions — while scope decides what those sections contain. This keeps the standing convention: the backend composes the home for the caller and gates each section with a link; the frontend renders what it is given and never inspects the role.

### One new context, `stewardship`: a small write side and a read side

The scope resolution has one gap: nobody can be accountable for anything in a domain today. Stewardship is a new fact, not a projection, so it needs an owner. It is owned by a **new bounded context, `stewardship`**, whose single language is *who answers for which concern in which domain, and what currently needs their attention*. The context has the shape of Access Delegation: one small aggregate plus event-fed caches.

**Write side.** A `Stewardship` aggregate keyed by (domain, concern) holding the steward's user reference: assign, reassign, release. Published events `StewardAssigned`, `StewardReleased`. The domain's deletion releases its stewardships (the Access Delegation cascade precedent); a user's disablement does not, so the gap stays visible as an attention item. Assignment requires `domains:write`; anyone who may read the domain may see its stewards. The Business Domains board's domain menu gains a *Stewards…* entry opening a dialog that lists every concern with its steward or "Unassigned", with affordances gated by the links the response carries.

Concerns live in code as a fixed value object, not in MetaModel: a concern is defined by the checks that produce its attention items, and a concern without checks would be accountability for nothing. The Capability Mapping domain architect stays where it is and is treated by the read side as the steward of last resort for every concern.

**Read side.** Per SD8 and invariant 2, everything else is projection: the context subscribes to published events from Capability Mapping, Architecture Modeling, Architecture Direction, Access Delegation, Auth, OnePagers and its own write side, and projects them into local, RLS'd, backfilled caches — subjects and names, anchors (who holds what), realisations, ownership states, TIME grades and dates, journeys with milestones, edit grants, completeness. From those caches it answers, at query time:

| Resource | Answers |
|----------|---------|
| `home` | The caller's composed home: scope summary, portfolio counts and TIME distribution, top attention items, My Work, journeys in motion. One request, links per section. |
| `attention-items` | The full ranked list for a scope (the caller's, or one domain for anyone who may read it), filterable by concern, each item naming its fixer and carrying an `x-edit-grants` routing affordance where the caller may grant. Backs "View all". |
| `stewardships` | Stewards per domain and concern; assignment commands. |

Attention items are computed from the caches at read time in a deterministic order (severity, then domain, then subject), never stored, exactly as spec 184 specifies for signals. Portfolio counts are the same caches counted per scope; they are served from `home` rather than by a separate dashboards context because a count over the stewardship caches is not a second language. H3-9 remains the move for KPI *definitions* and tenant-wide reporting beyond the home.

### OnePagers grows a published language

Completeness is the one signal without published events. OnePagers gains a `publishedlanguage` package with one event, `SubjectCompletenessChanged` (subject type, subject id, complete, missing count), raised by the existing subject-index projector whenever a subject's completeness bucket or missing count changes. Stewardship subscribes to it. This is the only change to an existing context's contract, and it is additive.

### Frontend: `/` becomes Home, the canvas moves to `/canvas`

A new `home` feature renders the composed `home` response with Mantine primitives, one query hook, sections gated on the links the response carries. The canvas keeps its route constant and its navigation entry. Attention rows deep-link to the subject (capability drawer on the domain board, application details, one-pager) through the existing deep-link machinery (spec 113). Rows the caller may route carry the existing `InviteToEditButton`.

### Mockup mapping

| Mockup element | Verdict |
|----------------|---------|
| Capabilities tile with Draft / Proposed / Accepted, "3 L3s pending confirmation", "Accounts Payable pending advance" | Dropped; illustrative. The tile shows Active / Planned / Deprecated. |
| "18 capabilities with no realising application mapped" | Built: a *structure* attention item. |
| TIME distribution bar | Built, per scope, from the TIME cache. |
| My Work with L1 and TIME letter per card | Built: capabilities EA-owned and applications owned by the caller; the letter is the dominant recorded grade across the subject's realisations. |
| Recent views | Dropped. |
| Journey Health strip | Built, per scope; overdue computation moves from the frontend timeline model into the read side so the home and the timeline agree. |

## Key Decisions

| ID | Decision | Rationale / alternative rejected |
|----|----------|----------------------------------|
| D1 | **Scope drives content; role drives only section visibility, via links** | The frontend never branches on role (spec 199). Alternative — a per-role page variant ("architect home", "stakeholder home") — rejected: two users with the same role and different accountabilities need different pages, and role variants would multiply with every future role. |
| D2 | **Accountability is a stewardship: one user per (domain, concern), any role** (user decision 2026-09-28) | Accountability inside a domain is divided by what can go wrong — owners, assessments, documentation, plans, and later budgets — and different people answer for each. Alternatives: a single domain owner (rejected — collapses distinct accountabilities onto one person), a user-level "my domains" preference (rejected — self-selection is following, not accountability), a list of stakeholders per domain (rejected — accountability without one accountable person per concern is nobody's). |
| D3 | **Stewardship is its own bounded context, with a small write side and the read side** | The assignment is a new fact with no existing owner; putting it on `BusinessDomain` would teach Capability Mapping the concern vocabulary, which belongs to the checks. One context keeps one language — steward, concern, attention item — per invariant 6, on the Access Delegation shape. SD8's "pure read side" governs the analysis, which stays pure; the aggregate is not analysis. Alternatives: the frontend composes the home from existing endpoints (rejected — cross-context ranking in the browser, and half the needed queries do not exist); a separate `accountability` context plus a pure-read `stewardship` context (rejected — two contexts for one language). |
| D4 | **Concerns are a fixed value object in code, not MetaModel vocabulary** | A concern exists because checks produce its attention items; a tenant-defined concern with no checks is accountability for nothing. Invariant 5 is about the model's vocabulary (scales, pillars, attributes), not about which computations exist. A new feature with a new check adds its concern in the same spec. |
| D5 | **Attention items are computed at read time and never persisted, acknowledged or dismissed** | Spec 184's framing for signals; an acknowledge workflow lets problems be hidden without being fixed. Only the caches are tables. |
| D6 | **OnePagers publishes `SubjectCompletenessChanged`** | The only way stewardship can know completeness without reading another context's table (spec 209). Alternative — recompute completeness from field-value events — rejected: OnePagers' completeness rule is its own, and duplicating it recreates the pre-208 drift. |
| D7 | **Steward sees, fixer fixes** | The steward of (domain, concern) is who the item appears for; the invite-to-edit target is the subject's owner, falling back to the domain architect. Alternative — route everything to the steward — rejected: an accountable person is usually not the one who edits the record. |
| D8 | **Application health is a derived composite with named checks, not a stored attribute** (criticality is a separate spec, user decision 2026-09-28) | Every check reads an event the read side already has. Criticality is a first-class application attribute in the SD6 shape and is specified outside this family; until it lands the home treats all applications in scope alike. |
| D9 | **Empty scope falls back to the tenant for readers, to a guided empty state for everyone else** | An admin or architect with no anchors is a portfolio-wide reader by role; a stakeholder with no anchors has nothing to steward and is told how to get scope (be assigned a stewardship, own an application, hold an edit grant). |
| D10 | **Overdue moves to the server** (refined 2026-09-29, spec 229 decision 1) | Two computations of "overdue" (timeline model, home) would drift. The rule lives in the shared kernel; the stewardship read side uses it for the home and Architecture Direction serves it to the timeline (slice F), since the timeline reads journeys from their owner. |
| D11 | **`/` is Home; the canvas is at `/canvas`** (user decision 2026-09-28) | The route constant already exists and redirects today; the navigation entry is unchanged. |

## Slice Map

Each slice becomes its own numbered spec and is deployable alone.

| Slice | Spec | Content | Depends on |
|-------|------|---------|------------|
| A | 226 | Stewardship context, write side: `Stewardship` aggregate and concerns, published events, domain-deletion release, `stewardships` resource, *Stewards…* dialog from the domain menu on the board | — |
| B | 227 | OnePagers published language: `SubjectCompletenessRecalculated` published by the subject-index projector on change; consumers backfill from the index | — |
| C | 228 | Home shell: caches and backfills for capabilities, domain assignments, applications and ownership, realisations, TIME, edit grants, user e-mail; scope resolution from anchors; `home` with scope, portfolio tiles and TIME distribution, My Work with grant expiry; `/` becomes Home, canvas at `/canvas`, old `/?view=` links redirect | A |
| E | 229 | Journeys in motion: journey and milestone caches, shared-kernel quarter rule, Journey Health on Home with the timeline's counts | C |
| D | 230 | Attention items: twelve checks across the five concerns, visibility by stewardship, architecture and anchors, ranking, fixer routing via a pre-filled invite-to-edit, `attention-items` list at `/attention`, application health on My Work | B, C, E |
| F | 231 | Timeline reads served overdue: Architecture Direction serves `overdue` and `currentPeriod` with the shared rule; the timeline model stops computing it | E |
