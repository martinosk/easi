# 169 — Discover: Candidate Themes

> **Status:** pending
> **Depends on:** [167 — Direction on an Enterprise Capability](167_Direction_Aggregate_Capture_done.md)

---

## Problem Statement

The architecture group has ~1300 physical capabilities across six business domains, with duplication baked in by history. The group regularly looks at a cluster of physicals and asks "is this a real cross-domain theme?" — and that conversation happens in side channels. The cluster, the question, and the eventual decision are invisible to anyone outside the room.

After 167 shipped, an architect can already create an Enterprise Capability and capture a Direction on it. What is still missing is the **pre-decision artifact**: a shared, visible record of *"we are considering this cluster; we have not decided yet,"* and — equally important — the record that *"we looked at this cluster and chose not to act,"* so the same cluster does not keep re-surfacing as a fresh idea.

This slice introduces **Discovery Candidates**: a queue of cross-domain themes the group is considering. A Candidate is resolved by recording what the group decided — linked to an Enterprise Capability, linked to a Direction, or rejected with a reason. Resolution is bookkeeping on the Candidate; the Enterprise Capability or Direction itself is created through its existing flow.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect (in session)** | Browse open candidates, see what each spans, resolve on the spot — record the outcome the group reached. |
| **Enterprise Architect (between sessions)** | Add a candidate when one surfaces in conversation, so it lands on the next session's agenda. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Discover candidate themes

  Scenario: Browsing open candidates
    Given the system has open Discovery Candidates
    When I open Discover
    Then I see every open Candidate with its name, source physical capabilities, and the domains involved
    And resolved and rejected Candidates are not in the view

  Scenario: Adding a candidate manually
    Given I am an architect with write access
    When I create a Candidate by name and attach two or more physical capabilities from different domains
    Then the Candidate appears in Discover with status open

  Scenario: A single-domain Candidate is not a cross-domain theme
    Given I attempt to create a Candidate with sources all in the same business domain
    Then the system rejects creation with a clear error
    And no Candidate is created

  Scenario: Resolving a Candidate by linking it to an Enterprise Capability
    Given I am viewing an open Candidate
    And an Enterprise Capability exists that captures the group's decision on this cluster
    When I resolve the Candidate by linking it to that Enterprise Capability
    Then the Candidate's status becomes resolved-as-enterprise-capability
    And the link to the Enterprise Capability is preserved for audit
    And the Candidate no longer appears in the default Discover view

  Scenario: Resolving a Candidate by linking it to a Direction
    Given I am viewing an open Candidate
    And a Direction exists that captures the group's decision on this cluster
    When I resolve the Candidate by linking it to that Direction
    Then the Candidate's status becomes resolved-as-direction
    And the link to the Direction is preserved for audit
    And the Candidate no longer appears in the default Discover view

  Scenario: Rejecting a Candidate
    Given I am viewing an open Candidate
    When I reject it with a reason
    Then the Candidate's status becomes rejected
    And the rejection reason is preserved for audit
    And the Candidate no longer appears in the default Discover view

  Scenario: An open Candidate's source becomes stale
    Given a Candidate references a physical capability that has since been deleted
    When I view the Candidate
    Then the missing reference is marked stale
    And the Candidate otherwise renders normally
    And I can still resolve or reject the Candidate

  Scenario: A read-only user sees Candidates but cannot resolve them
    Given I have read-only access
    When I fetch a Candidate
    Then the response carries no HATEOAS affordances for create, resolve, or reject
    And the UI surfaces no actions for those operations
```

---

## Business Rules & Invariants

1. **A Candidate spans ≥2 business domains.** Enforced at creation; single-domain creation is rejected.
2. **A Candidate has status `open`, `resolved-as-enterprise-capability`, `resolved-as-direction`, or `rejected`.** Each transition is a discrete past-tense event.
3. **`open` is the only status from which a Candidate can be resolved or rejected.** Terminal statuses are mutually exclusive and cannot be reopened — to reconsider, create a new Candidate.
4. **Resolution records a reference to an existing aggregate by ID** — an Enterprise Capability ID or a Direction ID. The Candidate does not create the referenced aggregate; that aggregate is created through its own existing flow. The reference is informational, not a structural relationship; deleting the referenced aggregate does not invalidate the Candidate's recorded resolution.
5. **Sources are set at creation and immutable thereafter.** To change sources, reject and create a new Candidate.
6. **Source references are eventually consistent with `capabilitymapping`.** Deleted sources surface as stale and do not block reading or resolution.
7. **Authorisation matches `architecturedirection`** (per spec 167). Create, resolve, and reject require the architect permission; read follows the bounded context's read permission. HATEOAS advertises operations only when authorised.

---

## Acceptance Criteria

- [ ] An architect can create a Candidate with name, optional description, and ≥2 sources spanning ≥2 domains; single-domain creation is rejected with a clear error
- [ ] An architect can resolve an open Candidate by linking it to an existing Enterprise Capability; the Candidate becomes `resolved-as-enterprise-capability`
- [ ] An architect can resolve an open Candidate by linking it to an existing Direction; the Candidate becomes `resolved-as-direction`
- [ ] An architect can reject an open Candidate with a reason; the Candidate becomes `rejected` and the reason is preserved
- [ ] The Discover view shows only `open` Candidates
- [ ] A Candidate's recorded resolution survives deletion of the referenced aggregate (it may render as a stale link, but the audit fact remains)
- [ ] Stale source references render with an explicit indicator and do not block resolution or rejection
- [ ] HATEOAS on a Candidate response advertises resolve / reject only when the caller is authorised; the UI gates its action buttons on those links
- [ ] Read-only users can view Candidates but receive no resolution affordances
- [ ] Every BDD scenario above has at least one corresponding test
- [ ] Every modified file scores 10.0 in CodeScene per `easi-codehealth`

---

## Architecture

### Ownership
`DiscoveryCandidate` lives in `architecturedirection`, alongside `Direction`. Same authorisation and audit patterns as 167.

### Domain Model
Aggregate: identity, name, optional description, set of source physical capability IDs (≥2, spanning ≥2 domains), status, and on resolution either an Enterprise Capability ID or a Direction ID (or, on rejection, a reason). Sources are immutable after creation. Events: `CandidateOpened`, `CandidateResolvedToEnterpriseCapability`, `CandidateResolvedToDirection`, `CandidateRejected`.

### API Surface
Discovery Candidates are a top-level resource under `architecturedirection`. List defaults to `open`. Create / resolve-as-enterprise-capability / resolve-as-direction / reject are advertised via HATEOAS based on caller authorisation and current status. Exact route shapes settled during implementation per `easi-api-standards`.

### Persistence
Event-sourced, same pattern as `Direction`. No cross-context writes — resolution stores an ID reference only.

### Frontend
A new Discover surface under the architecture-direction area. Open candidates list with name, sources, and involved domains; the detail surface exposes the three resolution actions. The architect may pick an existing EC/Direction via a search, or — as a UX convenience — deep-link from a Candidate into the EC-creation or Direction-capture flow with sources pre-filled, then come back and record the resolution. The deep-link is a frontend convenience; the behavioural contract is only the Candidate's recorded outcome.

### Cross-Context Integration
`architecturedirection` subscribes to `capabilitymapping` physical capability events for stale-source detection (same wiring already used by `Direction`). No write commands flow to other contexts; resolution is local bookkeeping.

---

## Design Decisions

1. **Candidate as its own aggregate.** A pre-decision artifact has its own lifecycle (browse → resolve), its own authorisation surface, and produces the rejected-cluster audit record that embedding into EC or Direction would lose.
2. **Resolution is bookkeeping, not orchestration.** The Candidate does not create the EC or Direction it resolves to — those are created through their own existing flows. This avoids cross-context atomic-promotion machinery for a problem that does not need it and keeps 167's published-language contract untouched.
3. **Sources immutable after creation.** Removes a class of edge cases around the ≥2-domain invariant and stale references mid-edit; reject and re-create is one action.
4. **Resolution reference is informational.** Deleting the referenced EC or Direction does not invalidate the Candidate's recorded outcome; the audit fact "we decided this cluster became X" is more valuable than a guaranteed live link.
5. **Manual creation only in this slice.** Automated detection (cross-domain signal mining, ranking) is a separate design problem; the aggregate is shaped to accept an automated opener later without re-modelling.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Resolution is bookkeeping | The architect must navigate to the EC/Direction flow separately and come back to record the resolution | Frontend can deep-link with sources pre-filled; the behavioural contract stays simple |
| Resolution reference is informational | A Candidate may end up linked to a deleted aggregate | Read-side renders the link as stale; the audit fact remains |
| Sources immutable | Editing a typo in the source list requires reject + re-create | Acceptable; source-list edits are rare and the audit value of immutability is high |
| Manual creation only | Architects do all the curation | Aggregate is forward-compatible with an automated opener; that is a future spec |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
