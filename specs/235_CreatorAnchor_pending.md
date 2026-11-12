# 235 — Creator Anchor

> **Status:** pending
> **Depends on:** 228 (home, caches, scope)
> **Roadmap alignment:** `SD8 / H2-6` — slice A of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decision Q4); extends the anchors of spec 228 rule 1

---

## Problem Statement

Home shows a user what they answer for. A user who creates applications and capabilities but owns none of them answers for nothing in EASI's eyes: they have no anchor, so an architect sees the whole tenant and a stakeholder sees an empty page. In a reviewed production tenant 293 of 305 applications had no owner; every one of them was created by somebody who is never shown it again.

A record nobody answers for has exactly one person connected to it: whoever created it. This slice makes that connection an anchor, for as long as nobody else answers for the record, and shows those records on My Work as "Created by you".

---

## Ubiquitous Language

| Term | Meaning |
|------|---------|
| **Creator** | The user who was the actor of the subject's creation event. Imported subjects are created by the user who ran the import. |
| **Unanswered** | An application in ownership state `unknown`; a capability whose L1 ancestor (or itself, if L1) has no EA owner that is a user of the tenant. |
| **Creator anchor** | The anchor between a user and an unanswered subject they created. |

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Architect / Admin** | See how many of the applications and capabilities they created still have nobody answering for them, and reach each one |
| **Importer** | See what their import left unanswered |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Creator anchor

  Scenario: Created and unowned applications are mine
    Given "Per Lund" created the applications "CRM" and "Portal"
    And neither has an owner
    And "Per Lund" holds no other anchor
    When "Per Lund" opens Home
    Then their scope kind is personal
    And My Work lists "CRM" and "Portal" as "Created by you"

  Scenario: An owner ends the creator anchor
    Given "Per Lund" created "CRM"
    When "Ole Berg" is assigned as owner of "CRM"
    Then "CRM" is no longer in "Per Lund"'s My Work

  Scenario: A nomination ends the creator anchor
    Given "Per Lund" created "CRM"
    When "Ole Berg" is nominated as owner of "CRM"
    Then "CRM" is no longer in "Per Lund"'s My Work

  Scenario: Clearing ownership restores it
    Given "Per Lund" created "CRM", owned by "Ole Berg"
    When the ownership of "CRM" is cleared
    Then "CRM" is in "Per Lund"'s My Work as "Created by you"

  Scenario: Created capabilities without an EA owner are mine
    Given "Per Lund" created the L1 "Billing" and its child "Invoicing"
    And "Billing" has no EA owner
    Then My Work lists "Billing" and "Invoicing" as "Created by you"
    When "Alice Smith" becomes EA owner of "Billing"
    Then neither is in "Per Lund"'s My Work

  Scenario: My Work summarises what I created
    Given "Per Lund" created 140 applications that have no owner
    When "Per Lund" opens Home
    Then My Work shows the count "140 created by you with no owner"
    And lists the first 12 by name with a link to all of them

  Scenario: Created subjects I also own are listed once
    Given "Per Lund" created "CRM" and holds an edit grant on it
    Then "CRM" appears once in My Work, as "Created by you" and "Edit access"

  Scenario: A disabled creator anchors nothing
    Given "Per Lund" created "CRM" and is disabled
    Then no user holds a creator anchor on "CRM"

  Scenario: Agents and system actors are not creators
    Given "CRM" was created by an actor that is not a user of the tenant
    Then no user holds a creator anchor on "CRM"
```

---

## Business Rules & Invariants

1. **Creator anchor, exactly** — the caller holds a creator anchor on every application they created whose ownership state is `unknown`, and on every capability they created whose L1 ancestor (or itself) has no EA owner that is a user of the tenant. It is added to the anchor list of spec 228 rule 1; nothing else about that rule changes.
2. **The creator is the actor of the creation event** — `ApplicationComponentCreated` and `CapabilityCreated`. It never changes. An actor that is not an active user of the tenant anchors nothing.
3. **Scope** — a creator-anchored application is an anchored application, and a creator-anchored capability is treated like an EA-owned capability, for portfolio, domains and TIME (spec 228 rules 3–5).
4. **My Work** — creator-anchored subjects carry the relation `created`. A subject with several relations appears once with all of them. Cards with only the relation `created` sort after all other cards, by name.
5. **Summary and cap** — My Work carries, per subject type, the number of cards whose only relation is `created`. At most 12 such cards per type are included; the count is always the full number.
6. **Computed at read time** — nothing about the anchor is stored beyond the creator on the cache row.
7. **Visibility of attention items** — a creator anchor counts as "anchored to the subject" in spec 230 rule 3.

---

## Acceptance Criteria

- [ ] The application and capability caches carry the creator's user id, set from the creation event's actor and seeded by a backfill migration from the event store's creation events
- [ ] Anchor resolution has a test per state: unknown (anchor), nominated, owned, managed (no anchor), cleared (anchor again); capability with and without an L1 EA owner, with a legacy free-text EA owner (anchor), at L1 and at L3
- [ ] A disabled creator and a non-user actor anchor nothing
- [ ] `GET /api/v1/home` returns relation `created` on the cards, the per-type `createdUnanswered` counts, and at most 12 created-only cards per type
- [ ] A caller with only creator anchors has scope kind `personal` and a portfolio per rule 3
- [ ] Home renders the "Created by you" label, the count line and a link that opens the application list or the capability tree filtered to "created by me"
- [ ] Assigning, nominating or clearing an owner, and setting or clearing an EA owner, change the creator's My Work on the next load
- [ ] Tenant isolation: a creator id matching a user of another tenant anchors nothing
- [ ] Every BDD scenario has at least one corresponding test
- [ ] `docs/architecture/Stewardship.md` lists the creator anchor
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

The `stewardship` read side. No other context changes: the creator is read from the envelope of events already subscribed to.

### Domain Model

The scope resolver gains one anchor kind, `created`, decided from the cached creator, ownership state and EA owner. No aggregate changes.

### API Surface

`GET /api/v1/home` only: `myWork.items[].relations` may contain `created`; `myWork.createdUnanswered` carries `{ applications, capabilities }`; `scope` names the creator anchors by count. Redaction follows spec 228 rule 8.

### Persistence

One nullable creator column on the application cache and on the capability cache, tenant-scoped under the existing RLS. A `backfill` migration reads the creation events (version 1 of each aggregate) from the event store, the same source Audit's artifact-creator read model uses.

### Frontend

`home` feature: relation label, count line, and a link to the existing "created by" filter with the caller preselected.

### Cross-Context Integration

None new. Projectors read the actor from the event envelope they already receive.

---

## Design Decisions

1. **The anchor ends when someone answers for the subject** — creating is not accountability; it is the only link an unowned record has. Ending it on nomination as well as ownership means the creator's list empties as soon as they point at someone. Alternative: a permanent anchor (rejected: importers carry hundreds of cards forever, and the list stops meaning "needs me").
2. **Capabilities follow the L1's EA owner** — spec 230 checks EA ownership on L1 only; using the same rule keeps the creator's list and the `capability-without-ea-owner` check in agreement.
3. **Cards are capped, the count is not** — an import of 300 applications must not make Home slow or unreadable, but the number is the message.
4. **Creator from the event store, not from Audit's endpoint** — Audit's read model is another context's query; the cache is fed by events the read side already consumes (spec 209).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Creator anchors make an architect's scope personal | An architect who created one unowned application stops seeing the tenant-wide portfolio (spec 228 rule 2) | The scope header says so; assigning an owner returns them to the tenant view |
| Imported subjects anchor to the importer | One person may hold most of a tenant's unanswered records | That is the true state; the count shows it and bulk fixes (spec 240) close it |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
