# 168 — Discover: Candidate Themes with Two-Path Promotion

> **Status:** pending
> **Depends on:** [166 — Logical Capability Rename](166_LogicalCapability_Rename_pending.md), [167 — Direction on a Logical Capability](167_Direction_Aggregate_Capture_pending.md)
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md)

---

## Problem Statement

The architecture group has ~1300 physical capabilities scattered across six business domains, with duplication baked in by history. Spotting which clusters of duplicates should *become* something — a consolidation, a logical grouping, or simply a deliberate "no, leave it alone" decision — is the first activity in the upstream of every direction. Today that activity happens in conversations and shared spreadsheets; the candidates and the decisions on them are not visible in the tool, so the group cannot work asynchronously and stakeholders cannot see what is being considered.

This slice introduces **Discovery Candidates** — proposed themes that might span multiple physical capabilities across multiple domains. A candidate is a placeholder for a *decision the group has not yet made*. Each candidate can resolve in one of three ways:

1. **Promote to a physical-consolidation Direction** — the group decided to actually merge.
2. **Promote to a Logical Grouping** — the group decided the cluster is a useful abstraction worth labelling, with or without consolidation intent.
3. **Reject** — the group looked and decided the cluster isn't a real theme.

The two-path promotion is the heart of the conceptual model: the same insight can resolve into either a change to physical reality (Direction) or a change to how we *talk* about reality (Logical grouping).

After this slice, an architect lands on Discover and sees the open consolidation-shaped questions the group still has to answer, and can move each one toward its resolution without leaving the surface.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect (in a working session)** | A surface to browse open candidates, see what each one spans, and resolve it on the spot — promote to Direction, promote to Logical Grouping, or reject. |
| **Enterprise Architect (between sessions)** | Add a candidate manually when one surfaces in a conversation, so it lands on the next session's agenda without a side-channel note. |
| **Domain Owner (consulted on a candidate)** | See what's being considered for their domain so they can react before the group converges. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Discover candidates and two-path promotion

  Scenario: Browsing open candidates
    Given the system has open Discovery Candidates
    When I open Discover
    Then I see every open Candidate
    And rejected and promoted Candidates are not in the default view
    And I can see for each Candidate: name, the physical capabilities it spans, and the domains involved

  Scenario: Adding a candidate manually
    Given I am an architect with write access
    When I create a Candidate by name and attach two or more physical capabilities from different domains
    Then the Candidate appears in Discover in open status

  Scenario: A candidate that does not span multiple domains is not a consolidation candidate
    Given I attempt to create a Candidate with sources all in the same business domain
    Then the system tells me a Candidate must span at least two domains
    And no Candidate is created

  Scenario: Promoting to a physical-consolidation Direction
    Given I am viewing an open Candidate
    When I promote it to a Direction
    Then a new Direction is created with type consolidate, pre-filled with the Candidate's sources
    And the Candidate is recorded as promoted-to-direction
    And opening the new Direction shows a back-reference to the originating Candidate

  Scenario: Promoting to a Logical Grouping
    Given I am viewing an open Candidate
    When I promote it to a Logical Grouping
    Then a new Logical Capability is created and the Candidate's sources are mapped to it
    And the Candidate is recorded as promoted-to-grouping
    And opening the new Logical Capability shows a back-reference to the originating Candidate

  Scenario: Rejecting a candidate
    Given I am viewing an open Candidate
    When I reject it with a reason
    Then the Candidate is recorded as rejected
    And it no longer appears in the default Discover view
    And the rejection is preserved for audit

  Scenario: A reader sees candidates but cannot resolve them
    Given I have read-only access
    When I open Discover
    Then I see the open Candidates
    And I cannot create, promote, or reject any Candidate

  Scenario: A candidate's source becomes stale
    Given a Candidate references a physical capability that has been deleted
    When I view the Candidate
    Then the missing reference is marked stale
    And the Candidate otherwise renders normally
```

---

## Business Rules & Invariants

1. **A Discovery Candidate spans two or more business domains.** A single-domain candidate is not a consolidation candidate by definition; the system rejects creation in that case.
2. **A Candidate references two or more physical capabilities by ID.** References are eventually consistent with `capabilitymapping`; deleted sources surface as stale but do not block reading or resolution.
3. **A Candidate has a status:** `open`, `promoted-to-direction`, `promoted-to-grouping`, or `rejected`. Status transitions are recorded as discrete past-tense events.
4. **`open` is the only status from which a Candidate can be resolved.** The three terminal statuses are mutually exclusive; once a Candidate is promoted or rejected it cannot be reopened. To reconsider, create a new Candidate.
5. **Promotion creates the resolved aggregate atomically with the status transition.** Promoting to a Direction emits both the Candidate's promotion event and the Direction's creation in the same logical operation; the same applies to grouping promotion. The system MUST NOT leave a Candidate in a half-promoted state.
6. **Promotion is idempotent at the user level.** A double-click does not produce two Directions or two Logical Capabilities. The handler establishes promotion intent once.
7. **The originating Candidate is recorded as a back-reference on the resolved aggregate.** A Direction or Logical Capability that came from a Candidate carries the Candidate's identifier so the audit trail "this Direction came from that Candidate" is queryable.
8. **The Candidate body carries a name, a description (optional), and a set of source physical capabilities.** It does not carry a proposed target domain, an application, or any other Direction-shaped fields — those are decided at promotion time, not at candidate-creation time.
9. **Manual creation is the entry point in this slice.** Automated detection (cross-domain similarity, shared application signal, etc.) is a separate concern; the aggregate exists today, the *opening* of candidates from signals is a future spec.
10. **Authorisation is gated.** Create, promote, and reject require an architect-level permission consistent with `architecturedirection`'s existing scheme (settled in spec 167). Reading a Candidate requires read access to the bounded context.

---

## Acceptance Criteria

- [ ] An architect can create a Candidate manually with name, optional description, and two or more source physical capabilities spanning two or more domains
- [ ] The system rejects single-domain Candidate creation with a clear error
- [ ] An architect can promote an open Candidate to a Direction; the resulting Direction starts in `draft`, has type `consolidate`, and carries the Candidate's sources pre-filled; the Candidate is recorded as `promoted-to-direction`
- [ ] An architect can promote an open Candidate to a Logical Grouping; the resulting Logical Capability has the Candidate's sources mapped to it; the Candidate is recorded as `promoted-to-grouping`
- [ ] An architect can reject an open Candidate with a reason; the Candidate is recorded as `rejected`
- [ ] Promotion is atomic: a system failure mid-promotion does not leave a Candidate in a half-resolved state
- [ ] Promotion is idempotent: a repeated promote command on the same Candidate does not create a second Direction or Logical Capability
- [ ] The Discover view shows every `open` Candidate; promoted and rejected ones are excluded by default but accessible via a filter
- [ ] A Candidate's resolved aggregate (Direction or Logical Capability) carries a back-reference to the originating Candidate
- [ ] HATEOAS affordances on a Candidate response advertise the three resolution operations (promote-to-direction, promote-to-grouping, reject) only when the calling user is authorised
- [ ] Stale source references render with an explicit indicator
- [ ] Read-only users see the Candidate list and detail but cannot resolve
- [ ] All BDD scenarios above have at least one corresponding test
- [ ] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

The `DiscoveryCandidate` aggregate lives in the `architecturedirection` bounded context, alongside `Direction`. The same authorisation, audit, and event-publishing patterns apply.

### Domain Model

The `DiscoveryCandidate` aggregate carries: an identity, a name, an optional description, a set of source physical capability IDs (≥2, spanning ≥2 domains), a status, and on resolution a reference to the produced Direction or Logical Capability. Status transitions are individual past-tense events. Source set membership is mutable while the Candidate is `open`; locked once resolved.

### API Surface

Discovery Candidates are exposed as a top-level resource under `architecturedirection`. The contract obligations are: list-open is the default view; create / promote / reject are operations advertised via HATEOAS based on caller authorisation and current status; resolution operations land their results in their respective aggregates (Direction or Logical Capability) and update the Candidate atomically. Exact route shapes are settled during implementation per the API standards skill.

### Persistence

Event-sourced. Resolution events carry the produced aggregate's identifier so projections can render the back-reference and so consumers downstream know which Candidate sourced which Direction or Logical Capability.

### Frontend

A new Discover surface lands under the architecture-direction area of the UI. The mockup at `mockups/architecture-direction.html` shows a matrix layout (domains × candidate themes) — that is one validated visualisation; whether it is the right one at production scale is settled during implementation, with the constraint that a user must be able to scan the open queue and resolve a Candidate without leaving the view.

The two-path promotion is exposed as two clear actions on a Candidate's detail; the rejection action is a third; cancelling closes without state change.

### Cross-Context Integration

`architecturedirection` subscribes to `capabilitymapping` physical capability events to validate Candidate sources at creation time and to detect stale references afterward. Promotion to a Direction is internal to `architecturedirection`. Promotion to a Logical Grouping creates a Logical Capability — the spec assumes the Logical Capability creation API is reachable from `architecturedirection` either directly (in-process call) or via a published-language command; the integration shape is settled during implementation.

---

## Design Decisions

1. **DiscoveryCandidate as its own aggregate.** Per the DDD memo. A Candidate is a distinct artifact with its own lifecycle (browse → resolve), separate authorisation surface, and an explicit historical record of "we considered this and decided X." Embedding the concept in Direction or Logical Capability would lose the audit trail for rejected candidates.

2. **Two-path promotion, not three or four.** The conceptual model identifies exactly two productive resolutions plus a rejection. A `stay` resolution is not offered as a fourth path because a deliberate "no change" already maps to the `stay` Direction type — promote to Direction with type `stay`. Alternative considered: a `confirm-no-action` resolution as a fourth path — rejected because it duplicates Direction's `stay` type.

3. **Manual creation only in this slice.** Automated candidate detection is valuable but is its own design problem (signal sources, ranking, false-positive handling). Bundling it would balloon the slice past the user's "validate as we go" preference. The aggregate is shaped so an automated source can land in a later spec without re-modelling.

4. **At-least-two-domain rule at the aggregate level.** A consolidation candidate that doesn't span domains isn't a consolidation candidate. Enforcing the rule at the aggregate keeps Discover's invariant honest and prevents the queue from filling with single-domain noise.

5. **Atomicity of promotion.** Promotion creates a downstream aggregate; partial failure would leave the system in an awkward state. Implementation can use either a single transactional handler (if both aggregates live in the same store) or a saga / event-driven settlement; the choice is settled during implementation per the EASI consistency patterns.

6. **No proposed target domain on the Candidate body.** The mockup carried `proposedTargetDomain` on the Candidate — useful as a UX hint, but it pre-judges a decision that should be made at promotion time. Carrying it at the aggregate level would conflate "we are still considering this" with "we have already named the target." If a UX hint is useful, it can be a non-aggregate annotation surfaced in the read model, not a body field on the aggregate. Open for revisit during implementation.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Manual creation only | Architects do all the curation | Aggregate is forward-compatible with automated sources; spec for those is a follow-up |
| At-least-two-domain rule | Real candidates that started single-domain (e.g. an architect adding rows over time) need ≥2 domains before saving | Acceptable: a single-domain row is by definition not a consolidation candidate |
| Promotion is terminal | Cannot reopen a rejected Candidate by accident | Acceptable: creating a fresh Candidate with the same name is one click |
| No `proposedTargetDomain` on the aggregate | UX hint loses one signal at the model level | Hint can be surfaced as a read-side annotation; aggregate stays clean |
| Promotion to Logical Grouping creates a Logical Capability cross-context | Cross-context coordination needed | Established EASI pattern (events / commands across contexts); decided at implementation |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
