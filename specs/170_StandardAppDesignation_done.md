# 170 — Standard Application on an Enterprise Capability

> **Status:** ongoing
> **Depends on:** [167 — Direction on an Enterprise Capability](167_Direction_Aggregate_Capture_done.md)
> **Followed by:** [171 — Denormalize Cross-Context Reference Names into Read Models](171_DenormalizeRefNamesIntoReadModels_pending.md)
>
> **Revision 2026-05-25** — Corrections applied during implementation, captured here so the spec stays an honest record:
>
> 1. **Aggregate identity is a random UUID, intrinsic to `StandardApplication`.** An early review iteration suggested using the EC's ID as the aggregate ID to make the per-EC invariant "intrinsic"; this violated tactical DDD (aggregate IDs identify the aggregate, never another aggregate) and caused a real event-store stream-collision bug in QA because the shared event store keys streams by `(tenant_id, aggregate_id)` only. The per-EC uniqueness invariant is now enforced at the command handler via `FindAggregateIDForEnterpriseCapability` on the read model, with the `uq_standard_applications_per_ec` unique constraint as a database-level backstop. This mirrors how `Direction` enforces "one active per EC" (167).
> 2. **Command handler verifies EC existence before persistence**, via the shared `services.ReferenceChecker.EnterpriseCapabilityExists` already used by 167. A security review caught the asymmetry where direction-captures verify but standard-set commands previously did not.
> 3. **HTTP semantics tightened:** PUT returns 201 + `Location` on first creation, 200 on replacement; GET history returns 200 with an empty entries list when no Standard has ever been set (not 404, which was ambiguous with EC-not-found).
> 4. **Error registry wired** for `ErrStandardApplicationAlreadyExists` → 409 and `ErrNarrativeRequiredForStandardApplication` → 400; without these the most common user errors would have surfaced as 500.
> 5. **Read-side denormalisation of the application name deferred to spec 171.** The current implementation resolves the application name in the React layer via `useComponents`; this works but is wrong-shaped (cross-context join in the client; GUIDs flash on load). 171 denormalises the name onto the read-model rows and removes the client-side lookup.

---

## Problem Statement

A daily decision an engineer or product manager faces is: *"For this capability, which application should I be using?"* Today the answer lives in tribal knowledge and slides — different domains use different apps for similar work, and the architecture group's view on which one *should* be the standard is not visible in the tool.

After 167 shipped, an Enterprise Capability surfaces its Direction but says nothing about the application that should realise it. This slice fills that gap: an EC can carry a **Standard Application** — the architecture group's recorded answer to "which app for this EC" — so anyone navigating to the EC can answer the question in five seconds. When the group changes its mind, the prior standard is preserved in history.

This is the Type-2 path from the conceptual model: physical capabilities can stay distributed across domains while the application landscape consolidates. The decision is independent of any Direction on the same EC.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Engineer / Product Manager** | Open an Enterprise Capability and learn in five seconds which application is the agreed standard, or that there isn't one yet. |
| **Enterprise Architect** | Set the agreed standard for an EC with a narrative explaining the choice; change it when the group reaches a new decision; have past standards preserved for audit. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Standard application on an Enterprise Capability

  Scenario: An EC with no standard surfaces that explicitly
    Given I am viewing an Enterprise Capability with no Standard Application
    Then the surface shows "no standard yet"

  Scenario: An architect sets a Standard Application
    Given I am an architect viewing an Enterprise Capability with no Standard Application
    When I set an application as the standard with a narrative
    Then the EC's surface shows the application as the current standard, with the narrative and the date it was set

  Scenario: An architect changes the Standard Application
    Given an Enterprise Capability has application A as its current Standard Application
    When I set application B as the new standard with a narrative
    Then the EC's surface shows application B as the current standard
    And application A is preserved in the EC's standards history with the date range it was current

  Scenario: A reader sees the Standard in context
    Given I have read access to an Enterprise Capability with a Standard Application
    When I open the Enterprise Capability
    Then I see the current Standard Application named clearly, with its narrative
    And I cannot set or change the Standard Application

  Scenario: A reader can review the history of past Standards
    Given an Enterprise Capability has had two or more different Standard Applications over time
    When I open its standards history
    Then I see each past Standard with its application, narrative, and the date range it was current

  Scenario: The Standard references an application that becomes stale
    Given the current Standard Application references an application that has since been deleted
    When I view the Enterprise Capability
    Then the missing reference is marked stale
    And the EC otherwise renders normally

  Scenario: A read-only user has no Standard-related affordances
    Given I have read-only access
    When I fetch an Enterprise Capability response
    Then the response carries no HATEOAS affordances for setting or changing the Standard Application
    And the UI surfaces no actions for those operations
```

---

## Business Rules & Invariants

1. **An Enterprise Capability has at most one current Standard Application.** Setting a Standard on an EC that already has one replaces it; the prior Standard is preserved in history.
2. **Setting a Standard requires a narrative.** One to two stakeholder-readable sentences naming why this application and what it covers (e.g. "covers the operational and reporting layers; excludes legacy COBOL flows"). The narrative is required at set/change time, not optional.
3. **Set and change are recorded as discrete past-tense events**, one per occurrence. Replay reconstructs the current Standard and the full history. There is no generic "standard changed" event.
4. **The Standard references one application by ID.** The reference is eventually consistent with the application catalog; deleted applications surface as stale but do not block reading or further changes.
5. **The Standard is independent of any Direction on the same EC.** Direction answers "where does the capability live"; Standard answers "which application realises it." Both can exist on the same EC and neither requires the other.
6. **Authorisation matches `architecturedirection`** (per spec 167). Setting and changing the Standard require the architect permission; read follows the EC's read permission. HATEOAS advertises operations only when authorised.

---

## Acceptance Criteria

- [ ] An architect can set a Standard Application on an EC with an application and a narrative; the EC's surface shows it as current
- [ ] An architect can change the Standard on an EC that already has one; the old Standard moves to history with its date range, the new Standard becomes current
- [ ] Each set / change is persisted as its own past-tense domain event; replay reconstructs the current Standard and the full history
- [ ] Setting or changing without a narrative is rejected with a clear error
- [ ] An EC with no Standard surfaces an explicit "no standard yet" state in its detail view
- [ ] An EC with a Standard surfaces the application and narrative in the same view as the rest of the EC's detail
- [ ] A reader can open the EC's standards history and see each past Standard with application, narrative, and date range
- [ ] A Standard referencing a deleted application renders with a stale indicator and does not block further changes
- [ ] HATEOAS on the EC response advertises set / change only when the caller is authorised; the UI gates its action buttons on those links
- [ ] Read-only users can view the current Standard and the history but receive no set/change affordances
- [ ] Every BDD scenario above has at least one corresponding test
- [ ] Every modified file scores 10.0 in CodeScene per `easi-codehealth`

---

## Architecture

### Ownership
The Standard Application is owned by `architecturedirection`, parallel to `Direction`. Same bounded-context split as 167 (Decision 9): `enterprisearchitecture` supplies EC identity and naming; `architecturedirection` records the group's decisions about the EC. Application IDs come from the application catalog (existing context); the reference is read-only.

### Domain Model
A small aggregate `StandardApplication` carries: its own random-UUID identity, the parent EC ID (as a typed `EnterpriseCapabilityRef`), the current application ID (as a typed `ApplicationRef`), and the current narrative. The aggregate's event stream is its source of truth; past Standards are reconstructed from past events for the history view. Events: `StandardApplicationSet` (covers both first-time set and replacement; payload includes the EC ID, the new application ID, the narrative, and the previous application ID if any).

No draft/proposed/agreed workflow. When the architect decides, they set it; the previous one moves to history. This contrasts deliberately with `Direction`'s lifecycle — see Design Decision 1.

### API Surface
The Standard is exposed under the Enterprise Capability's resource tree — the EC response surfaces the current Standard inline (or via HATEOAS link). Set and change are discrete operations on the EC, not a free-form PATCH. A history sub-resource lists past Standards in reverse chronological order. Exact route shapes settled at implementation time per `easi-api-standards`.

### Persistence
Event-sourced, same pattern as `Direction`. A read-side projection joins each EC with its current Standard for fast "show me the standard on this EC" queries; the projection also exposes the history list.

### Frontend
A Standard Application panel surfaces on the existing Enterprise Capability detail page, alongside the Direction panel from 167. The panel answers the five-second question (current app + narrative, or explicit "no standard yet"); a secondary action opens the history list. Architect actions (set, change) are gated by HATEOAS links from the EC response per the established pattern.

### Cross-Context Integration
`architecturedirection` subscribes to the application catalog for application existence and to detect stale references (same wiring as 167 uses for `capabilitymapping`). No write commands flow outward; the Standard is local bookkeeping on the architecture-direction side.

---

## Design Decisions

1. **No draft / proposed / agreed lifecycle.** `Direction` (167) models the in-flight debate because stakeholders care about the existence of an open question. The Standard's daily-alignment user (engineer/PM) reads only the agreed answer; modelling drafts here would add ceremony the use case does not request. When the group decides, the architect sets the Standard — that *is* the decision. Past Standards in history carry the audit; in-flight debate, if it needs visibility, can be added in a later slice.
2. **Set-and-replace, no separate "supersede" verb.** Replacement is just another `Set`. The history reconstructs from the event stream. Avoids a status proliferation (`superseded` vs `replaced` vs `archived`) that the user need does not justify.
3. **`StandardApplication` as a small aggregate in `architecturedirection`, not a field on the EC aggregate in `enterprisearchitecture`.** Follows 167's bounded-context split for the same reason: classification (EC) and group-decision (Standard) speak different ubiquitous languages. Keeping the Standard on the architecture-direction side preserves the existing supplier/customer relationship and avoids forcing every EC write to load Standard state.

6. **Aggregate identity is intrinsic; per-EC uniqueness is enforced at the handler.** The aggregate has its own random UUID. The "one current Standard per EC" invariant is enforced at the command handler via a read-model lookup (`FindAggregateIDForEnterpriseCapability`), with a database unique constraint as backstop. This matches `Direction`'s "one active per EC" pattern (167). Tactical DDD: aggregate IDs identify the aggregate, never another aggregate. (See revision 1 above for the failure that motivated locking this down.)
4. **Narrative required at set/change time.** A standard without a stated rationale is the slide-deck failure mode this spec exists to remove. The friction is intentional.
5. **Independent of Direction on the same EC.** The two answer different questions (Type-1 vs Type-2 in the conceptual model). The aggregates do not reference each other; a future cross-aggregate view can join them when synthesis is needed.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| No in-flight workflow | The architect cannot record "we are considering app B but have not agreed yet" in this slice | Acceptable — pre-decision exploration can live in a Discovery-style artifact (spec 169 pattern) if and when the need is real; the Standard itself stays focused on the agreed answer |
| Set-and-replace model | "Why did we replace A with B?" is answered by the new Standard's narrative, not by an explicit "supersede" record on A | The new narrative is the right place to explain the change; the history view shows the sequence |
| Small aggregate, no draft/proposed/agreed | Diverges from 167's lifecycle shape; reviewers may expect parallelism | Decision 1 names the divergence and grounds it in the differing user need |
| Cross-context for app catalog references | Stale-reference handling needed | Reuses 167's existing subscription wiring; same indicator pattern |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant — 6 HTTP integration tests against real Postgres covering first-set/replacement/history/per-EC uniqueness/aggregate-identity isolation
- [x] API documentation updated — swagger regenerated; `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, `backend/docs/docs.go`, `frontend/openapi.json` all updated
- [x] User sign-off — pending; the application-name rendering follow-up tracked under spec 171
