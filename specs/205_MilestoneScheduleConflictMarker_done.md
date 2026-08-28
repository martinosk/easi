# 205 — Milestone Schedule Conflict Marker

> **Status:** done
> **Depends on:** 196 (milestone reorder), 197 (timeline view)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)

---

## Problem Statement

Spec 196 made the milestone sequence architect-controlled, while every milestone may also carry a structured target period (182 rule 8). The two can disagree: a milestone targeted for Q4 2026 may sit below one targeted for Q1 2027. Today the drawer renders that silently, and a reader cannot tell whether the order or the dates are wrong. The timeline (197) shows the dates only, so the disagreement is invisible on both surfaces.

This slice keeps the plan sequence authoritative (196 decisions 1–3) and surfaces the disagreement as a question, not an error — the 184 signals posture: computed on read, never persisted, gone as soon as the underlying data is fixed.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain / Enterprise Architect** | Notice when the sequence they arranged contradicts the target periods they entered, and fix whichever is wrong. |
| **Engineer / Product Manager / Stakeholder** | Read a milestone list without wondering whether an out-of-date-order row is a mistake. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Milestone schedule conflict marker

  Scenario: A milestone placed after a later-targeted milestone is marked
    Given a journey lists "Seabook read-only" (Q1 2027) above "North Sea corridor migrated" (Q4 2026)
    When any user views the journey's milestones
    Then "North Sea corridor migrated" carries a schedule-conflict marker naming Q1 2027
    And "Seabook read-only" carries no marker

  Scenario: Chronological and undated lists carry no marker
    Given a journey whose dated milestones are listed in ascending period order
    And some milestones have no target period
    When any user views the journey's milestones
    Then no milestone carries a marker

  Scenario: Fixing either side clears the marker
    Given a marked milestone
    When the architect reorders the list into period order, or edits the periods into sequence order
    Then the marker disappears with no further action
```

---

## Business Rules & Invariants

1. **Conflict definition** — a dated milestone is in conflict when any dated milestone above it in the list has a strictly later target period. Undated milestones never conflict and never cause a conflict.
2. **Read-only signal** — the marker is derived on read from the milestone list; nothing is persisted, acknowledged, or dismissed (184 posture).
3. **Sequence stays authoritative** — the marker never reorders, sorts, or blocks a reorder; 196 rules are unchanged.
4. **Every reader sees it** — the marker is presentation, not an affordance; it is not permission-gated.

---

## Acceptance Criteria

- [x] A dated milestone placed below a later-dated milestone shows a marker that names the later period it sits behind
- [x] Chronological lists, undated milestones, and single-milestone lists show no marker
- [x] Reordering into period order (196) removes the marker without any other write
- [x] Every BDD scenario has at least one corresponding test; every business rule has a unit test
- [x] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

Frontend `journeys` feature only. No backend, API, or persistence change.

### Frontend

A pure function over the milestone list yields, per conflicting milestone, the latest earlier-listed period it contradicts. The drawer's milestone row renders a subtle marker with a tooltip explaining the contradiction.

### Cross-Context Integration

None.

---

## Design Decisions

1. **Marker, not sort** — auto-sorting by period would make the 196 reorder a tie-breaker and make drags snap back; 196 decision 1 keeps one intent per gesture. Rejected.
2. **Drawer only** — the timeline already places milestones by quarter, so the contradiction is only visible where the sequence is shown. Extending 184's signals drawer with a journey-level signal was rejected as scope for this slice; 184 can consume the same function later.
3. **Compare against the latest period above, not the row directly above** — a Q4 2026 milestone below Q1 2027 and Q2 2026 rows is still behind Q1 2027; reporting the immediate neighbour would hide that.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Client-side derivation | Any other consumer must recompute | The function is pure and exported from the journeys feature |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (browser spec; no backend change)
- [x] API documentation updated (no API change)
- [x] User sign-off
