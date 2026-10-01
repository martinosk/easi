# 229 — Home: Journeys in Motion

> **Status:** pending
> **Depends on:** 228 (home, portfolio scope), 197 (overdue definition), 211 (journey kinds and milestones)
> **Roadmap alignment:** `SD8 / H2-6` — slice E of [`docs/specs/personal-home.md`](../docs/specs/personal-home.md) (decision D10)

---

## Problem Statement

The third question the home answers is **what is in motion**. Journeys and their milestones are planned in Architecture Direction and seen on the Business Domains timeline (spec 197), one domain at a time. A steward or owner has no view of the journeys touching their part of the landscape across domains, and no count of how many have slipped.

"Overdue" is computed today only in the frontend timeline model (spec 197 rule 8). The home needs it on the server, and a second, independent computation would drift from the timeline's (design doc D10). This slice defines overdue once, in the shared kernel, uses it for the home's **Journey Health** section, and leaves the timeline's switch to the served value to spec 231.

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Journey in motion** | A journey with status `planned` or `in-flight` (spec 197's *active*). |
| **Journeys in scope** | Journeys in motion whose capability is a portfolio capability, or whose source or target application is a portfolio application (spec 228 rule 3). |
| **Current quarter** | The calendar quarter containing the request time, in UTC. |
| **Overdue** (spec 197 rules 2–4) | A journey or milestone with a target period before the current quarter and a status other than `done`. Undated items are never overdue. A journey's overdue state is its own target period's; its milestones are overdue or not each on their own, exactly as the timeline shows them. |
| **Slipping** | A journey in motion that is overdue or has at least one overdue milestone. Used for ordering only. |
| **Quarter rule** | The shared-kernel definition of a quarter, the current quarter and overdue; the single definition used by every backend consumer. |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Steward / Domain architect** | See which journeys touching their domains are moving, planned or late. |
| **Application owner** | See the journeys that migrate, consolidate or carve out their application. |
| **Architect without anchors** | See every journey in motion in the tenant, overdue first. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Journeys in motion on Home

  Background:
    Given today is 2026-09-29, in the quarter Q3 2026

  Scenario: Journey Health counts
    Given the caller's scope has 2 in-flight and 5 planned journeys
    And 1 of them is overdue and 3 of their milestones are overdue
    When they open Home
    Then Journey Health shows "2 in flight", "5 planned", "1 journey overdue" and "3 milestones overdue"

  Scenario: Overdue by the journey's target period
    Given an in-flight journey targeting Q2 2026
    Then it is overdue

  Scenario: An overdue milestone does not make its journey overdue
    Given a planned journey targeting Q4 2026 with a milestone "Pilot" targeting Q1 2026 in status in-flight
    Then the journey is not overdue
    And "Pilot" is an overdue milestone
    And the journey's row names "Pilot" as overdue

  Scenario: A done milestone is never overdue
    Given a journey whose only past-dated milestone is done
    And whose target period is Q3 2026
    Then it is not overdue

  Scenario: The current quarter is not overdue
    Given an in-flight journey targeting Q3 2026
    Then it is not overdue

  Scenario: Undated journeys are never overdue
    Given a planned journey with no target period and no dated milestones
    Then it is not overdue

  Scenario: Rows are ordered by urgency
    When the caller opens Home
    Then Journey Health lists up to 5 journeys: overdue first, then other slipping journeys, then in flight, then planned
    And within each group by target period, undated last, then capability name

  Scenario: A row describes the journey
    Given the in-flight migration journey on "Customer Data Management" targeting Q4 2026 with 2 of 5 milestones done
    Then its row shows "Customer Data Management", "Migration", "in flight", "Q4 2026", "2 of 5 milestones" and the capability's domains

  Scenario: Journeys reach the caller through their applications
    Given "Mette Gram" owns "CRM" and holds no other anchor
    And a planned consolidation journey on "Billing" in "Finance" lists "CRM" as a source application
    When they open Home
    Then Journey Health lists that journey

  Scenario: Finished and abandoned journeys are not in motion
    Given a journey in the caller's scope is completed
    Then Journey Health does not count or list it

  Scenario: Tenant scope
    Given an architect without anchors
    When they open Home
    Then Journey Health covers every journey in motion in the tenant

  Scenario: Empty scope
    Given a stakeholder without anchors
    When they open Home
    Then Journey Health is not shown

  Scenario: Opening a journey
    When the caller selects a journey row
    Then the Business Domains board opens with that capability's drawer
    When they select "Open timeline"
    Then the Business Domains timeline opens

  Scenario: Nothing in motion
    Given the caller's scope has no journey in motion
    When they open Home
    Then Journey Health reads "No journeys in motion"
```

---

## Business Rules & Invariants

1. **In motion** — status `planned` or `in-flight`; `done` and `abandoned` journeys are neither counted nor listed.
2. **Scope** — journeys whose capability is a portfolio capability, or whose source or target application is a portfolio application; `tenant` scope covers all; `empty` scope has no Journey Health section.
3. **Overdue** — per the Ubiquitous Language, evaluated at request time against the current quarter in UTC; never stored.
4. **One definition** — the quarter rule lives in the shared kernel and is the only backend implementation of overdue; the stewardship read side calls it with the cached periods and statuses.
5. **Counts** — over journeys in scope: in flight, planned, journeys overdue, milestones overdue; the same four quantities, with the same meanings, as the timeline's summary (spec 197). Overdue journeys are also counted under their status.
6. **Rows** — at most 5, ordered per the scenario; each carries journey id, capability id and name, kind, status, target period, milestones done and total, whether the journey is overdue, the number of overdue milestones and the first overdue milestone's label (by milestone order), and the capability's effective domains.
7. **Section follows permission** — the section is present only when the caller holds `architecture-direction:read`.
8. **Caches only** — journeys and milestones are read from a local, event-fed, backfilled cache (spec 209).

---

## Acceptance Criteria

- [ ] The shared-kernel quarter rule (quarter value, current quarter from a time, `Before`, overdue for a dated or undated item with a status) has a table test covering the scenarios above and the quarter boundary (23:59 UTC on 30 Sep vs 00:00 UTC on 1 Oct)
- [ ] `GET /api/v1/home` carries a `journeys` section with counts and rows per rules 1–7
- [ ] Journey scope reaches journeys through capability and through source and target applications; tenant and empty scopes behave per rule 2
- [ ] The journey cache handles every event listed under Persistence; its backfill migration seeds journeys and milestones; an integration test runs it against a seeded tenant
- [ ] Home renders Journey Health with the four counts, the rows, "Open timeline" and the empty message, with Mantine primitives
- [ ] Journey rows open `/business-domains?capability={id}`; "Open timeline" opens `/business-domains?presentation=timeline`
- [ ] Every BDD scenario has at least one corresponding test
- [ ] `docs/architecture/Stewardship.md` and `docs/backend/cross-context-events.md` list the journey subscriptions
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

The `stewardship` read side gains a journey cache and the Journey Health query. The shared kernel gains the quarter rule. Architecture Direction is unchanged in this slice (spec 231 moves it onto the rule).

### Domain Model

- **Quarter rule** (shared kernel): a `Quarter` value (year, quarter 1–4), `CurrentQuarter(time)` in UTC, `Before`, and `IsOverdue(period *Quarter, done bool, now time.Time)` for a single journey or milestone. Pure, no dependencies.
- **Journey Health query** (stewardship): journeys in scope → counts and rows (rules 1–6).

### API Surface

`GET /api/v1/home` gains a `journeys` section: `{ inFlight, planned, overdueJourneys, overdueMilestones, items: [...] }`, omitted per rule 7 and for `empty` scope.

### Persistence

New tables in the `stewardship` schema, tenant-scoped with RLS, seeded by a `backfill` migration from Architecture Direction's journey tables:

| Cache | Content | Fed by |
|-------|---------|--------|
| Journey cache | id, capability id, kind, status, target period, source application ids, target application id | `JourneyPlanned`, `JourneyStarted`, `JourneyCompleted`, `JourneyAbandoned`, `JourneyDetailsUpdated`, `JourneySourceApplicationsChanged`, `CapabilityDeleted` |
| Milestone cache | journey id, milestone id, label, target period, status, position | `JourneyMilestoneAdded`, `JourneyMilestoneUpdated`, `JourneyMilestoneRemoved`, `JourneyMilestonesReordered` |

Events after `JourneyPlanned` carry no capability id; the cache keys on the journey id.

### Frontend

A `JourneyHealth` section in the `home` feature: count badges, legend-free rows (status badge, overdue marker, target period, milestone progress, domains), "Open timeline", empty message. Rendered when the response carries `journeys`.

### Cross-Context Integration

| Direction | Events | Purpose |
|-----------|--------|---------|
| Architecture Direction → Stewardship | every `Journey*` event except `JourneyProgressUpdated` | Journey and milestone caches |
| Capability Mapping → Stewardship | `CapabilityDeleted` (already subscribed) | Drop journeys of a deleted capability |

---

## Design Decisions

1. **The quarter rule lives in the shared kernel** — overdue is needed by Architecture Direction (the timeline, spec 231) and by the stewardship read side; neither may import the other, and a published language carries constants, not behaviour. A small, pure shared-kernel value is the one place both can use. This refines design doc D10 ("the read side becomes the single definition"), which would have had the timeline read overdue from the stewardship context. Alternatives: stewardship owns it and Architecture Direction serves the timeline from a stewardship endpoint (rejected — the timeline would depend on a read side for a fact about its own journeys); each context implements it (rejected — the drift D10 exists to prevent).
2. **UTC defines the current quarter** — the server has no user time zone; the frontend used browser local time. Once spec 231 lands the timeline shows the server's value, so both agree; the only visible difference is the first hours of a quarter for users east or west of UTC.
3. **Journeys reach the caller through applications too** — an owner cares about the journey that retires their application, even when the journey's capability is in someone else's domain.
4. **The timeline's four counts, not the mockup's three** — the timeline already distinguishes overdue journeys from overdue milestones (spec 197); Home uses the same two so a number on Home is the number on the timeline for the same scope. Overdue is a property, not a status, so overdue journeys are also counted under their status. Alternative: one "overdue" badge meaning "overdue or with an overdue milestone" (rejected — a third meaning of overdue next to the timeline's two).
5. **No journey name** — journeys have no name of their own; the mockup's "Digital Customer Journey" is illustrative. The row leads with the capability and the kind.
6. **Rows link to the capability drawer** — the journey is planned and edited there; the timeline is one click further through "Open timeline".

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Quarter rule in the shared kernel | A domain rule outside any context | It is a calendar definition with one clause; the kernel guard keeps it free of context imports |
| Until spec 231, the timeline still computes overdue locally in browser time | Home and timeline can disagree in the first hours of a quarter | Spec 231 is small and follows directly |
| A second copy of journeys in the stewardship schema | Storage and backfill cost | Journeys number in the hundreds per tenant |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
