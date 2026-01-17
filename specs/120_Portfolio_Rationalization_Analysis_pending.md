# Portfolio Rationalization Analysis

**Status**: pending

**Series**: Application Landscape & TIME (4 of 4)
- Spec 117: Portfolio Metadata Foundation
- Spec 118: Pillar Fit Type & TIME Suggestions
- Spec 119: TIME Classification & Application Landscape
- **Spec 120**: Portfolio Rationalization Analysis (this spec)

## User Value

> "As an enterprise architect, I want to identify enterprise capabilities that have no designated standard (no INVEST), so I can prioritize standardization decisions."

> "As a program manager planning consolidation, I want to see exactly which domain architects need to coordinate for a given consolidation initiative, so I can set up the right meetings."

> "As a domain architect, I want proactive visibility into cross-domain consolidation initiatives that affect my domain, so I'm not surprised by enterprise decisions."

## Dependencies

- Spec 117: Portfolio Metadata Foundation (provides domain architect)
- Spec 119: TIME Classification & Application Landscape (provides TIME classifications)

---

## Domain Concepts

### Standardization Gap

An enterprise capability that has realizations but **no application marked INVEST**. This indicates:
- Multiple applications serve this capability
- No standard has been designated
- A strategic decision is needed

### Consolidation Coordination

For a given enterprise capability with mixed TIME classifications:
- **Target state**: The INVEST application(s) - the standard
- **Migration sources**: Applications marked MIGRATE - need to move to standard
- **Elimination candidates**: Applications marked ELIMINATE - need to retire
- **Affected domains**: Business domains that use any of these applications
- **Contacts**: Domain architects who need to coordinate

---

## User Experience

### Standardization Gaps View

Analysis view showing enterprise capabilities with no INVEST:

```
┌─────────────────────────────────────────────────────────────────┐
│  Standardization Gaps                                            │
│  Enterprise capabilities with no designated standard (INVEST)    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Found 4 enterprise capabilities needing standardization         │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ EXPENSE MANAGEMENT                               [View] │    │
│  │ 3 realizations: 2 TOLERATE, 1 MIGRATE, 0 INVEST        │    │
│  │ No application marked INVEST - needs standard decision │    │
│  │ Domains: Finance, Procurement                          │    │
│  │ Contacts: Alice Smith, David Chen                      │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ DOCUMENT MANAGEMENT                              [View] │    │
│  │ 4 realizations: 1 TOLERATE, 2 MIGRATE, 1 ELIMINATE     │    │
│  │ No application marked INVEST - needs standard decision │    │
│  │ Domains: Legal, HR, Finance                            │    │
│  │ Contacts: Grace Lee, Bob Johnson, Alice Smith          │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ TIME TRACKING                                    [View] │    │
│  │ 2 realizations: 0 classified                           │    │
│  │ No TIME classifications yet                            │    │
│  │ Domains: HR, Operations                                │    │
│  │ Contacts: Bob Johnson, Carol Williams                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Consolidation Coordination View

Detail view for planning consolidation of a specific enterprise capability:

```
┌─────────────────────────────────────────────────────────────────┐
│  Consolidation Planning: PAYROLL PROCESSING                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Target State                                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ SAP HR                                          INVEST  │    │
│  │ Acquired via: TechCorp · Built by: TechCorp Eng         │    │
│  │ "Strategic platform for all payroll"                    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Migration Required                                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Legacy Payroll System                          MIGRATE  │    │
│  │ Built by: Finance IT                                    │    │
│  │ "Still needed for union rules, migrate Q3"              │    │
│  │ Used by: Operations                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Elimination Planned                                             │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Spreadsheet Tracker                           ELIMINATE │    │
│  │ Built by: Finance IT                                    │    │
│  │ "Shadow IT, no longer needed once SAP rolled out"       │    │
│  │ Used by: Finance                                        │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Coordination Required                                           │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Domain           Architect         Applications         │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ Finance          Alice Smith       SAP HR (INVEST)      │    │
│  │                                    Spreadsheet (ELIM)   │    │
│  │ HR               Bob Johnson       SAP HR (INVEST)      │    │
│  │ Operations       Carol Williams    Legacy (MIGRATE)     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Summary: 3 domain architects need to coordinate                 │
│  Action: Migrate Operations payroll to SAP HR,                   │
│          Eliminate Finance spreadsheet tracker                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Navigation

Access these views from:
1. Enterprise Architecture → new "Analysis" dropdown or sub-tabs
2. Direct link from Application Landscape when viewing an enterprise capability

---

## API Requirements

All endpoints follow REST Level 3 with HATEOAS. Responses include `_links` for navigation.

### Standardization Gaps
- `GET /standardization-gaps` - Enterprise capabilities with no INVEST
  - Returns list with:
    - enterpriseCapabilityId, name
    - realizationCount
    - timeBreakdown: { invest, tolerate, migrate, eliminate, unclassified }
    - affectedDomains: [{ domainId, domainName, domainArchitectId, domainArchitectName }]

### Consolidation Contacts
- `GET /consolidation-contacts/by-enterprise-capability/{id}` - Coordination details for an enterprise capability
  - Returns:
    - enterpriseCapability: { id, name }
    - targetState: [{ componentId, componentName, origins, rationale }] (INVEST apps)
    - migrationSources: [{ componentId, componentName, origins, rationale, domains }]
    - eliminationCandidates: [{ componentId, componentName, origins, rationale, domains }]
    - coordinationMatrix: [{ domainId, domainName, domainArchitectId, domainArchitectName, applications }]
  - Where `origins` = { acquisitions: [...], vendors: [...], teams: [...] }

---

## Read Models

### StandardizationGapsReadModel

Projection that identifies gaps:
```
enterpriseCapabilityId: EnterpriseCapabilityId
enterpriseCapabilityName: string
realizationCount: number
hasInvest: boolean
timeBreakdown: {
  invest: number
  tolerate: number
  migrate: number
  eliminate: number
  unclassified: number
}
affectedDomains: [{
  domainId: DomainId
  domainName: string
  domainArchitectId: UUID | null
  domainArchitectName: string | null  // resolved from user service, "Unknown user" if deleted
}]
```

Updated when:
- RealizationDispositionSet
- RealizationDispositionRemoved
- CapabilityLinkedToEnterpriseCapability
- CapabilityUnlinkedFromEnterpriseCapability
- BusinessDomainUpdated (architect change)

### ConsolidationContactsReadModel

Query-time aggregation (or materialized view) combining:
- Enterprise capability details
- All realizations with their TIME classifications
- Domain membership of each capability
- Domain architect for each domain

---

## Checklist

- [ ] Specification approved
- [ ] StandardizationGapsReadModel projector
- [ ] Standardization gaps API endpoint
- [ ] Standardization Gaps UI view
- [ ] ConsolidationContactsReadModel (query or projection)
- [ ] Consolidation contacts API endpoint
- [ ] Consolidation Coordination UI view
- [ ] Navigation integration (tabs/links)
- [ ] Tests passing
- [ ] User sign-off
