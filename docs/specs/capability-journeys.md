# Design: Capability Journeys

> Status: DRAFT — awaiting Phase 1 human approval
> Author: agent + maosk
> Date: 2026-07-12
> Source of truth: [`mockups/capability-journey-mockup.html`](../../mockups/capability-journey-mockup.html) (v3). Where the mockup and this design conflict, the mockup wins unless a decision below records otherwise.

## Problem Statement

The Domain Board (spec 179) answers "where are we today": business domains, their capabilities, and the applications realising them. It says nothing about **where we are going** or **how we get there**. Migrations, consolidations, carve-outs, and capability moves live in slides; TIME judgements exist only as system-computed suggestions nobody has signed; and there is no way to see, per capability, whether the landscape is settled, in flight, or not yet started.

Capability Journeys adds a time dimension to the landscape as three lenses over the same board:

| Lens | Question | Data |
|------|----------|------|
| **Now** | Where are we today? | Realisations + per-realisation TIME assessments + standard/legacy roles |
| **Journey** | What is done, in flight, not started? | Journeys (migration / consolidation / carve-out / move) with progress and milestones |
| **Target** | Where are we going? | Standard roles + journey target applications + move destinations |

A **Signals** engine surfaces where these layers disagree (an Eliminate-assessed app with no journey; a standard app assessed Eliminate) — governance questions, not errors.

## Research Summary

- **The Now lens already exists.** Spec 179 rebuilt Business Domains as the Domain Board on this mockup's Now pattern (domain grid → collapsible L1 groups → capability cards → app chips with `realizationLevel` tint and `Inherited` dimming). It explicitly deferred TIME/standard badges to a later spec — this one.
- **Realisations** (`capabilitymapping.CapabilityRealization`) carry level `Full|Partial|Planned`, notes, origin `Direct|Inherited`. Inheritance propagates direct realisations upward to ancestor capabilities (spec 131B), always as `Full`, deduplicated per (capability, component). No role concept exists.
- **TIME is computed-only.** Spec 118's `TimeSuggestionCalculator` derives a suggestion per realisation from strategic-fit pillar gaps at query time. No architect-assessed TIME is persisted anywhere. Spec 119 (pending, never built) proposed assessed TIME per *enterprise capability* + application — superseded by this design's per-realisation model.
- **`architecturedirection`** records the architecture group's decisions: `Direction` (consolidate/decompose/stay on an EC, draft→proposed→agreed→rejected, sources compose the EC per spec 172) and `StandardApplication` (one per EC, set-and-replace with history, spec 170). Both use handler-level one-per-parent uniqueness with DB backstop, stale-reference marking, denormalised name caches, and the `architecture-direction:*` permission family.
- **Capability levels** are L1–L4; domain membership is effective (a capability belongs to a domain if it or an ancestor is assigned).
- Frontend: board data is fan-out queries per domain (`useDomainBoardData`), view model in `domainBoardViewModel.ts`; tokens already define the status colours the lenses need (`--status-positive/progress/neutral/future/danger`).

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **TIME Assessment** | An architect's recorded TIME grade (Invest / Tolerate / Migrate / Eliminate) for one application's realisation of one domain capability — "per app *per capability*", with assessor and date. An assessment, not a commitment. |
| **TIME Suggestion** | The system-computed grade from strategic-fit gaps (spec 118). Evidence; never overwritten by an assessment. |
| **Realization Role** | Per-realisation designation: `standard` (the blessed app for this capability today), `legacy` (kept but not blessed), or unset (unclassified). Independent of TIME. |
| **Journey** | The recorded change story of one domain capability: kind, from-apps → to-app, status, progress, target period, milestones. At most one active per capability. |
| **Journey kind** | `migration` (realisation moves to another app), `consolidation` (several apps merge onto one), `carve-out` (functionality extracted from one app into another), `move` (the capability relocates to another domain/parent under a new name). |
| **Milestone** | A dated step inside a journey: label, target period, own status. |
| **Target period** | A structured year + quarter (sortable), never display text. |
| **Lens** | Presentation mode of the Domain Board: Now / Journey / Target. |
| **Signal** | A computed disagreement between assessments, roles, and journeys: gap, priority, contradiction, tension, carve-out candidate. |
| **Stale assessment** | An assessment older than 12 months. |

## Proposed Approach

### Ownership: three new aggregates in `architecturedirection`

TIME assessments, realization roles, and journeys are all *the architecture group's judgements and plans layered over* the `capabilitymapping` structure — the same ubiquitous language as `Direction` and `StandardApplication` (intent, judgement, consensus), not the structural language of `capabilitymapping`. The context already holds the architect permission family, stale-reference machinery, and name caches, and already references capabilities and applications read-only.

| Aggregate | Keyed by | Mirrors |
|-----------|----------|---------|
| `TimeAssessment` | (capability, component) pair, intrinsic UUID | `StandardApplication` mechanics: set-and-replace, history from events, handler-level uniqueness + DB backstop |
| `RealizationRoles` | capability (one aggregate per capability, holding component → role) | makes "at most one standard per capability" intrinsic |
| `CapabilityJourney` | capability (one *active* per capability, enforced handler-level) | `Direction` lifecycle discipline: discrete transition events, immutable kind, terminal states |

`capabilitymapping`, `architecturemodeling`, `enterprisearchitecture` are unaffected on the write side. Signals and board queries are query-time read-side composition inside `architecturedirection` (the `TimeSuggestionReadModel` precedent), joining its own read models with denormalised capability/component/domain names.

### Frontend: lenses on the existing Domain Board

No new page. The Business Domains board gains a Now / Journey / Target lens switcher (lens in the URL for deep links); Now is today's board plus TIME badges and role-tinted chips; Journey and Target are alternative card renderers over the same view model. The capability drawer gains assessment, role, and journey sections, all HATEOAS-gated. The assign rail stays a Now-lens affordance.

Sub-capability journey progress (the mockup's L3 rows) is **derived** from the descendants' real realisations against the journey's from/to apps — no duplicate bookkeeping (see D6).

### Relationship to existing concepts

- **Specs 119 and 120 are superseded**: assessed TIME moves from EC-level to realisation-level (119), and 120's rationalization analysis rested on 119 plus EC linking removed by spec 172; its standardization-gap questions are covered at realisation level by the signals (184).
- **`StandardApplication` (EC-level) and Direction are unchanged.** Realization roles answer "which app for this *domain capability* today"; the EC standard remains "which app for this *enterprise capability*". They coexist at different granularity.
- **TIME suggestions (118) remain** as read-only reference/prefill beside assessments.

## Key Decisions

| # | Decision | Rationale / alternative rejected |
|---|----------|----------------------------------|
| D1 | **TIME assessed per capability-app realisation** (user decision 2026-07-12) | Mockup: "TIME grades are assessed per application *per capability*". EC-level (spec 119) rejected: wrong granularity, sparse coverage. |
| D2 | **Standard/legacy is a per-realisation designation** (user decision 2026-07-12) | Deriving from EC `StandardApplication` via direction composition would leave the board blank wherever no EC direction exists. Deriving from TIME (Invest ⇒ standard) rejected: the tension signal exists precisely because role and TIME can disagree. |
| D3 | **Journey from/to apps are catalog references only** (user decision 2026-07-12) | A future app ("Pricing Engine") is created as a component first; its realisation can be `Planned`. Free text rejected: no rollups, links, or timeline; reconciliation debt. Deleted refs use the existing stale pattern. |
| D4 | **Move journeys are plan-only** (user decision 2026-07-12) | Completing a move records it; the actual reparent/rename/reassignment goes through existing capability operations. Auto-execute rejected: cross-context write cascade from a status transition. |
| D5 | **Lenses extend the Domain Board** (user decision 2026-07-12) | Spec 179 built the board on this mockup's Now pattern. A separate Landscape page rejected: two near-identical boards. |
| D6 | **Sub-capability journey status is derived**, not stored | For each descendant with direct realisations: realises to-app `Full` and no from-app → done; realises to-app (any level) alongside a from-app, or to-app only at `Planned/Partial` → in flight; only from-apps → not started. Single source of truth; the mockup's stored child rows are reproduced by derivation. |
| D7 | **One aggregate per concern**, all in `architecturedirection` | Follows 167/170 precedent (Direction vs StandardApplication stayed separate). A new bounded context rejected: identical language, pure infrastructure overhead. Putting role/TIME on the `CapabilityRealization` aggregate rejected: architect-permission surface and judgement language don't belong in `capabilitymapping`. |
| D8 | **Timeline-readiness is an architectural constraint, not a feature** | A later iteration will show how the architecture evolved and will evolve. This series must not preclude it: every state change is an event carrying actor + occurred-at (house standard); target periods are structured and sortable; journeys and assessments are never hard-deleted — done/abandoned journeys and superseded assessments remain reconstructable history. No new state may live outside the event store. |
| D9 | **Signals are computed, never persisted** | "Questions for governance, not errors" (mockup). No acknowledge/dismiss workflow — resolving the underlying data resolves the signal. |

## Slice Map

| Spec | Slice | Depends on |
|------|-------|------------|
| 180 | TIME assessment on capability realisations (aggregate, API, badges + drawer assess UI, rollup, stale) — supersedes 119 | 118, 179 |
| 181 | Realization role standard/legacy (aggregate, API, chip tinting, drawer role control) | 179 |
| 182 | Capability Journey (aggregate incl. move kind, API, drawer capture/progress/milestones UI) | 179 |
| 183 | Domain Board lenses (switcher, journey & target renderers, moves ghost/arriving + trace, changes-only, summary) | 180, 181, 182 |
| 184 | Signals (computed engine, toolbar count, drawer, click-through) | 180, 181, 182, 183 |

Each slice is independently deployable: 180–182 each deliver visible value in the existing board/drawer before the lens switcher (183) exists.
