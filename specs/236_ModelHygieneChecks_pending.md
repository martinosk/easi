# 236 — Model Hygiene Checks

> **Status:** pending
> **Depends on:** 230 (attention items: checks, visibility, ranking, Home summary), 237 (normalised name)
> **Roadmap alignment:** `SD8 / H2-6` — slice B of [`docs/specs/data-quality.md`](../docs/specs/data-quality.md) (decision Q5); extends the check list of spec 230 rule 1

---

## Problem Statement

Spec 230 finds unowned applications, unassessed realisations, incomplete one-pagers and unrealised capabilities. A review of a production tenant found five further gaps that no check covers: 185 of 305 applications with hosting unknown, 49 applications realising no capability, 20 L1 capabilities assigned to no business domain, and duplicate application and sibling capability names. None of these is visible anywhere today except by exporting the model.

This slice adds five checks to the attention engine so that these gaps are counted on Home and listed for the people whose scope covers them.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Steward of documentation or structure** | See these gaps in their domains |
| **Admin / Architect** | See them tenant-wide, including subjects that belong to no domain |
| **Owner / Creator** | See them on their own applications and capabilities |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Model hygiene checks

  Scenario: Hosting unclassified
    Given the application "CRM" has hosting "unknown"
    Then a "hosting unclassified" item exists for "CRM"
    When its hosting is classified as "saas"
    Then the item is gone

  Scenario: Application realises nothing
    Given the application "Portal" directly realises no capability
    Then an "application realises nothing" item exists for "Portal"
    When "Portal" is linked to the capability "Customer Self Service"
    Then the item is gone

  Scenario: A part of another application is not reported as realising nothing
    Given "Portal API" is a part of "Portal" and realises no capability
    Then no "application realises nothing" item exists for "Portal API"

  Scenario: L1 capability outside every domain
    Given the L1 capability "Tracking" is assigned to no business domain
    Then a "capability outside every domain" item exists for "Tracking"
    And no such item exists for its children

  Scenario: Duplicate application name
    Given two applications are named "Sevron" and "sevron "
    Then a "duplicate name" item exists for each, naming the other
    When one is renamed "Sevron Workshop"
    Then both items are gone

  Scenario: Duplicate capability name among siblings
    Given the capability "Load and Discharge Execution" has two children named "Load Execution"
    Then a "duplicate name" item exists for each child, naming the other

  Scenario: Two L1 capabilities with one name
    Given two L1 capabilities are named "Accounts Receivable"
    Then a "duplicate name" item exists for each

  Scenario: The same name under different parents is not a duplicate
    Given "Damage Recording" exists under "Unit Management" and under "Yard Management"
    Then no "duplicate name" item exists for either

  Scenario: Subjects outside every domain reach tenant-wide readers
    Given "Portal" realises nothing and so belongs to no domain
    Then its items are visible to callers with scope kind tenant, to its owner and to its creator
    And to no steward

  Scenario: Home counts the new checks
    Given the caller sees 185 "hosting unclassified" items
    When they open Home
    Then Needs Attention shows "185 applications with hosting unclassified"
```

---

## Business Rules & Invariants

1. **The checks** — added to spec 230 rule 1, fixed in code:

   | Key | Concern | Severity | Subject | Finding |
   |-----|---------|----------|---------|---------|
   | `hosting-unclassified` | documentation | low | application | hosting classification `unknown` |
   | `application-unrealised` | structure | medium | application | no direct realisation, and not a part of another application |
   | `capability-outside-domain` | structure | medium | capability | L1 assigned to no business domain |
   | `application-duplicate-name` | structure | medium | application | another application has the same normalised name |
   | `capability-duplicate-name` | structure | medium | capability | another capability with the same parent (or both L1) has the same normalised name |

2. **Normalised name** — the shared-kernel definition of spec 237 rule 2: trimmed, case-folded, internal whitespace collapsed to one space.
3. **Duplicate items name the others** — the item's context lists the other subjects sharing the name (id and name), at most five, with the total.
4. **Item domains** — as in spec 230; `application-unrealised` and `capability-outside-domain` always have none, and duplicates may have none.
5. **Fixer** — `hosting-unclassified`: the application's owner, else the domain architect, as for other application subjects. `application-unrealised` and the duplicate checks on applications: the owner. `capability-outside-domain` and `capability-duplicate-name`: the EA owner of the capability or its L1 ancestor. Null when none qualifies.
6. **Everything else is spec 230** — visibility, ranking, routing, read-time computation, no dismissal, permissions per subject, Home summary.
7. **Concern descriptions** — "documentation" reads "Every one-pager is complete and every application's hosting is classified"; "structure" reads "Every capability is realised, owned and in a domain; every application realises a capability; names are unique where they must be".

---

## Acceptance Criteria

- [ ] Each check has tests for a finding, a non-finding and its boundary: hosting `third-party-hosted` (no finding), inherited-only realisation (finding), a composed and an aggregated part (no finding), L1 vs L2 outside a domain, names differing only by case and whitespace (duplicate), by punctuation (not duplicate), same name under different parents (not duplicate), an L1 and an L2 sharing a name (not duplicate)
- [ ] The application cache carries hosting and the parent application id, fed by `ApplicationHostingClassified`, `ComponentAttached` and `ComponentDetached`, seeded by a backfill migration
- [ ] `GET /api/v1/attention-items` returns the new checks with their context; `check` accepts the new keys
- [ ] Duplicate checks run as one grouped query per subject type, not one query per subject
- [ ] Items with no domain are returned to scope kind `tenant` and to anchored users, and are absent from every domain view
- [ ] Home's Needs Attention shows the new checks' rows with their labels
- [ ] The stewards dialog shows the concern descriptions of rule 7
- [ ] Every BDD scenario has at least one corresponding test
- [ ] `docs/architecture/Stewardship.md` lists the checks
- [ ] Every modified file scores 10.0 per `easi-codehealth`

---

## Architecture

### Ownership

The `stewardship` read side: five `Check` values, their queries, two cache columns. No other context changes.

### Domain Model

No aggregate changes. `Check` list grows; `Concern` descriptions change.

### API Surface

No new resources. The attention item's `context` gains `duplicates` (`total`, `items[]` of id and name) for the duplicate checks.

### Persistence

Application cache: hosting classification and parent application id, with a `backfill` migration from Architecture Modeling's tables. A functional index on the normalised name of the application cache and on (parent, normalised name) of the capability cache.

### Frontend

Labels for the new checks on Home and the attention list; duplicate items render the other subjects as links.

### Cross-Context Integration

| Direction | Events | Purpose |
|-----------|--------|---------|
| Architecture Modeling → Stewardship | `ApplicationHostingClassified`, `ComponentAttached`, `ComponentDetached` | Hosting and containment on the application cache |

---

## Design Decisions

1. **No new concern** — hosting is a documentation fact and the other four are structure. A concern is something a person can be accountable for; "hygiene" is not a distinct accountability.
2. **Duplicates are exact normalised matches only** — items cannot be dismissed (spec 230 rule 8), so a check must have no legitimate false positive. Two applications with the same name always need a rename or a merge. Near matches such as "Legislation Update Server" and "Legislation Update Service" are advised at creation (spec 237). Alternative: similarity-based findings with a "distinct" mark (rejected: dismissal under another name).
3. **Capability duplicates are scoped to siblings and to L1s** — the same name under different parents is ordinary modelling.
4. **Parts are exempt from `application-unrealised`** — a part contributes through its parent (spec 216); reporting each part would repeat one gap.
5. **Missing description and missing criticality are not checks** — both are one-pager fields; marking them required makes `one-pager-incomplete` report them, which keeps the tenant in control of what "complete" means.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| `hosting-unclassified` at severity low | In a tenant where most applications are unclassified it is the longest list | Low sorts last; bulk classify (spec 240) is the intended fix |
| Subjects without a domain reach no steward | Only tenant-wide readers, owners and creators see them | Those are the people who can link or assign them; once linked, stewards see the subject |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
