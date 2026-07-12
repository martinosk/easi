# 180 — TIME Assessment on Capability Realisations

> **Status:** done
> **Depends on:** 118 (TIME suggestions — read-only reference), 179 (Domain Board surfaces)
> **Supersedes:** 119 (TIME Classification & Application Landscape — assessed TIME moves from enterprise-capability level to realisation level)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)

---

## Problem Statement

TIME exists in EASI only as a system-computed suggestion (spec 118) — evidence without a verdict. Architects have no way to record the grade they actually stand behind, so the answer to "is this app Invest or Eliminate for this capability?" still lives in slides.

The mockup settles the granularity: TIME is assessed **per application per capability — on the realisation**. The same application can be Invest for one capability and Eliminate for another; that mixed profile is exactly what drives carve-out thinking. Each assessment records who graded it and when, and goes stale after 12 months so an unrevisited judgement is visibly old.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain Architect** | Grade each realising application per capability in their domain; re-grade when the judgement changes; see the system suggestion as reference while grading. |
| **Enterprise Architect** | See an application's grade profile across the whole landscape (I×n · T×n · M×n · E×n) to spot carve-out candidates and rationalisation targets. |
| **Engineer / Product Manager** | Open a capability and see, per realising app, the architecture group's TIME grade, who assessed it, and whether it is stale. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: TIME assessment on a capability realisation

  Scenario: An architect assesses a realisation
    Given capability "Booking management" is realised by application "Seabook"
    And the realisation has no TIME assessment
    When the architect assesses it as "Migrate"
    Then the capability's app chip on the Domain Board shows an "M" badge
    And the capability drawer shows "Migrate — for this capability", the assessing user, and the assessment date

  Scenario: The system suggestion is shown as reference while assessing
    Given the realisation has a computed TIME suggestion of "Eliminate"
    When the architect opens the assessment control
    Then the suggestion "Eliminate" is displayed as a pre-filled reference the architect can accept or override

  Scenario: Re-assessing replaces the current grade and preserves history
    Given a realisation assessed "Tolerate" in January
    When the architect re-assesses it as "Eliminate"
    Then "Eliminate" is the current grade with the new assessor and date
    And the prior "Tolerate" assessment is reconstructable from the event history

  Scenario: An unassessed realisation says so explicitly
    Given a realisation with no assessment
    Then its app chip carries no TIME badge
    And the drawer shows "unassessed" for that application

  Scenario: An assessment older than 12 months is marked stale
    Given a realisation assessed 13 months ago
    When any user views it in the drawer
    Then the assessment carries a "stale" marker alongside the grade

  Scenario: An application's grades roll up across the landscape
    Given "Seabook" is assessed on four realisations: Invest ×1, Tolerate ×1, Migrate ×1, Eliminate ×1
    When a user views any Seabook app row in the capability drawer
    Then the row shows "Across landscape: I×1 · T×1 · M×1 · E×1"

  Scenario: Inherited realisations are not assessable and carry no badge
    Given capability "Accounts receivable" surfaces "Seabook" as an inherited realisation
    Then the inherited chip shows no TIME badge
    And no assess affordance is offered on the inherited row

  Scenario: Deleting a realisation removes its assessment from all read surfaces
    Given an assessed realisation
    When the realisation is deleted
    Then the assessment no longer appears on any board, drawer, or rollup
    And the assessment events remain in the event store

  Scenario: An architect removes an assessment
    Given an assessed realisation
    When the architect removes the assessment
    Then the realisation presents as unassessed
    And the removal is recorded as a discrete event

  Scenario: Read-only users see assessments but cannot write
    Given a user without the architect permission
    When they fetch a capability's realisations
    Then assessments are visible but the responses carry no assess/remove affordances
```

---

## Business Rules & Invariants

1. **Granularity** — an assessment belongs to exactly one (domain capability, application component) pair for which a **direct** realisation exists. At most one current assessment per pair.
2. **Grades** — `Invest` | `Tolerate` | `Migrate` | `Eliminate`. Nothing else, no null grade (absence of an assessment is the unassessed state).
3. **Provenance** — every assessment records the acting user and the server-side timestamp on the event; both are immutable facts of that assessment.
4. **Rationale is optional**, bounded in length. An assessment is "a per-capability assessment, not a commitment" (mockup legend) — friction is deliberately lower than `StandardApplication`'s required narrative.
5. **Re-assessment replaces**; history is reconstructed from events, never overwritten. Recording the same grade again is a valid re-affirmation (refreshes assessor/date, resets staleness).
6. **Stale** — an assessment whose date is more than 12 months old carries a read-side stale flag. Computed at query time; no event, no job.
7. **Direct realisations only.** Inherited realisations are mechanical `Full` copies deduplicated per ancestor; grading them would be ambiguous when several descendants share the app. Inherited rows surface no grade and no affordance.
8. **Realisation deletion hides the assessment** from every read surface (board, drawer, rollups, signals); events are retained. Deletion never blocks on assessments.
9. **Removal is explicit** — an architect can remove an assessment; recorded as its own past-tense event.
10. **Authorisation** — writes require the `architecture-direction:*` architect permission (167 rule 12); reads follow board read permission. Affordances are HATEOAS-advertised only when authorised.
11. **Suggestions are untouched** — spec 118's computed suggestion remains a separate, read-only value displayed beside the assessment; an assessment never overwrites it.

---

## Acceptance Criteria

- [x] An architect can assess a direct realisation with a grade (+ optional rationale); the current grade, assessor, and date surface in the capability drawer within one refetch
- [x] Direct app chips on the Domain Board render the grade badge (I/T/M/E); unassessed and inherited chips render none
- [x] The assessment control shows the computed suggestion (when one exists) as pre-filled reference
- [x] Re-assessment replaces the current grade; the full assessment history is reconstructable from the event stream
- [x] Assessments older than 12 months carry a stale marker in the drawer
- [x] Each app row in the drawer shows the landscape rollup (count of current assessments per grade for that application)
- [x] Removing an assessment returns the pair to unassessed; recorded as a discrete event
- [x] Deleting the underlying realisation removes the assessment from all read surfaces without blocking
- [x] Write affordances are HATEOAS-gated on the architect permission; read-only users receive none
- [x] Spec 119 is renamed `_superseded` and points here
- [x] Every BDD scenario has at least one corresponding test; every business rule has a unit test
- [x] Every modified file scores 10.0 per `easi-codehealth` (two justified inherents: `time_assessment_read_model.go` 9.38, `time_assessment_reference_projector_test.go` 9.68 — module-level string-heavy profile from the three-reference name cache, same shape as the 10.0 `StandardApplication` benchmark at larger domain surface)

---

## Architecture

### Ownership

`architecturedirection` owns the `TimeAssessment` aggregate (design doc D7): it records an architect judgement, exactly the context's language, with the permission family, stale-reference machinery, and name caches already in place. `capabilitymapping` (realisations) and `architecturemodeling` (components) are referenced read-only. `enterprisearchitecture` (suggestions) is queried read-only for the reference display.

### Domain Model

`TimeAssessment` — intrinsic UUID identity; carries capability ref, component ref, grade VO, optional rationale, assessor, assessed-at. Events: `TimeAssessmentRecorded` (first grade and every re-grade; carries previous grade when replacing, plus the realisation id captured at assess time — `SystemRealizationDeleted` carries only that id, so the read model cascades deletions by it) and `TimeAssessmentRemoved`. Uniqueness per (capability, component) pair is enforced at the command handler via read-model lookup with a DB unique constraint as backstop — the exact `StandardApplication` pattern (170 revision 1). The handler verifies a direct realisation exists for the pair at assess time (every assess, including re-assess) via a gateway backed by capabilitymapping's own read model.

### API Surface

Assessments are exposed in the capability/realisation resource tree with discrete assess/remove operations (no free-form PATCH), plus a bulk query keyed by capability set (or domain) so the board loads assessments for a whole domain in one request, and a per-application rollup query. Shapes settled at implementation per `easi-api-standards`; HATEOAS affordances gate all writes.

### Persistence

Event-sourced per house pattern. Read model: one row per current assessment with denormalised capability/component/assessor names (reference_name_cache pattern; assessor names keyed by email under a new `user` entity type, fed by `UserCreated`), unique (tenant, capability, component). Stale flag computed in the query, not stored. Migration 125 creates the read model; migration 126 backfills the user name cache from `auth.users` and repairs email-fallback assessor names — new caches are always backfilled when introduced.

### Frontend

- `AppChip` (Domain Board): grade badge on direct chips.
- Capability drawer app rows: grade + "for this capability" + assessor/date + stale marker + landscape rollup; assess/re-assess/remove control HATEOAS-gated, suggestion shown as reference.
- New queries follow `easi-frontend-data`; `mutationEffects.ts` invalidates board realisation/assessment queries on write.

### Cross-Context Integration

Subscribes to `capabilitymapping` realisation-deleted, capability-deleted, and capability name events (hide assessment rows, capability name cache), `architecturemodeling` component events (name cache, component-deletion cascade), and `auth` user-created events (assessor name cache). Publishes its events into the published language for read-side consumers within the context (board queries, spec 184 signals). No outbound writes.

---

## Design Decisions

1. **Realisation-level, not EC-level (supersedes 119)** — user decision 2026-07-12; the mockup's per-app-per-capability grading is the whole point (mixed profiles drive carve-outs). EC-level grading (119) also depended on EC linking removed by 172.
2. **Owned by `architecturedirection`, not `capabilitymapping`** — grading is judgement, not structure; keeps architect permissions out of the structural context. Alternative (field on `CapabilityRealization`) rejected: every realisation write would load judgement state, and CM's language stays clean.
3. **Aggregate with intrinsic UUID; pair-uniqueness at handler + DB backstop** — direct application of 170's hard-won revision (stream-collision bug).
4. **Rationale optional** — deliberate divergence from `StandardApplication`'s required narrative: assessments are frequent, low-ceremony judgements ("assessment, not commitment"); a required narrative would suppress adoption. Divergence named per house style.
5. **Direct realisations only** — inherited rows are deduplicated mechanical copies; when two descendants share an app the ancestor's inherited row has no single source assessment. Rejected alternative (surface the source realisation's grade on inherited chips) fails exactly that case.
6. **Stale computed read-side at a fixed 12 months** — mockup constant; an event or configurable threshold adds machinery no scenario asks for.
7. **Suggestion and assessment stay separate values** — evidence vs verdict (118/119 split); overwriting the suggestion would destroy the ability to see divergence.
8. **One published serialised form for TIME grades** — `Invest`/`Tolerate`/`Migrate`/`Eliminate` (capitalised) across suggestion (118) and assessment (180). Discovered during implementation: 118's calculator bypassed its `TimeClassification` VO and serialised UPPERCASE while the frontend type claimed capitalised. Fixed at the model: the VO owns the canonical form, the calculator routes through VO constants, and the frontend parses grades at the boundary (`normalizeTimeGrade`) rather than casting. The two contexts keep separate VOs (judgement vs evidence) but share one published language.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Realisation-level granularity | Many more assessable pairs than EC-level; coverage will be partial | Unassessed is an explicit, honest state; suggestions prefill; rollups work with partial data |
| Rationale optional | "Why is this Eliminate?" may be unanswered | Assessor + date always recorded; the architect is one click away |
| Direct-only assessment | Ancestor cards show unbadged inherited chips | Correct by construction; the direct realisation one level down carries the grade |
| Handler-level uniqueness | Cross-aggregate invariant not aggregate-intrinsic | Established pattern (167/170) + DB unique constraint backstop |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [x] User sign-off
