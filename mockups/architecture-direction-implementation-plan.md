# Architecture Direction — Vertical-Slice Implementation Plan

The rollout plan for the model in `architecture-direction-model.md` and the strategic DDD in `architecture-direction-ddd.md`. Each slice is end-to-end (backend + frontend), independently demoable, and revertible. Each slice must pass the model's load-bearing test: *does it help an individual make a daily alignment decision in five seconds?* Slices that fail are not scheduled.

The DDD memo's seven candidates are kept in spirit but re-shaped. Pure-groundwork items (rename, context skeleton, subscriptions) are folded into the first value-delivering slice that needs them so every slice on the timeline ships an alignment answer.

---

## One-page overview

| # | Slice | Spec filename | Size | One-line value |
|---|---|---|---|---|
| 1 | Rename to Logical Capability | `166_LogicalCapability_Rename_pending.md` | M | The vocabulary on screen finally matches the conversation in the room. |
| 2 | Direction on a Logical Capability | `167_Direction_Aggregate_Capture_pending.md` | L | A user opens a Logical Capability and sees in five seconds whether there is a direction, what type, and where it is on the agenda. |
| 3 | Discover view — themes that need a decision | `168_Discover_Candidate_TwoPath_pending.md` | M | A user lands on Discover and sees the open consolidation candidates the group still has to decide. |
| 4 | Standard App designation (Type-2 path) | `169_StandardAppDesignation_pending.md` | M | A user asks "which app should I be using for this logical?" and gets the agreed answer in five seconds. |
| 5 | Direction Map — physical movement canvas | `170_DirectionMap_Canvas_pending.md` | L | A user sees the proposed physical re-shape of the business on one canvas, by status. |
| 6 | Target Architecture by horizon | `171_TargetArchitecture_HorizonView_pending.md` | M | A user picks Now / Next / Later and sees the resulting landscape. |
| 7 | Open Discussions inbox | `172_OpenDiscussions_Inbox_pending.md` | S | A user opens one list and sees everything the group still needs to decide, across Directions and Designations. |

Total: 7 slices, ~8–10 weeks of focused work. Sequence is strictly left-to-right; only slices 3 and 4 are eligible to be reordered against each other.

---

## Slice 1 — Rename to Logical Capability

**Spec filename:** `166_LogicalCapability_Rename_pending.md`
**Bounded contexts touched:** `enterprisearchitecture` (rename in place); `capabilitymapping` (no change); frontend `enterprise-architecture` feature.
**Size:** M

**User value.** The vocabulary on screen finally matches the conversation in the room. The conflated term "Enterprise Capability" disappears, replaced by "Logical Capability". Every alignment conversation that follows refers to the same thing the screen says. This is small but load-bearing — without it, every later slice has to explain the rename inline.

**In scope.** Rename the `EnterpriseCapability` aggregate, `EnterpriseCapabilityLink` (→ `LogicalCapabilityMapping`), strategic-importance aggregate, REST routes, read-model tables, page titles, navigation labels, and the term "Enterprise Capability" anywhere it surfaces. Event upcasters translate historical events at deserialization. All existing behaviour is preserved.

**Out of scope.** Any new fields, any new behaviour, the `architecturedirection` context, the package rename of `enterprisearchitecture` itself (the package keeps its name; only the aggregate inside changes). No new view tabs.

**Acceptance criteria.**
- A user logging in after deploy sees "Logical Capabilities" in the navigation and on the page; no UI surface still says "Enterprise Capability".
- All existing data is reachable at the new routes; old routes either redirect or are removed in the same release.
- Event store is unchanged on disk; replay produces identical projections.
- Existing tests pass with no behavioural changes; tests reference the new names.

**Validation question.** *Before we continue, can you confirm the term "Logical Capability" reads correctly in every place it now appears, and that no daily workflow you currently rely on has regressed?*

**Dependencies.** None.

---

## Slice 2 — Direction on a Logical Capability

**Spec filename:** `167_Direction_Aggregate_Capture_pending.md`
**Bounded contexts touched:** new `architecturedirection` context (greenfield); `enterprisearchitecture` (read-only reference); frontend Logical Capability detail view.
**Size:** L

**User value.** A user opens any Logical Capability and within five seconds knows: *is there a direction on this; what type (consolidate / decompose / stay); where is it on the agenda (draft / proposed / agreed)*. Architects can capture and progress a direction without leaving the detail view. This is the first slice where the model genuinely answers the alignment question.

**In scope.** The `architecturedirection` bounded context skeleton, the `Direction` aggregate (type, source capabilities, placements, horizon, status, narrative), status workflow (`draft → proposed → agreed | rejected`), the integration subscription that lets `architecturedirection` know which physical capabilities exist, and a Direction panel on the Logical Capability detail page that shows current status with a status badge and lets architects create / edit / advance directions.

**Out of scope.** Discover view, Standard App designation, the Direction Map canvas, the horizon timeline view, the Open Discussions list, batch operations, history visualisation. No realization tracking. No notification fan-out beyond what the existing notification infra already does.

**Acceptance criteria.**
- A user with the right permission can create a `consolidate` direction on a Logical Capability, pick 2+ source physical capabilities, set placements + horizon + narrative, and save it as `draft`.
- The Logical Capability detail view shows the direction's type, status, and one-line narrative in a panel visible above the fold.
- An architect can advance status `draft → proposed → agreed` and the badge updates in five seconds.
- Status transitions are recorded as separate past-tense events; replay reconstructs the current status correctly.
- Deleting a referenced physical capability surfaces a "stale reference" indicator on the direction; it does not block the deletion.

**Validation question.** *Before we continue, can you confirm that opening any Logical Capability now answers the alignment question in five seconds for the simple case (single direction, one logical), and that the create / advance flow matches how the architect group actually works?*

**Dependencies.** Slice 1.

---

## Slice 3 — Discover view: themes that need a decision

**Spec filename:** `168_Discover_Candidate_TwoPath_pending.md`
**Bounded contexts touched:** `architecturedirection` (`DiscoveryCandidate` aggregate); frontend new `Discover` route under Architecture Direction.
**Size:** M

**User value.** A user opens Discover and sees the open consolidation candidates the group has not yet decided on. From each candidate, the user can promote to either a physical-consolidation Direction *or* a Logical grouping (with optional Standard App later) — or reject. The two-path promotion is the heart of the model.

**In scope.** The `DiscoveryCandidate` aggregate (open / promoted-to-direction / promoted-to-grouping / rejected), manual creation of a candidate (architect adds a theme by name + sources), the Discover tab in the UI matching the mockup's matrix layout, and the two-path promotion handlers. Candidates link forward to the Direction or Logical Capability they produced.

**Out of scope.** Automatic generation of candidates from cross-domain signal (deferred — the aggregate exists; how it is opened is downstream). The application-portfolio surface for candidates. Any decision UI for Standard Apps (slice 4).

**Acceptance criteria.**
- An architect can add a manual candidate by name, attach 2+ physical capabilities from different domains, save.
- The Discover tab lists every open candidate; rejected and promoted ones are not in the default view.
- Promoting to a Direction creates a `Direction` (slice 2's flow) pre-filled from the candidate; promoting to a grouping creates a `LogicalCapability` pre-filled with the candidate's sources mapped in.
- After promotion, opening the produced Direction / Logical Capability surfaces a back-link to the originating candidate.

**Validation question.** *Before we continue, does the matrix layout and the two-path promotion match how the group actually triages candidates in the room? Specifically — is the manual-add path enough for now, or is the automated-detection question urgent?*

**Dependencies.** Slice 2 (so promote-to-direction has a target). Independent of slice 4.

---

## Slice 4 — Standard App designation (Type-2 path)

**Spec filename:** `169_StandardAppDesignation_pending.md`
**Bounded contexts touched:** `architecturedirection` (`StandardAppDesignation` aggregate); ACL to the existing application catalog; frontend Logical Capability detail and a new Application Portfolio tab.
**Size:** M

**User value.** A user asks "for this logical capability, which application should I be using?" and gets an answer in five seconds: *agreed standard is X; alternatives currently in use are Y, Z*. This is the Type-2 path from the model — physical reality stays distributed, the standard exists at the logical layer where it actually applies.

**In scope.** The `StandardAppDesignation` aggregate (one-app-per-logical, status workflow, supersession on replacement), the ACL wrapping the application-catalog lookup, a Standard App panel on the Logical Capability detail view, and a basic Application Portfolio tab listing every Logical Capability with its agreed standard app (or "none / under discussion").

**Out of scope.** Direction Map canvas. Cross-application impact analysis. Bulk designation. Cost / fit-score data on the Application Portfolio tab.

**Acceptance criteria.**
- An architect can attach a draft Standard App designation to any Logical Capability, advance it to `agreed`, and the panel updates.
- Replacing an `agreed` designation with a new one supersedes the old (which is preserved for audit, not deleted).
- The Application Portfolio tab lists every Logical Capability with its standard app or "no standard yet"; a non-architect can scan it.
- A Discover candidate (slice 3) promoted to a grouping can have a Standard App attached as a follow-up without reopening the candidate.

**Validation question.** *Before we continue, does the Application Portfolio tab as a flat list answer the daily question for product managers and engineers, or does it need grouping (by domain, by current app, by horizon)?*

**Dependencies.** Slice 2. Independent of slice 3.

---

## Slice 5 — Direction Map: physical movement canvas

**Spec filename:** `170_DirectionMap_Canvas_pending.md`
**Bounded contexts touched:** read-model module aggregating `architecturedirection` + `capabilitymapping` events; frontend new Direction tab using the existing canvas/dockview infrastructure.
**Size:** L

**User value.** A user opens the Direction tab and sees, on one canvas, every proposed physical re-shape of the business — which capabilities consolidate into which target domain, which decompose, which stay. Filtering by status (Draft / Proposed / Agreed) lets the architect group focus the conversation. This is where the *shape* of the target architecture becomes visible at a glance.

**In scope.** A read projection that joins Directions to their source capabilities and target domains, a canvas view rendering domains as zones and Directions as movement arrows between them, status filter, click-through from any Direction marker to its detail / edit drawer (re-using slice 2's editor).

**Out of scope.** Horizon scrubbing (slice 6). The application-portfolio overlay. Editing layout / pinning positions. Decomposition entry-point UX (the model says decomposition starts from a single physical capability; that is a separate flow not in this slice).

**Acceptance criteria.**
- The canvas renders all current Directions grouped by status, with the visual key matching the agreed mockup.
- A user can click any Direction on the canvas, the drawer opens with full context, edits propagate back to the canvas without reload.
- The status filter narrows the canvas in under a second on the production data set.
- A non-architect viewing the canvas can describe the proposed re-shape of one domain in five seconds.

**Validation question.** *Before we continue, does the canvas at production scale (~1300 capabilities, dozens of Directions) still pass the five-second test, or do we need pre-filters by domain or horizon before this becomes usable?*

**Dependencies.** Slice 2. Slice 4 not required.

---

## Slice 6 — Target Architecture by horizon

**Spec filename:** `171_TargetArchitecture_HorizonView_pending.md`
**Bounded contexts touched:** read-model module (extends slice 5's projection with horizon filter); frontend new Target Architecture tab.
**Size:** M

**User value.** A user picks Now, Next, or Later and sees the resulting landscape of physical capabilities by domain, with each capability tagged as native, inbound (moving in from another domain), decomposed, or transitional. This makes the *path* from current to target visible without dates — exactly the abstraction the model insists on.

**In scope.** A horizon scrubber (Now / Next / Later), a per-domain layout showing the projected physical capabilities for the chosen horizon, visual classification of each capability (native / inbound / decomposed / transitional), and the standard app inline on each capability when one is agreed (slice 4 data).

**Out of scope.** Editing from this view (read-only synthesis). Date-based projections. Comparison views (diff between Now and Later) — if needed, that becomes its own spec. Roadmap-style swim-lanes.

**Acceptance criteria.**
- Switching horizon recomputes the landscape in under a second on production data.
- A capability that consolidates from three domains to one renders as `inbound` in the target domain at the chosen horizon, and as `transitional` in each leaving domain.
- Standard-app annotations appear on capabilities whose Logical has an agreed designation at or before the chosen horizon.
- A non-architect product manager can answer "where will my capability live next year?" in five seconds.

**Validation question.** *Before we continue, does the three-state horizon enum carry enough resolution for the conversations you have, or does the lack of a fourth ("never" / "long horizon") show up as a gap?*

**Dependencies.** Slices 2, 4, 5.

---

## Slice 7 — Open Discussions inbox

**Spec filename:** `172_OpenDiscussions_Inbox_pending.md`
**Bounded contexts touched:** read-model in `architecturedirection` (a flat projection over both Direction and StandardAppDesignation events); frontend new tab.
**Size:** S

**User value.** A user opens one list and sees everything the group still has to decide — every Draft and Proposed Direction, every Draft and Proposed Standard App designation. Click any item, jump to its decision context. This is the agenda for the next architect-group meeting, generated rather than maintained.

**In scope.** A combined projection of all `Direction` and `StandardAppDesignation` aggregates currently in `draft` or `proposed` state, sortable by last-touched, with the originating narrative inline. Click-through navigates to the editor in slice 2 or slice 4. No new aggregate; this is read-side.

**Out of scope.** Comment threads / message-board UI. Notifications. Per-user inboxes. Voting. Anything that turns this into a project tool.

**Acceptance criteria.**
- The list updates in under a second when a Direction is advanced from `draft` to `proposed` (or out of the list when it goes to `agreed`).
- Each row shows: type (Direction / Standard App), affected logical, status, narrative summary, last-touched timestamp.
- Clicking a row opens the corresponding editor in the existing tab.
- Empty state reads "Nothing on the agenda" — that itself is a five-second alignment answer.

**Validation question.** *Before we close out the rollout, is this list the agenda the group actually wants to walk through, or do you need it cut by domain or by stakeholder before it becomes the meeting prep tool?*

**Dependencies.** Slices 2, 4.

---

## Cumulative-value timeline

- **After slice 1** the vocabulary on every screen matches the strategy conversation. The rename pays off every later slice.
- **After slice 2** the user can ask "is there a direction on this Logical Capability?" and answer in five seconds. The core alignment question works for the single-capability case.
- **After slice 3** the user can also ask "what consolidation themes still need a decision?" and the group has a triage surface with the two-path promotion.
- **After slice 4** the user can additionally ask "which application should I be using for this logical?" — the Type-2 path is live, and the Application Portfolio tab is usable on its own.
- **After slice 5** the user can also see the *shape* of the proposed physical re-shape on one canvas and filter by status. The Direction Map is the architect-group's working surface.
- **After slice 6** the user can pick a horizon and see the landscape that horizon implies, including standard-app annotations. The path from now to target is visible without dates.
- **After slice 7** the user has a single agenda surface for everything the group still has to decide. The model is feature-complete against the mockup.

The rollout is reversible at every step — slice N can be rejected without orphaning slices 1..N-1. Slices 3 and 4 can be reordered against each other if user feedback prioritises Type-2 over Discover or vice versa.

---

## Open decisions before slicing starts

These are pulled out so they do not get buried in the spec drafting.

1. **Top-level navigation placement.** The mockup uses an "Architecture Direction" page with five tabs. Does that live alongside Enterprise Architecture in the main nav, replace it, or sit under it? Slice 2 needs the answer before its UI lands.
2. **Permissions.** Slices 2–4 introduce new write actions (create direction, advance status, attach standard app, promote candidate). Do these inherit `enterprise-arch:*` or get their own `architecture-direction:*` permissions? The DDD memo defers this; the spec writer needs the call before slice 2.
3. **Manual vs automatic candidates.** Slice 3 ships manual-add only. Confirm that is acceptable for the first cut, or whether a basic cross-domain-signal detector blocks slice 3.
4. **Old route preservation.** Slice 1 — for one release, do the old `enterprise-capabilities` API routes 301-redirect, duplicate-handler, or get removed immediately? Affects external consumers (if any).
5. **Decomposition entry point.** The model says decomposition starts from a single physical capability, not from Discover. Slice 2 supports `decompose` as a Direction type, but the *entry-point UX* for "I want to decompose this capability" is unspecified. Confirm whether that is in slice 2's UI or deferred to a later spec.
6. **Scale of the canvas.** Slice 5 lands a canvas; production has ~1300 capabilities. Confirm whether the slice ships with a default filter (by domain, by horizon) or unfiltered. Affects acceptance criteria.
7. **"Stay" direction usage.** The model says `stay` is rare, used only when the group has explicitly evaluated. Confirm that slice 2's UI exposes `stay` on equal footing with `consolidate` / `decompose`, or hides it behind a less-prominent affordance.
