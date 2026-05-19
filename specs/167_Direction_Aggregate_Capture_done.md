# 167 — Direction on an Enterprise Capability

> **Status:** done
> **Depends on:** —
> **Conceptual basis:** [`mockups/architecture-direction-model.md`](../mockups/architecture-direction-model.md), [`mockups/architecture-direction-ddd.md`](../mockups/architecture-direction-ddd.md)
>
> **Revision 2026-05-19** — Reopened during user review after a strategic-DDD consult. The per-placement name is *kept* (a Direction is a recorded group decision; its rendering must not silently mutate when an external aggregate is edited), but the field is reshaped:
>
> 1. Renamed `resultingNameHint` → `resultingName` (the "hint" wording was the source of the precedence ambiguity that triggered this review).
> 2. **Defaulted at capture time** from the parent Enterprise Capability's current canonical name. The architect sees one pre-filled input, not a separate "EC name vs hint" decision.
> 3. **Immutable once the Direction transitions to `agreed`**, alongside `type` and `horizon`. Renames of the Enterprise Capability after that point no longer touch what past Directions say they will produce.
> 4. While the Direction is in `draft` or `proposed`, the architect may edit the name freely; the UI may also offer a "refresh from EC" affordance when the EC has been renamed since capture, but no auto-propagation.
>
> The bounded-context split (`architecturedirection` separate from `enterprisearchitecture`) is unchanged and confirmed.
>
> The aggregate has never been deployed; the event payload field is renamed cleanly with no upcaster.

---

## Problem Statement

Today an Enterprise Capability carries no information about whether the architecture group has a direction on it — the decisions live in conversations, slides, and shared documents, not in the tool. Anyone outside the room cannot see the direction without asking an architect. That is the exact alignment-decision friction the model exists to remove.

This slice introduces the **Direction** concept — a structured statement attached to an Enterprise Capability that says: *what the group intends to do here* (consolidate / decompose / stay), *where it is in the group's decision process* (draft / proposed / agreed / rejected), and *why* (a stakeholder-readable narrative). After this slice ships, an individual making a daily decision can open an Enterprise Capability and answer "is there a direction on this?" in five seconds.

Direction is the load-bearing addition. Subsequent slices (Discover, Direction Map, Target Architecture, Open Discussions) are read-side surfaces over the same aggregate. This spec defines the aggregate and the simplest write-side flow: capture, advance, reject, view in context.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect (author)** | Capture and progress a direction on an Enterprise Capability without leaving its detail surface; advance status as the group reaches consensus. |
| **Enterprise Architect (reader)** | See the current direction on any Enterprise Capability they navigate to, with enough context to know whether it is final, under discussion, or still being shaped. |
| **Domain Owner / Product Manager** | When making an investment or design decision involving an Enterprise Capability, see at a glance whether the architecture group has a direction, what it is, and how settled it is. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Direction on an Enterprise Capability

  Scenario: An enterprise capability with no direction shows that explicitly
    Given I am viewing an Enterprise Capability with no Direction
    Then the detail surface shows an explicit "no direction set" state
    And the state is distinguishable from a Direction in draft

  Scenario: An architect captures a draft direction
    Given I am an architect viewing an Enterprise Capability
    When I capture a Direction with a type, one or more source physical capabilities, and a narrative
    Then the Direction is created in draft status
    And it appears on the Enterprise Capability detail surface within five seconds

  Scenario: An architect advances a direction toward agreement
    Given a Direction exists in draft
    When I advance the status to proposed
    Then the Direction's status updates and the change is recorded as a discrete event
    When I advance the status to agreed
    Then the Direction's status updates and the change is recorded as a discrete event

  Scenario: A direction can be rejected from any non-terminal status
    Given a Direction exists in draft or proposed
    When I reject the Direction
    Then the Direction's status becomes rejected and is preserved for audit
    And the Direction no longer presents as the active alignment answer for the Enterprise Capability

  Scenario: A reader sees the current direction in context
    Given I have read-only access to an Enterprise Capability with a Direction
    When I open the Enterprise Capability
    Then I see the Direction's type, status, and narrative
    And I cannot create, advance, or reject the Direction

  Scenario: A direction references physical capabilities that may change
    Given a Direction references a physical capability that has since been deleted
    When I view the Direction
    Then the missing reference is marked as stale
    And the Direction otherwise renders normally

  Scenario: A placement's resulting name defaults from the Enterprise Capability at capture
    Given I am capturing or editing a Direction in draft or proposed
    When I add a placement
    Then the placement's resulting name is pre-filled with the Enterprise Capability's current canonical name
    And I may accept the default or type a different name

  Scenario: Resulting name is frozen when the Direction is agreed
    Given a Direction has transitioned to agreed
    When I attempt to edit any placement's resulting name
    Then the edit is rejected
    And the only path to change the resulting name is reject-and-replace

  Scenario: Renaming the Enterprise Capability does not mutate past Directions
    Given an Enterprise Capability has a Direction in any status with placements whose resulting names were defaulted from the Enterprise Capability's previous name
    When the Enterprise Capability is renamed
    Then the resulting names on the existing Direction's placements are unchanged
    And, for Directions still in draft or proposed, a "refresh from Enterprise Capability" affordance is available but not automatic

  Scenario: At most one active direction per enterprise capability
    Given an Enterprise Capability has an agreed Direction
    When an architect attempts to capture a new Direction on the same Enterprise Capability
    Then the system prevents the second active Direction
    And the architect is offered the path of rejecting the existing Direction first
```

---

## Business Rules & Invariants

1. **An Enterprise Capability has at most one active Direction at a time.** A Direction is *active* if its status is `draft`, `proposed`, or `agreed`. `Rejected` Directions are preserved but not active.
2. **A Direction is its own aggregate**, owned by the new `architecturedirection` bounded context. Its lifecycle is independent of the Enterprise Capability and of any physical capability it references.
3. **A Direction has a type.** One of: `consolidate` (multiple physicals merge into one), `decompose` (one physical splits into multiple), `stay` (explicitly confirmed no change). Type is immutable once the Direction is created — to change intent, reject and capture a new Direction.
4. **Type / source-cardinality invariants:** `consolidate` requires N ≥ 2 source physical capabilities; `decompose` requires exactly 1 source; `stay` requires exactly 1 source. The aggregate enforces these at the command boundary.
5. **A Direction references one or more physical capabilities by ID.** These references are eventually consistent with the `capabilitymapping` context. The Direction does not embed physical capability state; it carries only the references and any local annotations needed for the narrative.
6. **A Direction has a horizon.** One of `now` / `next` / `later`. Horizon is required at creation. Horizon is mutable while the Direction is `draft` or `proposed`; immutable once `agreed`.
7. **A Direction has 0..N target placements.** A placement is a target business-domain reference plus a `resultingName` — the name the resulting physical capability will take in that domain. **Type / placement-cardinality invariants:** `consolidate` requires exactly 1 placement (N physicals merge into 1, which lives in exactly one target domain); `decompose` requires 1 or more placements (one physical splits into N); `stay` carries zero placements. The aggregate enforces these at the command boundary.

7a. **`resultingName` is captured on the Direction, defaulted from the Enterprise Capability, and frozen on `agreed`.** At capture and edit time, the UI pre-fills `resultingName` from the parent Enterprise Capability's current canonical name; the architect may accept the default or override per placement. The value is stored on the Direction aggregate (not resolved through the EC at read time), so the Direction's rendering is a stable record of what the group decided. While the Direction is in `draft` or `proposed`, `resultingName` is mutable; on transition to `agreed` it becomes immutable alongside `type` and `horizon`. To change a `resultingName` after agreement, follow the same reject-and-replace path that applies to `type` changes. Renaming the Enterprise Capability does not retroactively change `resultingName` on any existing Direction — pre-`agreed` Directions may be re-defaulted on explicit user action; `agreed` Directions are never touched.
8. **A Direction has a status workflow:** `draft` → `proposed` → `agreed`, with `rejected` reachable from `draft` or `proposed` (terminal). Forward-only on the agreement axis: a Direction does not transition from `agreed` back to `proposed` — to revisit, reject and replace.
9. **Status transitions are recorded as discrete past-tense domain events**, one event per transition (e.g. an event for proposing, an event for agreeing, an event for rejecting). Replay reconstructs the current status. There is no generic "status changed" event.
10. **A Direction has a stakeholder-readable narrative.** One to two sentences naming what the group decided and why. The narrative is required before a Direction can advance from `draft` to `proposed`.
11. **A Direction's source-capability references can become stale.** When a referenced physical capability is deleted, the Direction surfaces a stale-reference indicator but does not block reading or further status transitions. The architect can edit the source list to remove the stale reference (legal pre-`agreed` only).
12. **Authorisation is gated.** Capture, advance, and reject require an architect-level permission. Reading a Direction follows the same read permission as the underlying Enterprise Capability. The permission family is `architecture-direction:*`, scoped per tenant, mirroring `enterprise-arch:*`.
13. **Direction is a published-language concept of `architecturedirection`.** Other contexts that need to know about Directions (notably the Discover view in slice 168 and the Direction Map in slice 170) subscribe to its events. No context outside `architecturedirection` writes Directions.

---

## Acceptance Criteria

- [x] An architect can create a Direction on any Enterprise Capability with type, source physical capabilities (cardinality matching the type per rule 4), horizon, placements (cardinality matching the type per rule 7; each placement carries a target business-domain reference and a `resultingName`), and narrative; the Direction starts in `draft`
- [x] The placement field is named `resultingName` (not `resultingNameHint`) across the aggregate, command DTO, event payload, read model, API, and frontend
- [x] When capturing or editing a placement in a `draft` or `proposed` Direction, the UI pre-fills `resultingName` from the parent Enterprise Capability's current canonical name; the architect may override it per placement
- [x] On transition to `agreed`, attempts to edit `resultingName` on any placement are rejected by the aggregate with a clear error (same posture as type/horizon immutability) — enforced transitively via `ChangePlacements`, which already gates on `requireEditable()`; covered by existing `TestChangePlacements_OnAgreed_Fails`
- [x] Renaming the parent Enterprise Capability does not retroactively change `resultingName` on any existing Direction (pre-`agreed` or `agreed`) — by design, since `resultingName` is stored on the Direction aggregate, not resolved through the EC at read time
- [ ] *(deferred polish)* Pre-`agreed` Directions surface an explicit "refresh from Enterprise Capability" affordance per placement when the placement's `resultingName` differs from the EC's current name — not required for sign-off; tracked here so it does not get lost
- [x] An architect can advance a Direction's status from `draft` → `proposed` → `agreed`, and reject from `draft` or `proposed`
- [x] Each status transition is persisted as its own past-tense domain event; replaying the event store reconstructs the current status
- [x] An Enterprise Capability cannot host two simultaneously-active Directions; the second creation attempt is rejected with a clear error
- [x] Viewing an Enterprise Capability with a Direction surfaces, within the same view, the Direction's type, current status, and narrative — readable by any user with view permission on the Enterprise Capability
- [x] Viewing an Enterprise Capability without a Direction surfaces an explicit "no direction" state distinguishable from a draft
- [x] A Direction whose source list contains a deleted physical capability still renders, with the stale references clearly marked
- [x] Read-only users see the Direction but cannot create, advance, or reject
- [x] HATEOAS affordances on an Enterprise Capability's response advertise create-direction and advance-direction operations only when the calling user is authorised; readers see no such affordances
- [x] Other bounded contexts can subscribe to `architecturedirection` events for read-side use; the published-language event contract is documented
- [x] All BDD scenarios above have at least one corresponding test
- [x] CodeScene `pre_commit_code_health_safeguard` passes on every modified file

---

## Architecture

### Ownership

A new bounded context: `architecturedirection`. It owns the `Direction` aggregate end-to-end — write-side, read-side projections, API surface, frontend integration. The `enterprisearchitecture` context owns Enterprise Capabilities and is referenced read-only. The `capabilitymapping` context owns physical capabilities and is also referenced read-only via event subscriptions.

### Domain Model

The `Direction` aggregate carries: an identity, an Enterprise Capability ID, a type (`consolidate` / `decompose` / `stay`), one or more source physical capability IDs (cardinality per rule 4), an optional set of target placements (cardinality per rule 7), a horizon (`now` / `next` / `later`), a status, and a narrative. Status is reconstructed from the event log; status transitions are individual past-tense events.

Invariants are listed in Business Rules; the aggregate enforces them at the command boundary. Cross-aggregate invariants (uniqueness of active Direction per Enterprise Capability) are handled with the established pattern in EASI for one-active-per-parent (verified at the command-handler level via the read model).

### API Surface

Direction is exposed under the Enterprise Capability's resource tree — a Direction belongs to an Enterprise Capability and is fetched and written through routes that begin with `/api/v1/enterprise-capabilities/{id}/...`. The exact route shape is settled at implementation time per the API standards skill, but the contract obligation is: a single Enterprise Capability response surfaces (a) any active Direction inline or via a HATEOAS link, and (b) HATEOAS affordances for any operation the calling user is authorised to perform on it.

Status transitions are exposed as discrete operations (one for advancing, one for rejecting) rather than a free-form PATCH on status — this keeps the wire format honest about which transitions are valid.

### Persistence

Event-sourced, following the established EASI pattern. The aggregate's events stream is its source of truth; read models are projected from it. A read-side projection joins Directions with their parent Enterprise Capability for fast "show me the direction on this Enterprise Capability" queries.

### Frontend

A `Direction` panel surfaces on the existing Enterprise Capability detail page, occupying enough room for type, status, narrative, and source-capability list at a glance. Where the panel sits and how it composes with existing detail content is settled during implementation; the constraint is that the alignment question — *is there a Direction on this; what type; what status* — must be answerable in five seconds without scrolling or drill-down.

Status transitions are exposed as actions on the panel and gated by HATEOAS affordances per EASI's existing pattern.

### Cross-Context Integration

`architecturedirection` subscribes to:
- `enterprisearchitecture` Enterprise Capability events, to know which capabilities exist and which it can host a Direction on.
- `capabilitymapping` physical capability events, to know which sources are valid and to detect when a referenced source has been deleted (driving the stale-reference indicator).

`architecturedirection` publishes Direction lifecycle events for downstream contexts. No outbound write commands cross context boundaries.

---

## Design Decisions

1. **Direction as its own aggregate, not embedded in Enterprise Capability.** Independent lifecycle (a Direction can be drafted, debated, and rejected without touching the Enterprise Capability), distinct authorisation surface (architect-only writes vs broader Enterprise Capability reads), and clean separation between the steady-state classification (Enterprise Capability) and the change proposal (Direction). The DDD memo committed to this; the spec follows.

2. **Status transitions as discrete past-tense events, not a generic StatusChanged.** Aligns with the established EASI event-sourcing pattern; lets read-side projections subscribe to specific transitions (e.g. "every time a Direction is agreed, recompute the daily-alignment heat map") without filtering. Rejected because the alternative (one generic event with a status field) reads as less truthful in the event log and forces every consumer to know the status transition table.

3. **At most one active Direction per Enterprise Capability.** Multiple in-flight Directions on the same Enterprise Capability lead to ambiguous alignment answers. Reject-and-replace is cleaner than concurrent drafts. Alternative considered: allow multiple drafts with one designated as "primary" — rejected as a needless concept that fails the five-second test (which Direction is the answer?).

4. **Stale references surface but do not block.** A Direction whose source capability has been deleted is still meaningful (it carries a recorded group decision). Hiding it would lose history; blocking deletion of physical capabilities to protect Directions is the wrong tail wagging the dog. Alternative considered: hard-delete the Direction when a source is deleted — rejected because it loses the historical record of a decision the group made.

5. **Type is immutable; reject-and-replace to change.** A Direction's *type* (consolidate / decompose / stay) is the central commitment. Changing it mid-flight obscures the audit trail of what the group decided when. Reject-and-replace makes the change explicit. Alternative considered: type as mutable until `agreed` — rejected because it muddles the meaning of the draft → proposed transition.

6. **Permission family `architecture-direction:*`, scoped per tenant.** Mirrors the existing `enterprise-arch:*` pattern. Settled here (not deferred) so downstream specs 168–172 inherit the same scheme without re-deciding.

7. **`resultingName` lives on the Direction aggregate, is defaulted from the Enterprise Capability at capture, and is frozen on `agreed`.** The earlier draft of this spec gave each placement an optional `resultingNameHint`. In user review that was identified as carrying two names for the same thing — the EC's canonical name and the per-placement hint — with no defined precedence, inviting drift. A first revision proposed dropping the field entirely and resolving the name through the EC at read time. A strategic-DDD consult rejected that revision on aggregate-self-containment and audit-trail grounds: a Direction is a recorded group decision; a cheap, frequent EC rename should not silently mutate the apparent content of any past Direction, especially an `agreed` one. The accepted design keeps the field but reshapes it: rename to `resultingName` (drop the "hint" wording that invited the precedence ambiguity), default it from the EC at capture so the UX shows one input not two, and freeze it on transition to `agreed` alongside `type` and `horizon`. Renames of the EC after `agreed` do not propagate; the only path to change a frozen `resultingName` is reject-and-replace. Optimises for *audit-trail fidelity and aggregate self-containment* over *single source of truth at read time*; the spec's stated purpose (making recorded group decisions visible and trustworthy) makes the former the right fit.

8. **Placement cardinality rule duplicated between aggregate and capture form (accepted).** `validatePlacementCardinality` on the aggregate is the authoritative enforcer. The capture form additionally hides the "+ Add placement" button and shows an inline error using its own copy of the rule (`canAddPlacement` / `describePlacementRequirement`). Unlike status-transition actions on a persisted Direction — which are HATEOAS-driven by `_links` on the read model — the capture form is creation-time UX over a single POST contract, with no per-row affordance to attach. A HATEOAS-driven alternative was considered (sending the cardinality constraint as metadata, or splitting the create flow into per-placement appends) and rejected as over-engineered for form validation. The duplication is named here so it is acknowledged, not accidental; any future change to placement cardinality rules requires touching both sides.

9. **Bounded-context split (`architecturedirection` separate from `enterprisearchitecture`) confirmed.** The strategic-DDD consult considered merging the two and rejected the merge. `enterprisearchitecture` speaks classification (name, category, grouping of physicals); `architecturedirection` speaks intent and consensus (type, horizon, narrative, draft→proposed→agreed→rejected). Two distinct ubiquitous languages sharing an identifier is the textbook signature of a Customer/Supplier relationship, not one merged context. The "one active Direction per EC" rule is a non-transactional alignment rule — handler-level enforcement via the read model is the right tool, and merging would force every EC write to load Direction state. Relationship: `enterprisearchitecture` upstream as supplier (EC IDs and names are read-only inputs); `architecturedirection` downstream as customer. No ACL needed; published language is small and stable.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Direction is its own aggregate | Cross-aggregate invariant (one active per Enterprise Capability) needs handler-level enforcement, not aggregate-internal | Established EASI pattern; verified by read-model lookup at command time |
| Status transitions as discrete events | More event types in the published language | Each carries clearer semantics for projections; the count is bounded (4 transitions) |
| Stale references render | Direction may show degraded data | Stale indicator is explicit; user can edit the source list |
| At most one active Direction | A group exploring two alternatives at once cannot model both in the tool | Two alternatives can be discussed in the existing draft's narrative; if they diverge enough to need separate aggregates, that is itself a signal the group is past discussion and into decision |
| Type immutability | Editing a misclassified Direction requires reject-and-replace | The cost (one extra event in the log) is small relative to the audit-trail clarity gained |
| `resultingName` stored on the Direction (snapshot at capture, frozen on `agreed`) | A `resultingName` and the EC name can diverge over time; readers seeing an old Direction may notice its names no longer match the EC's current name | Acceptable — and the point. The Direction's job is to record what the group agreed to. Pre-`agreed` Directions can be re-defaulted on explicit user action; `agreed` Directions are deliberately frozen so the audit trail remains stable when the EC is renamed. The alternative (resolve through EC at read time) was considered and rejected for silently mutating past decisions. |

---

## Checklist

- [x] Specification ready
- [x] Implementation done — original slice plus the 2026-05-19 revision (see "Revision applied" log below)
- [x] Unit tests implemented and passing — 7/7 architecture-direction frontend; full backend suite green
- [ ] Integration tests implemented if relevant
- [x] API documentation updated — swagger regenerated; `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, `backend/docs/docs.go`, and `frontend/openapi.json` all show `resultingName` (no remaining `resultingNameHint`)
- [x] User sign-off

### Revision applied — 2026-05-19

Field renamed `resultingNameHint` → `resultingName`. No upcaster (aggregate was never deployed); no SQL migration (placements are JSONB).

Backend:
- `valueobjects.Placement` — `resultingNameHint` → `resultingName`; `MaxResultingNameHintLength` → `MaxResultingNameLength`; `ErrResultingNameHintTooLong` → `ErrResultingNameTooLong`. Accessor, equality, ctor arg all renamed.
- `commands.PlacementInput` — `ResultingNameHint` → `ResultingName`.
- `events.DirectionDrafted` / `events.DirectionPlacementsChanged` placement payload — `ResultingNameHint` → `ResultingName`; JSON tag `"resultingName"`.
- `aggregates.Direction` — no change needed. `ChangePlacements` already gates on `requireEditable()`, so post-`agreed` `resultingName` edits are blocked transitively. Existing `TestChangePlacements_OnAgreed_Fails` proves it.
- `readmodels.DirectionPlacementDTO` — field and JSON tag renamed.
- `infrastructure/api/handlers.go` — `PlacementRequest.ResultingNameHint` → `ResultingName`. Swagger regenerated.

Frontend:
- `types.ts` — `DirectionPlacement.resultingNameHint` and `PlacementInput.resultingNameHint` renamed to `resultingName`.
- `CaptureDirectionForm.tsx` — added `useEnterpriseCapability` to fetch the parent EC; new helper `useEnterpriseCapabilityName` derives the default; placement `add()` pre-fills `resultingName` with the EC's current canonical name. Input remains free-text and overrideable. Codehealth-driven refactor extracted `isReadyToSubmit`, `buildCaptureRequest`, `requiresExactlyOneSource`, `requiresPlacements` helpers — file at 10.0.
- `DirectionPanel.tsx` — renamed field reference in placement rendering.

Tests:
- `placement_test.go`, `direction_projector_test.go`, `transition_handlers_test.go` — field references updated; `TestNewPlacement_NameTooLong` replaces `TestNewPlacement_HintTooLong`.
- New `CaptureDirectionForm.test.tsx` — asserts "adding a placement pre-fills `resultingName` from the EC's canonical name." Written failing first, then made green.

Deferred (not required for sign-off):
- "Refresh from Enterprise Capability" affordance on pre-`agreed` placements when `resultingName` differs from the EC's current name. Tracked in the Acceptance Criteria list.
- A frontend test asserting that user-overridden `resultingName` on one placement is preserved when adding a second placement. Pre-fill behavior is symmetric per add, so the existing implementation satisfies it, but no explicit test covers the multi-placement case.
