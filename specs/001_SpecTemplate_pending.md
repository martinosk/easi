# NNN — Short Description

> **Status:** pending | ongoing | done | superseded
> **Depends on:** _(link to prerequisite specs, if any)_

---

## Problem Statement

_What problem does this solve, and why does it matter now? 1–3 paragraphs. Reference the user or system impact._

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Role** | What they need to accomplish |

---

## User-Facing Behavior (BDD Scenarios)

_One `Scenario` per distinct observable behavior. These are the acceptance tests — every scenario here must have a corresponding test._

```gherkin
Feature: Short feature name

  Scenario: Happy path
    Given ...
    When ...
    Then ...

  Scenario: Edge case or error case
    Given ...
    When ...
    Then ...
```

---

## Business Rules & Invariants

_Number every rule. Treat each as a test case._

1. **Rule name** — description
2. **Rule name** — description

---

## Acceptance Criteria

_Measurable, pass/fail conditions. Derived from the BDD scenarios and business rules above._

- [ ] Criterion with a clear pass condition
- [ ] Criterion with a clear pass condition

---

## Architecture

_High-level design intent. Describe where the change fits and what constraints apply. Leave implementation details (SQL, component names, file paths) to the implementer._

### Ownership

_Which bounded context owns this change? Which existing contexts are affected and how (read-only reference, event subscription, shared read model)?_

### Domain Model

_New or modified aggregates, entities, and value objects. What are the key invariants? What domain events does this produce or consume?_

### API Surface

_New or changed API capabilities at a contract level — resources exposed, operations permitted, permission model. Not a full endpoint table._

### Persistence

_What data needs to survive? Any notable consistency or isolation requirements (e.g., tenant isolation, eventual consistency across contexts)._

### Frontend

_Which views or user journeys are affected? What new capabilities does the UI need to expose?_

### Cross-Context Integration

_Which other bounded contexts are affected? What events flow between them and in which direction?_

---

## Design Decisions

_Number each decision. Include the rationale and alternatives rejected._

1. **Decision** — rationale. Alternatives considered: _X_ (rejected because _Y_).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Choice made | Cost incurred | How it is managed |

---

## Checklist

- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
