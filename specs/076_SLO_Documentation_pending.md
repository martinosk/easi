# SLO Documentation

## Status
**Pending**

## User Need
Product teams lack a structured way to establish Service Level Objectives with business stakeholders. There is no shared language or process for discussing baseline performance, setting targets, and acknowledging the cost tradeoffs of improved service levels. Teams need a tool that facilitates this conversation and tracks the maturity of each SLO agreement.

**Architectural insight:** SLO documentation enables analysis of whether monoliths should be split. When one application realizes capabilities with vastly different SLOs (e.g., 99.99% availability for payments vs 95% for reporting), it reveals over-engineering: the entire application must meet the highest SLO, even for capabilities that don't need it. This is a concrete, data-driven argument for architectural decomposition decisions.

## Dependencies
- Capability Realizations (existing) - SLOs attach to the relationship between an application and a capability

---

## Success Criteria

- SLOs can be created on capability realizations (app + capability combination)
- Each SLO has a type, target value, owner, and maturity state
- Maturity states reflect the conversation journey: Proposed, Agreed, Tracking, Active
- SLOs are visible when viewing capability realizations
- Different stakeholders can see SLO status across the landscape

---

## Vertical Slices

### Slice 1: Create and Manage SLOs on Capability Realizations

Enable users to define SLOs with maturity tracking on existing capability realizations.

**Backend:**
- [ ] Create SLO aggregate with: type, target value, owner, maturity state
- [ ] SLO types constrained to: availability, latency_p99, error_rate, throughput
- [ ] Maturity states: proposed, agreed, tracking, active
- [ ] SLO references a capability realization by ID
- [ ] Events: SLOCreated, SLOUpdated, SLODeleted, SLOMaturityChanged
- [ ] Read model for querying SLOs by capability realization
- [ ] REST API for CRUD operations on SLOs

**Frontend:**
- [ ] Add SLO to a capability realization from the UI
- [ ] Edit SLO properties (type, target, owner)
- [ ] Change maturity state (dropdown or explicit progression)
- [ ] Delete SLO
- [ ] Display SLOs when viewing a capability realization

### Slice 2: SLO Variance Analysis (Future)

Provide views that surface applications with high SLO variance across their realized capabilities, enabling data-driven architectural decisions about monolith decomposition.

Example query: "Show me apps where capability SLOs vary by more than one 9 of availability."

*Details to be defined when this slice is prioritized.*

**API Endpoints (Slice 1):**
```
POST   /api/v1/capability-realizations/{id}/slos
GET    /api/v1/capability-realizations/{id}/slos
GET    /api/v1/slos/{id}
PUT    /api/v1/slos/{id}
PATCH  /api/v1/slos/{id}/maturity
DELETE /api/v1/slos/{id}
```

---

## Out of Scope (for now)

- Actual/measured SLO values (handled by external monitoring)
- Historical tracking of maturity state changes
- Aggregated views/dashboards across domains
- SLI definitions (indicators that measure the SLO)
- Alerting or status indicators (red/yellow/green)
- Bulk operations on SLOs
- SLOs directly on applications without capability context

---

## Domain Model

**SLO Aggregate:**
- ID (value object)
- CapabilityRealizationID (reference)
- Type (value object: availability | latency_p99 | error_rate | throughput)
- TargetValue (structured value object, type-specific):
  - availability: numeric percentage (e.g., 99.9)
  - latency_p99: numeric value + unit (e.g., 200, "ms")
  - error_rate: numeric percentage (e.g., 0.1)
  - throughput: numeric value + unit (e.g., 1000, "req/s")
- Owner (value object: string)
- MaturityState (value object: proposed | agreed | tracking | active)

**Invariants:**
- Type must be one of the allowed values
- Maturity state must be one of the allowed values
- Target value must be non-empty
- Owner must be non-empty

---

## Acceptance Criteria

- [ ] User can create an SLO on a capability realization
- [ ] User can specify type from constrained list (availability, latency_p99, error_rate, throughput)
- [ ] User can set target value as structured data (numeric value, optional unit based on type)
- [ ] User can assign an owner
- [ ] User can set initial maturity state (any of the four states)
- [ ] User can change maturity state at any time
- [ ] User can edit SLO properties
- [ ] User can delete an SLO
- [ ] SLOs appear when viewing capability realization details
- [ ] API returns appropriate HTTP status codes per CLAUDE.md conventions

---

## Checklist
- [x] Specification ready
- [x] User sign-off on spec
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] Final user sign-off
