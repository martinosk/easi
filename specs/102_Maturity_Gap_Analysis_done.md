# Maturity Gap Analysis

**Status**: done

## User Value

> "As an enterprise architect, I want to see which enterprise capabilities have multiple implementations with varying maturity levels, so I can identify standardization opportunities and prioritize investments for consolidation."

> "As a portfolio manager, I want a dashboard showing capabilities with maturity gaps, so I can make evidence-based investment decisions."

## Dependencies

- Spec 100: Enterprise Capability Groupings

---

## Domain Concepts

### Target Maturity

A desired maturity level (0-99) set on an enterprise capability indicating where all implementations should converge.

### Maturity Gap Candidate

An enterprise capability that:
- Has 2+ linked domain capabilities (implementations)
- Shows maturity variance across implementations

### Maturity Gap

The difference between a domain capability's current maturity and the target maturity (or highest implementation if no target set).

### Investment Priority

Derived from gap size:
- High: gap > 40
- Medium: gap 15-40
- Low: gap 1-14
- None: gap = 0

---

## User Experience

### Entry Point

Enterprise Architecture page with three tabs:
- Enterprise Capabilities
- Maturity Analysis
- Unlinked Capabilities

### Maturity Analysis Dashboard

Shows enterprise capabilities with 2+ implementations, their maturity distribution, and gap metrics.

Features:
- Summary stats (candidate count, total implementations, average gap)
- Sort by gap or implementation count
- Maturity distribution bar for each capability
- View Details action to drill into gap analysis

### Maturity Gap Detail View

For a selected enterprise capability:
- Target maturity display with set/edit action
- Implementation comparison with horizontal bars
- Implementations grouped by investment priority (High/Medium/Low/On Target)

### Unlinked Capabilities View

Domain capabilities not linked to any enterprise capability:
- Filter by business domain
- Search by name
- Grouped by business domain
- Shows maturity section for each

---

## API Endpoints

### Set Target Maturity
`PUT /enterprise-capabilities/{id}/target-maturity`

### Get Maturity Analysis Candidates
`GET /enterprise-capabilities/maturity-analysis`
- Query: `sortBy` (gap | implementations)

### Get Maturity Gap Detail
`GET /enterprise-capabilities/{id}/maturity-gap`

### Get Unlinked Capabilities
`GET /domain-capabilities/unlinked`
- Query: `businessDomainId`, `search`

---

## Implementation Notes

- Target maturity is stored on enterprise_capabilities table (integer 0-99, nullable)
- Gap calculation uses target maturity if set, otherwise max implementation maturity
- Maturity sections: Genesis (0-24), Custom Build (25-49), Product (50-74), Commodity (75-99)
- Investment priority thresholds: High (>40), Medium (15-40), Low (1-14), None (0)

---

## Checklist

- [x] Target maturity field on enterprise capabilities
- [x] Maturity analysis candidates query
- [x] Maturity gap analysis query
- [x] Unlinked capabilities query
- [x] Maturity Analysis tab UI
- [x] Gap detail view with set target modal
- [x] Unlinked capabilities view
- [x] Tests passing
