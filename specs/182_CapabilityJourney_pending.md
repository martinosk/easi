# 182 — Capability Journey

> **Status:** pending
> **Depends on:** 179 (Domain Board drawer as the capture surface)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)

---

## Problem Statement

The change story of a capability — "Booking management is migrating from Seabook to Phoenix, route by route, done Q2 2027, 60% there" — lives in slides and heads. EASI records where realisations *are*, and (per spec 167/170) what the group *intends* at EC level, but nothing captures the executed path between now and target for a domain capability: what kind of change, from which apps to which app, how far along, and which milestones remain.

This slice introduces the **Journey**: the recorded change story of one domain capability. Four kinds — `migration`, `consolidation`, `carve-out`, and `move` (the capability relocates to another domain/parent under a new name) — with a status workflow, manual progress, a structured target period, and milestones. A capability has at most one active journey; completed and abandoned journeys are kept forever as the history of how the landscape evolved.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Capture a journey on a capability, advance its status, maintain progress and milestones, and record its completion or abandonment. |
| **Engineer / Product Manager / Stakeholder** | Open a capability and see in five seconds whether change is planned, in flight, or done — from what, to what, by when. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Journey on a domain capability

  Scenario: An architect plans a migration journey
    Given capability "Booking management" is realised by "Seabook" and "Phoenix"
    When the architect captures a migration journey from "Seabook" to "Phoenix" with a note and target period Q2 2027
    Then the journey exists in status "planned"
    And the capability drawer shows the transition (kind, from → to, status, target period) and the note

  Scenario: Kind-specific source cardinality is enforced
    Given the architect is capturing a journey
    Then a consolidation with fewer than two from-apps is rejected
    And a carve-out with other than exactly one from-app is rejected
    And a migration with zero from-apps is rejected

  Scenario: The target application must exist in the catalog
    Given no application named "Pricing Engine" exists
    When the architect attempts to capture a carve-out targeting "Pricing Engine"
    Then the capture is rejected
    And after creating the "Pricing Engine" component the same capture succeeds

  Scenario: An architect advances and progresses a journey
    Given a planned journey
    When the architect starts it, sets progress to 60, and marks a milestone done
    Then the journey is "in flight" at 60% with the milestone recorded as done
    And each change is recorded as a discrete event with actor and timestamp

  Scenario: Completing a journey freezes it and touches nothing else
    Given an in-flight migration journey
    When the architect completes it
    Then the journey is "done" and no further edits are accepted
    And no capability, realisation, or domain assignment was modified by the completion

  Scenario: A journey can be abandoned
    Given a planned or in-flight journey
    When the architect abandons it
    Then the journey is "abandoned", frozen, and preserved for audit

  Scenario: At most one active journey per capability
    Given a capability with a planned or in-flight journey
    When the architect attempts to capture a second journey on it
    Then the capture is rejected with a reference to the existing journey
    And after the existing journey is completed or abandoned, a new capture succeeds

  Scenario: An architect plans a move journey
    Given L1 capability "Invoicing" in domain "Ferry freight"
    When the architect captures a move journey to domain "Group functions" under parent "Accounts payable", as "Freight invoicing", realised by "SAP S/4"
    Then the journey exists in status "planned" with the destination recorded
    And completing it later does not reparent, rename, or reassign the capability

  Scenario: Deleted references go stale but never block
    Given a journey whose from-app has since been deleted from the catalog
    When any user views the journey
    Then the missing reference is marked stale and the journey renders normally

  Scenario: Read-only users see journeys but cannot write
    Given a user without the architect permission
    When they fetch a capability with a journey
    Then the journey is fully readable but the response carries no capture/advance/edit affordances
```

---

## Business Rules & Invariants

1. **One active journey per capability.** A journey is *active* while `planned` or `in-flight`. A second capture while one is active is rejected, referencing the existing journey. `done` and `abandoned` journeys never block a new capture.
2. **Kind** — `migration` | `consolidation` | `carve-out` | `move`. Immutable after capture; to change kind, abandon and capture anew (167 decision 5 posture).
3. **Source cardinality by kind** — migration ≥ 1 from-app; consolidation ≥ 2; carve-out exactly 1; move 0..n (the capability's current realisations are the implicit sources).
4. **Target application is required for every kind** and must reference an existing application component at capture time; the target must not be among the from-apps. A not-yet-built app is created in the catalog first (its realisation may be `Planned`).
5. **Move destination** — a move requires a target business domain; optionally a target parent capability, which must effectively belong to that domain at capture time; and a resulting name, defaulted from the capability's current name and editable while the journey is editable (167 rule 7a posture: stored on the journey, never re-resolved).
6. **Status workflow** — `planned` → `in-flight` → `done`; `abandoned` reachable from `planned` or `in-flight`. Terminal states (`done`, `abandoned`) freeze the journey entirely. One discrete past-tense event per transition; no generic status-changed event.
7. **Progress** — a manually maintained integer 0–100, optional. Not derived from milestones (they measure different things; the mockup's 35% with 1/3 milestones done is intentional).
8. **Milestones** — an ordered list; each has a required label, an optional structured target period, and its own status (`planned` | `in-flight` | `done`). Editable only while the journey is active; every change is a recorded event.
9. **Target periods are structured** — year + quarter value objects, sortable and comparable. Never free text.
10. **Plan-only** — no journey transition ever mutates `capabilitymapping` or any other context. Completing a move records the completion; the actual reparent/rename/reassignment goes through existing capability operations.
11. **References go stale, never block** — deleted capabilities, components, domains, or parent capabilities referenced by a journey are marked stale on read (167 decision 4); journeys render and transition regardless.
12. **Journeys are never hard-deleted.** Done and abandoned journeys are the retained history of the landscape's evolution.
13. **Every event carries the acting user and occurred-at timestamp.**
14. **Authorisation** — capture, edit, and transitions require the `architecture-direction:*` architect permission; reads follow board read permission; all writes HATEOAS-gated.

---

## Acceptance Criteria

- [ ] An architect can capture a journey of each kind with kind-appropriate fields; violations of rules 3–5 are rejected with clear errors
- [ ] The capability drawer shows the journey: kind, from → to, status, progress, target period, note, milestones, and destination for moves
- [ ] Status transitions (start, complete, abandon) are discrete operations producing discrete past-tense events; replay reconstructs the status
- [ ] Progress, note, target period, and milestones are editable while active and rejected once terminal
- [ ] A second capture on a capability with an active journey is rejected; capture succeeds after completion/abandonment; prior journeys remain queryable as history
- [ ] Completing any journey (including a move) provably performs no write outside `architecturedirection`
- [ ] Journeys with deleted references render with stale markers
- [ ] A bulk query returns journeys (active + terminal) for a set of capabilities/domain in one request, for the board
- [ ] Write affordances are HATEOAS-gated; read-only users receive none
- [ ] Every BDD scenario has at least one corresponding test; every business rule has a unit test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

`architecturedirection` owns the `CapabilityJourney` aggregate (design doc D7): a journey is the group's recorded plan and its progress — intent language, architect permissions, existing stale-reference and name-cache machinery. `capabilitymapping` (capabilities, domains), `architecturemodeling` (components) referenced read-only.

### Domain Model

`CapabilityJourney` — intrinsic UUID; carries capability ref, kind VO, from-app refs, to-app ref, status VO, optional progress, optional target period VO (year + quarter), note, ordered milestones (entity list local to the aggregate), and for moves: target domain ref, optional target parent ref, resulting name. Events (discrete, past-tense): `JourneyPlanned`, `JourneyStarted`, `JourneyCompleted`, `JourneyAbandoned`, `JourneyProgressUpdated`, `JourneyDetailsUpdated` (note / target period), `JourneyMilestoneAdded` / `JourneyMilestoneUpdated` / `JourneyMilestoneRemoved`, `JourneySourceApplicationsChanged`. One-active-per-capability enforced at the command handler via read model, DB partial-unique backstop (167/170 pattern). Capture verifies component/domain/parent existence via the established reference checker.

### API Surface

Journeys under the capability resource tree: the capability's active journey inline or by link, discrete transition operations, milestone sub-operations, and a journey-history sub-resource. A bulk query by capability set / domain serves the board. Shapes per `easi-api-standards`; HATEOAS gates every write.

### Persistence

Event-sourced. Read models: current journey per capability (active + most recent terminal) and journey history, with denormalised capability/component/domain names and stale flags; partial unique index on active journeys per (tenant, capability).

### Frontend

New `journeys` feature folder. The Domain Board capability drawer gains a Journey section: capture form (kind-driven fields, app pickers over the catalog, structured period picker), transition table (kind, from → to, status, progress, target period), plan-summary note, milestone list with per-milestone status, and status/progress actions — all HATEOAS-gated. Query keys + `mutationEffects.ts` invalidation for board journey queries.

### Cross-Context Integration

Subscribes to `capabilitymapping` capability/domain deletion events and `architecturemodeling` component deletion events for stale marking (existing wiring). Publishes journey lifecycle events into the published language (consumed by 183 board queries and 184 signals within the context). No outbound writes — rule 10 is absolute.

---

## Design Decisions

1. **Journey attaches to the domain capability** — the mockup's unit of storytelling; realisations and target designations (180/181) are per capability, so the change story must be too. Attaching to apps (a migration touches ≥2) or ECs (most capabilities sit under none) rejected.
2. **`move` is a journey kind, not a separate aggregate** — a move and an app-transition compete for the same slot in the story ("what is happening to this capability?"); one aggregate makes the one-active-story invariant intrinsic. The mockup renders them as alternatives on a card, never both.
3. **Catalog references only** — user decision 2026-07-12. Future apps enter the catalog first; `Planned` realisation level already models "not live yet". Free text rejected: no rollups, links, signals, or sortable history.
4. **Plan-only completion** — user decision 2026-07-12. A status transition must not trigger a cross-context write cascade; the existing reparent/rename operations already carry their own invariants (spec 131) and stay the single way to mutate the model.
5. **Manual progress, independent of milestones** — mockup-literal; percent measures scope (routes migrated), milestones measure plan steps. Deriving one from the other rejected as false precision.
6. **Structured target periods** — periods must sort and compare (design doc D8); a free-text "Q2 2027" cannot. Quarter granularity matches how architects actually plan (mockup throughout).
7. **One active journey per capability** — the five-second answer degenerates if two stories compete (167 decision 3). Parallel workstreams are milestones within one journey; genuinely independent successive changes are successive journeys.
8. **Terminal journeys retained forever** — they are the "how we got here" record (design doc D8); the mockup renders done journeys as first-class content ("Consolidated from three tracking tools, completed 2025").

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| One active journey per capability | A capability undergoing two simultaneous independent changes cannot model both | Milestones cover phased work within one story; if truly independent, that is a signal the capability should be decomposed |
| Plan-only moves | Completing a move leaves model and record temporarily inconsistent until the architect executes the reparent | The journey UI can link to the existing operations; signals (184) can flag the mismatch |
| Catalog refs for future apps | Capturing a journey to a not-yet-decided app requires creating a component first | Component creation is one dialog; `Planned` realisation level exists for exactly this |
| Manual progress | Progress can go stale | Progress carries its event timestamp; staleness is visible in the drawer |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
