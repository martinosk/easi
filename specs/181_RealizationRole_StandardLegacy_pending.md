# 181 — Realization Role: Standard / Legacy per Capability Realisation

> **Status:** pending
> **Depends on:** 179 (Domain Board surfaces)
> **Related:** 170 (StandardApplication on an EC — unchanged, different granularity)
> **Design doc:** [`docs/specs/capability-journeys.md`](../docs/specs/capability-journeys.md)

---

## Problem Statement

For a given domain capability with several realising applications, the tool cannot say which one is the blessed app *today* and which are kept-but-legacy. Spec 170's `StandardApplication` answers this per Enterprise Capability, but most domain capabilities sit under no EC direction, so the Domain Board has nowhere to show "this is the standard, that is legacy".

This slice adds a **realization role** — `standard` or `legacy`, unset meaning unclassified — set by architects per capability-app realisation. The role is the current designation of the landscape; it is deliberately independent of TIME assessment (spec 180): a standard app assessed Eliminate is a legitimate, visible tension, not a validation error. The role also becomes the substrate of the Target lens (spec 183): where no journey says otherwise, the target state of a capability is its standard-role application.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Domain Architect** | Mark, per capability, which realising app is the standard and which are legacy; correct the designation as the landscape shifts. |
| **Engineer / Product Manager** | See at a glance on the board which app to build against for a capability. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Realization role on a capability realisation

  Scenario: An architect marks a realisation as the standard
    Given capability "Booking management" is realised by "Phoenix" and "Seabook", neither with a role
    When the architect marks "Phoenix" as standard
    Then the "Phoenix" chip renders in the standard style on the Domain Board
    And the capability drawer shows "Phoenix — standard"

  Scenario: Marking a new standard displaces the previous one
    Given "Seabook" holds the standard role for "Booking management"
    When the architect marks "Phoenix" as standard
    Then "Phoenix" is the standard
    And "Seabook" becomes unclassified
    And the displacement is reconstructable from the event history

  Scenario: Any number of realisations can be legacy
    Given a capability realised by three applications
    When the architect marks two of them as legacy
    Then both render in the legacy style
    And no error is raised

  Scenario: An architect clears a role
    Given "Seabook" is marked legacy for a capability
    When the architect clears the role
    Then "Seabook" presents as unclassified

  Scenario: Roles and TIME are independent
    Given "Salesforce" holds the standard role for "CRM"
    When the architect assesses the same realisation as "Tolerate"
    Then both the standard role and the Tolerate grade are stored and displayed
    And neither write is rejected because of the other

  Scenario: Inherited realisations carry no role
    Given an ancestor capability surfaces an app through an inherited realisation
    Then the inherited chip renders without role styling
    And no role affordance is offered on the inherited row

  Scenario: Deleting a realisation clears its role
    Given a realisation holding the standard role
    When the realisation is deleted
    Then the role is cleared as a recorded reaction
    And the capability presents with no standard

  Scenario: Read-only users see roles but cannot write
    Given a user without the architect permission
    When they fetch a capability's realisations
    Then roles are visible but the responses carry no assign/clear affordances
```

---

## Business Rules & Invariants

1. **Role values** — `standard` | `legacy`. Absence of a role is the unclassified state; there is no stored "unclassified" value.
2. **At most one standard per capability** across its direct realisations. Assigning standard to another app displaces the previous holder to unclassified in the same command; the displacement is explicit in the event payload.
3. **Legacy is unbounded** — any number of a capability's realisations may be legacy.
4. **Direct realisations only** — a role attaches to a (capability, component) pair with a direct realisation, verified at assign time. Inherited rows surface no role.
5. **Role is independent of TIME assessment (180) and of the EC-level `StandardApplication` (170)** — no write-side validation couples them. Disagreement between role and TIME is surfaced as a signal (spec 184), never blocked.
6. **Realisation deletion clears the role** via an event-driven reaction (172 R6 pattern); the clearing is a recorded event, not a silent read-side hide.
7. **Every role change records the acting user and timestamp** on the event.
8. **Authorisation** — writes require the `architecture-direction:*` architect permission; affordances are HATEOAS-advertised only when authorised.

---

## Acceptance Criteria

- [ ] An architect can assign standard / legacy to a direct realisation and clear it; the board chip styling and drawer update within one refetch
- [ ] Assigning standard when another app holds it atomically displaces the previous holder to unclassified, reconstructable from events
- [ ] Multiple legacy roles per capability are accepted; a second standard is impossible by construction
- [ ] Role writes never validate against TIME assessments or EC standards, and vice versa
- [ ] Inherited chips render no role styling and offer no role affordances
- [ ] Deleting a realisation emits a recorded role-clearing reaction; the drawer shows the capability with no standard afterwards
- [ ] Board chips render: standard → positive-tinted, legacy → progress/amber-tinted, unclassified → neutral (tokens, per mockup)
- [ ] Write affordances are HATEOAS-gated; read-only users receive none
- [ ] Every BDD scenario has at least one corresponding test; every business rule has a unit test
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

`architecturedirection` (design doc D7). `capabilitymapping` is referenced read-only for realisation existence; realisation-deleted events drive the clearing reaction.

### Domain Model

`RealizationRoles` — **one aggregate per capability** (intrinsic UUID; one-per-capability enforced handler-level with DB backstop, 170 pattern), holding the component → role map for that capability's direct realisations. This makes rule 2 (single standard, atomic displacement) an intrinsic aggregate invariant instead of a cross-aggregate dance. Events: `RealizationRoleAssigned` (carries capability, component, role, and the displaced component when a standard moves) and `RealizationRoleCleared` (explicit clear and the deletion reaction).

### API Surface

Role assign/clear as discrete operations on the capability's realisation resource tree; roles included in the bulk board query alongside assessments (180) so one request serves a domain. HATEOAS gates writes. Shapes per `easi-api-standards` at implementation time.

### Persistence

Event-sourced. Read model: (tenant, capability, component) → role rows with denormalised names; joined into the board query.

### Frontend

- `AppChip`: role tinting (standard/legacy) composed with the existing `realizationLevel` and `Inherited` styling — role tint wins on direct chips, per mockup.
- Capability drawer app rows: role display + assign/clear control beside the TIME control from 180.
- `mutationEffects.ts`: role mutations invalidate the board realisation/role queries.

### Cross-Context Integration

Subscribes to `capabilitymapping` realisation-deleted events → `RealizationRoleCleared` reaction. No outbound writes.

---

## Design Decisions

1. **Per-realisation designation, independent of EC `StandardApplication`** — user decision 2026-07-12. Deriving from the EC standard via direction composition leaves the board blank wherever no EC direction exists; deriving from TIME (Invest ⇒ standard, spec 119's shortcut) makes the standard-but-Eliminate tension unrepresentable, and that tension is a first-class signal (184). The EC standard remains authoritative at its own granularity; a later read surface may compare the two, which is exactly a signal, not a constraint.
2. **One `RealizationRoles` aggregate per capability, not one per pair** — the single-standard invariant with atomic displacement lives inside one aggregate boundary. Per-pair aggregates (the 180 shape) rejected: displacing a standard would require writing two aggregates in one command, a non-transactional cross-aggregate mutation.
3. **Displacement to unclassified, not to legacy** — demoting the old standard to legacy would encode a judgement the architect didn't make; unclassified is the honest default, one click from legacy if intended.
4. **Deletion clears via reactor, not read-side hiding** — a role without its realisation is meaningless state; the 172 R6 reactor pattern keeps the aggregate honest and the audit trail explicit. (Contrast 180, where an assessment is a historical judgement worth keeping — there, hiding suffices.)

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Role independent of EC standard | Two "standard" notions at different granularity can disagree | By design; different questions ("this capability" vs "this EC"); disagreement is visible on both surfaces |
| Aggregate per capability | Role events for unrelated apps share a stream | Streams stay tiny (a handful of realisations per capability) |
| Displacement to unclassified | Old standard loses any designation silently unless re-marked | Displacement is explicit in the event and visible in the drawer |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
