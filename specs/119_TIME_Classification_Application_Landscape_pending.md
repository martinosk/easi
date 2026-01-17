# TIME Classification & Application Landscape

**Status**: pending

**Series**: Application Landscape & TIME (3 of 4)
- Spec 117: Portfolio Metadata Foundation
- Spec 118: Pillar Fit Type & TIME Suggestions
- **Spec 119**: TIME Classification & Application Landscape (this spec)
- Spec 120: Portfolio Rationalization Analysis

## User Value

> "As an enterprise architect, I want to review TIME suggestions, override them with my own judgment, and record my rationale, so TIME classifications become official decisions rather than just suggestions."

> "As a domain architect, I want to see the TIME classifications for applications in my domain, so I understand the enterprise-level strategy and can plan accordingly."

> "As a CIO, I want a portfolio dashboard showing how many applications are marked INVEST vs TOLERATE vs MIGRATE vs ELIMINATE, so I can communicate our rationalization posture to the board."

## Dependencies

- Spec 117: Portfolio Metadata Foundation (provides origin, domain architect)
- Spec 118: Pillar Fit Type & TIME Suggestions (provides suggested TIME)

---

## Domain Concepts

### Responsibility Split

| Level | Owner | Concern |
|-------|-------|---------|
| Domain | Domain Architect | "How well are apps serving THIS domain?" (Strategic Fit) |
| Enterprise | Enterprise Architect | "How should we rationalize across ALL domains?" (TIME) |

Strategic Fit analysis operates at domain level. TIME operates at enterprise level. They complement each other:
- Strategic Fit provides the **evidence** (pillar gaps)
- TIME is the **verdict** (what to do about it)

### TIME Classification

TIME is set at the **enterprise capability + application** level:
- "SAP HR's realization of Enterprise Capability 'Payroll Processing' is INVEST"
- "Legacy Payroll's realization of Enterprise Capability 'Payroll Processing' is MIGRATE"

The same application can have different TIME classifications for different enterprise capabilities.

### Suggested vs Actual TIME

- **Suggested TIME**: Calculated by the system (Spec 118)
- **Actual TIME**: Set by the architect, may differ from suggestion
- **Rationale**: Why the architect chose this classification

**Key insight**: An application with TIME = INVEST is implicitly the **standard** for that enterprise capability. No separate "standard designation" is needed.

---

## Data Model

### RealizationDisposition (Enterprise Architecture - new aggregate)

```
enterpriseCapabilityId: EnterpriseCapabilityId
componentId: ComponentId
actualTime: TIME
rationale: string (max 500 chars)
decidedBy: string
decidedAt: timestamp
```

TIME enum: `TOLERATE | INVEST | MIGRATE | ELIMINATE`

Uniqueness: One disposition per (enterpriseCapabilityId, componentId) pair.

### Application Landscape Read Model

Combines disposition with suggestion and metadata:
```
enterpriseCapabilityId: EnterpriseCapabilityId
enterpriseCapabilityName: string
componentId: ComponentId
componentName: string
origins: {
  acquisitions: [{entityId, entityName, acquisitionDate}]
  vendors: [{vendorId, vendorName}]
  teams: [{teamId, teamName, department}]
}
suggestedTime: TIME | null
actualTime: TIME | null
rationale: string | null
decidedBy: string | null
decidedAt: timestamp | null
affectedDomains: [{domainId, domainName, domainArchitectId, domainArchitectName}]
```

---

## User Experience

### Application Landscape Tab

New tab in Enterprise Architecture page:

```
┌─────────────────────────────────────────────────────────────────┐
│  Enterprise Architecture                                         │
├─────────────────────────────────────────────────────────────────┤
│  [Capabilities] [Maturity] [Unlinked] [Strategic Fit] [Landscape]│
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Application Landscape                                           │
│  TIME classification for enterprise capability realizations      │
│                                                                  │
│  Filter: [All TIME ▼] [All Acquisitions ▼] [All Vendors ▼]      │
│                                                                  │
│  Summary:                                                        │
│  ┌────────┬────────┬─────────┬───────────┬──────────────┐       │
│  │ INVEST │TOLERATE│ MIGRATE │ ELIMINATE │ Unclassified │       │
│  │   15   │    8   │   12    │     5     │      7       │       │
│  └────────┴────────┴─────────┴───────────┴──────────────┘       │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Enterprise Capability: PAYROLL PROCESSING                       │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ SAP HR                                                   │    │
│  │ Acquired via: TechCorp · Built by: TechCorp Eng          │    │
│  │ Suggested: INVEST                                        │    │
│  │ Actual: [INVEST ▼]                              [Save]  │    │
│  │ Rationale: "Strategic platform for all payroll"         │    │
│  │ Domain contacts: Alice (Finance), Bob (HR)              │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ Legacy Payroll System                                    │    │
│  │ Built by: Finance IT                                     │    │
│  │ Suggested: ELIMINATE                                     │    │
│  │ Actual: [MIGRATE ▼]                             [Save]  │    │
│  │ Rationale: "Still needed for union rules, migrate Q3"   │    │
│  │ Domain contacts: Carol (Operations)                      │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Enterprise Capability: CUSTOMER ONBOARDING                      │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ CRM System                                               │    │
│  │ Purchased from: Salesforce                               │    │
│  │ Suggested: TOLERATE                                      │    │
│  │ Actual: [-- Select --▼]                         [Save]  │    │
│  │ Rationale: ___________                                   │    │
│  │ Domain contacts: Eve (Sales), Frank (Marketing)         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Inline Classification

When architect selects a TIME classification:
1. Dropdown shows: INVEST, TOLERATE, MIGRATE, ELIMINATE
2. Rationale field becomes editable
3. Save button commits the decision
4. On save: records decidedBy (current user), decidedAt (now)

### Filters

- **TIME filter**: All, INVEST, TOLERATE, MIGRATE, ELIMINATE, Unclassified
- **Acquisition filter**: All, or specific AcquiredEntity (e.g., "TechCorp", "AcmeCo")
- **Vendor filter**: All, or specific Vendor (e.g., "SAP", "Salesforce")

---

## API Requirements

All endpoints follow REST Level 3 with HATEOAS. Responses include `_links` for navigation.

### TIME Classification (Enterprise Architecture)
- `GET /realization-dispositions` - Get all dispositions
- `GET /realization-dispositions/by-enterprise-capability/{id}` - Filter by enterprise capability
- `GET /realization-dispositions/by-component/{id}` - Filter by component
- `GET /realization-dispositions/by-time/{time}` - Filter by TIME classification
- `PUT /realization-dispositions/{enterpriseCapabilityId}/{componentId}` - Set/update disposition
  - Body: `{ actualTime, rationale }`
- `DELETE /realization-dispositions/{enterpriseCapabilityId}/{componentId}` - Remove disposition

### Application Landscape (Enterprise Architecture)
- `GET /application-landscape` - Aggregated view
  - Returns: summary counts + list grouped by enterprise capability
  - Includes: suggested TIME, actual TIME, origins, affected domains with architects

---

## Events

### New Events (Enterprise Architecture)
- `RealizationDispositionSet` - TIME classification set or updated
  - Payload: enterpriseCapabilityId, componentId, actualTime, rationale, decidedBy, decidedAt
- `RealizationDispositionRemoved` - TIME classification removed
  - Payload: enterpriseCapabilityId, componentId

---

## Checklist

- [ ] Specification approved
- [ ] RealizationDisposition aggregate (domain model)
- [ ] RealizationDispositionSet event
- [ ] RealizationDispositionRemoved event
- [ ] Set/update disposition command handler
- [ ] Remove disposition command handler
- [ ] Application Landscape read model
- [ ] Application Landscape API endpoint
- [ ] Disposition CRUD API endpoints
- [ ] Application Landscape UI tab
- [ ] Summary counts display
- [ ] Inline TIME classification (dropdown + rationale + save)
- [ ] Filters (TIME, Acquisition, Vendor)
- [ ] Tests passing
- [ ] User sign-off
