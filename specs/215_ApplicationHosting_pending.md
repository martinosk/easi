# 215 — Application Hosting Classification

> **Status:** pending
> **Depends on:** [214_ApplicationOwnership](214_ApplicationOwnership_pending.md) (statistics endpoint)
> **Roadmap alignment:** SD6 / H1-2

---

## Problem Statement

Nothing in EASI records where an application runs. Architects cannot filter the landscape by hosting model, and cloud-migration or SaaS-consolidation conversations happen outside the tool. A hosting classification on the application record is also the deliberate trailhead toward a technology portfolio (parked, B3) without building one.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | See the hosting distribution of the landscape; filter by hosting model |
| **Application Owner** | Record where their application runs |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Application hosting classification

  Scenario: New applications are unclassified
    Given an architect creates an application component
    Then its hosting classification is "Unknown"

  Scenario: Classifying an application
    Given an application with hosting "Unknown"
    When an architect classifies it as "SaaS"
    Then the application shows hosting "SaaS"

  Scenario: Filtering by hosting
    Given applications classified across several hosting models
    When an architect filters the application list by "On-premises"
    Then only applications with that classification are listed

  Scenario: Hosting distribution
    Given applications classified across several hosting models
    When an architect views the application statistics
    Then a count per hosting classification is shown alongside the ownership counts
```

---

## Business Rules & Invariants

1. **Closed classification** — hosting is exactly one of: `on-premises`, `cloud`, `saas`, `third-party-hosted`, `unknown`; components start as `unknown`.
2. **Reclassification is unrestricted** — any classification may replace any other; each change is its own domain event.
3. **HATEOAS affordance** — the classify affordance appears on the component for actors with `components:write`.

---

## Acceptance Criteria

- [ ] Component responses and list rows carry the hosting classification
- [ ] Classification raises `ApplicationHostingClassified` and is visible immediately
- [ ] The application list filters by hosting classification
- [ ] The statistics endpoint from spec 214 includes a hosting distribution
- [ ] Existing components read as `unknown` after deployment

---

## Architecture

### Ownership

ArchitectureModeling only. No cross-context integration.

### Domain Model

`HostingClassification` value object on the ApplicationComponent aggregate; command raises `ApplicationHostingClassified`. Aggregates without the event read as `unknown`.

### API Surface

A classify operation on the component (`components:write`); classification and filter on existing list/read endpoints; distribution added to the spec-214 statistics response.

### Persistence

Hosting column on the component read model; backfill migration defaults existing rows to `unknown`.

### Frontend

Components feature: classification control on the details panel, hosting facet on the list, distribution added to the statistics tile.

---

## Design Decisions

1. **Fixed enum in the AM domain, not MetaModel configuration** — hosting models are an industry-stable vocabulary; tenant configurability would add a MetaModel dependency for no expressed need. Alternative: MM-configured list (rejected as speculative).
2. **A value on the component, not a technology entity** — SD6 deliberately stops short of a technology portfolio; the enum is the B3 trailhead and nothing more.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Closed enum | A tenant wanting finer categories (e.g. PaaS) cannot add one | Revisit via a roadmap amendment if pulled; the enum is one value object |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
