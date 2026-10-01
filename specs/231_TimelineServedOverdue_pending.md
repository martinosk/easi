# 231 — Timeline Reads Served Overdue

> **Status:** pending
> **Depends on:** 229 (shared-kernel quarter rule), 197 (journey timeline)
> **Roadmap alignment:** `SD8 / H2-6` — completes decision D10 of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) as refined by spec 229 decision 1

---

## Problem Statement

After spec 229, "overdue" has two implementations: the shared-kernel quarter rule on the server, used by Home, and `isPastDue` in the frontend timeline model (spec 197 rule 8), evaluated in the browser's local time. They agree today by construction and will drift the first time either changes, and they already disagree in the hours around a quarter boundary for any user not on UTC. Home's "1 journey overdue" and the timeline's count for the same journeys can differ.

This slice makes Architecture Direction serve the overdue state of every journey and milestone it returns, computed with the shared quarter rule, and makes the timeline render the served value. The frontend keeps quarter arithmetic only for laying out columns.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect / Steward** | See the same overdue journeys and milestones on the timeline as on Home. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: The timeline shows the served overdue state

  Scenario: Overdue journey on the timeline
    Given today is 2026-09-29 and an in-flight journey targets Q2 2026
    When a user opens the timeline
    Then the journey is marked overdue
    And the summary counts it among journeys overdue

  Scenario: Overdue milestone on the timeline
    Given a planned journey has a milestone targeting Q1 2026 that is not done
    When a user opens the timeline
    Then the milestone is marked overdue
    And the summary counts it among milestones overdue

  Scenario: Home and timeline agree
    Given a domain's journeys
    When its domain architect compares Home's Journey Health for that domain with the domain's timeline
    Then the journeys overdue and milestones overdue counts are the same

  Scenario: Quarter boundary follows the server
    Given it is 2026-10-01 00:30 UTC and the user's browser is at UTC−5 (2026-09-30)
    And an in-flight journey targets Q3 2026
    When the user opens the timeline
    Then the journey is marked overdue, as on Home
    And the timeline's current-quarter column is Q4 2026

  Scenario: Journey response carries overdue
    When a client reads capability journeys
    Then every journey and every milestone carries an overdue flag
```

---

## Business Rules & Invariants

1. **Served, not computed** — `GET /capability-journeys` returns `overdue` on each journey and each milestone, and the collection returns the `currentPeriod` it evaluated against; both computed at request time with the shared-kernel quarter rule; never stored (spec 197 rule 8 holds).
2. **Same meaning as spec 197** — journey overdue by its own target period and status; milestone overdue by its own target period and status; undated and done never overdue. Only journeys in motion are marked; terminal journeys carry `overdue: false`.
3. **The frontend never evaluates overdue** — the timeline model reads the flags; `isPastDue`, `isJourneyOverdue` and `isMilestoneOverdue` are removed. The current-quarter column uses the served `currentPeriod` rather than the browser's date.
4. **Architecture Direction's `TargetPeriod` uses the shared rule** — its `Before` delegates to, or is replaced by, the shared-kernel quarter, so the backend holds one definition.

---

## Acceptance Criteria

- [ ] The capability-journeys DTO carries `overdue` on journeys and milestones and `currentPeriod` on the collection, with handler tests for overdue, current quarter, done, undated and terminal journeys
- [ ] Architecture Direction evaluates overdue only through the shared-kernel rule; no other backend overdue computation exists (grep test or review note in the spec's sign-off)
- [ ] The timeline model derives row and summary overdue state from the served flags and the current column from `currentPeriod`; its tests use served flags
- [ ] No frontend code compares a target period with the current date to decide overdue
- [ ] The frontend journey types and the MSW mocks include the new fields; `npm run build` type-checks tests
- [ ] Every BDD scenario has at least one corresponding test
- [ ] Swagger regenerated
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Architecture Direction owns the journey read and the flags. The frontend `business-domains` timeline consumes them. No other context changes.

### Domain Model

No aggregate changes. The journey read model's DTO mapping evaluates the shared quarter rule for each journey and milestone at request time.

### API Surface

`GET /capability-journeys`: journeys and milestones gain `overdue` (boolean); the collection gains `currentPeriod` (`{year, quarter}`). Additive.

### Persistence

None.

### Frontend

`timelineModel.ts` reads `overdue` and `currentPeriod`; `period.ts` keeps `periodRank`, `comparePeriods` and formatting for column layout, and `currentTargetPeriod` is removed if no caller remains.

---

## Design Decisions

1. **Architecture Direction serves the flag, not the stewardship context** — journeys are Architecture Direction's; the timeline reads their state from their owner. Both contexts evaluate the same shared-kernel rule (spec 229 decision 1).
2. **Serve `currentPeriod` too** — a timeline that marks items overdue by the server's quarter but draws its "now" column by the browser's would contradict itself at a quarter boundary.
3. **Flags, not a server-computed summary** — the timeline filters and groups journeys in the browser (domain lens, spec 197); a summary served for the unfiltered collection would not match what is shown.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| UTC quarter on the timeline | A user west of UTC sees the next quarter begin a few hours early | Same on Home; one consistent clock beats two correct-looking ones that disagree |
| A time-dependent response | Caching the journeys response across a quarter boundary would serve stale flags | Responses are not cached across requests today; TanStack Query refetches on focus |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
