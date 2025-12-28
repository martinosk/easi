# Discover and Analyze Standardization Opportunities

**Status**: pending

## User Value

> "As an enterprise architect, I want to see which enterprise capabilities have multiple implementations with varying maturity levels, so I can identify standardization opportunities and prioritize investments for consolidation."

> "As a portfolio manager, I want a dashboard showing capabilities marked for standardization with their maturity gaps, so I can make evidence-based investment decisions."

## Dependencies

- Spec 100: Enterprise Capability Groupings
- Spec 098: Strategy Pillars

---

## Domain Concepts

### Standardization Candidate

An enterprise capability that:
- Has importance rated for standardization pillar
- Has 2+ linked domain capabilities (implementations)
- Shows maturity variance across implementations

### Maturity Gap

The difference between a domain capability's current maturity and the target maturity (highest implementation).

### Investment Priority

Derived from:
- Strategic importance rating
- Maturity gap size
- Number of implementations

---

## User Experience

### Entry Point

Enterprise Architecture → Standardization Analysis (tab)

### Standardization Candidates Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│  Enterprise Architecture                                    │
├─────────────────────────────────────────────────────────────┤
│  [Enterprise Capabilities] [Standardization Analysis]       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Standardization Candidates                                 │
│  Enterprise capabilities marked for standardization         │
│  with multiple implementations                              │
│                                                             │
│  Summary: 8 candidates · 24 implementations · Avg gap: 35   │
│                                                             │
│  Sort by: [Importance ▼]                                    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                                                     │    │
│  │  PAYROLL                              ★★★★★        │    │
│  │  HR Operations                                      │    │
│  │  4 implementations · 4 domains                      │    │
│  │                                                     │    │
│  │  Maturity Distribution:                             │    │
│  │  ▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░        │    │
│  │  Genesis (2)  Custom (1)  Product (1)               │    │
│  │                                                     │    │
│  │  Gap: 50 points (Genesis → Product)                 │    │
│  │                                                     │    │
│  │  [View Details]                                     │    │
│  │                                                     │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │                                                     │    │
│  │  INVOICE PROCESSING                   ★★★★☆        │    │
│  │  Finance Operations                                 │    │
│  │  3 implementations · 2 domains                      │    │
│  │                                                     │    │
│  │  Gap: 75 points (Genesis → Commodity)               │    │
│  │                                                     │    │
│  │  [View Details]                                     │    │
│  │                                                     │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Maturity Gap Detail View

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back to Standardization Analysis                         │
├─────────────────────────────────────────────────────────────┤
│  Payroll · Maturity Gap Analysis                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Standardization Importance: ★★★★★ Critical                 │
│  Multiple implementations with varying quality.             │
│  Consolidation would reduce TCO by ~40%.                    │
│                                                             │
│  Target Maturity: Product (65)                              │
│  (Derived from highest implementation)                      │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Implementation Comparison                                  │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  Customer Service  ████████████████████████████░░░░░ 65     │
│  Finance           ██████████████████░░░░░░░░░░░░░░░ 45     │
│  IT Support        ██████░░░░░░░░░░░░░░░░░░░░░░░░░░░ 15     │
│  HR                ████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 10     │
│                    ├──────────────────────────────────┤     │
│                    0         50        100   Target: 65     │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Investment Priority                                        │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  High Priority (gap > 40):                                  │
│    • IT Support: Payroll Mgmt (gap: 50)                     │
│    • HR: Pay & Benefits (gap: 55)                           │
│                                                             │
│  Medium Priority (gap 15-40):                               │
│    • Finance: Compensation Admin (gap: 20)                  │
│                                                             │
│  On Target:                                                 │
│    • Customer Service: Salary Processing                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Unlinked Capabilities View

Help discover capabilities that might need enterprise grouping:

```
┌─────────────────────────────────────────────────────────────┐
│  Enterprise Architecture                                    │
├─────────────────────────────────────────────────────────────┤
│  [Enterprise Capabilities] [Standardization] [Unlinked]     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Unlinked Capabilities                                      │
│  Domain capabilities not linked to any enterprise           │
│  capability - potential overlap to discover                 │
│                                                             │
│  Total: 47 capabilities across 8 domains                    │
│                                                             │
│  Filter by domain: [All ▼]   Search: [_____________ 🔍]     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ IT Support                                          │    │
│  │   • Help Desk Management         [Custom]           │    │
│  │   • Incident Tracking            [Product]          │    │
│  │   • Asset Management             [Genesis]          │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Finance                                             │    │
│  │   • Budget Planning              [Custom]           │    │
│  │   • Expense Reporting            [Commodity]        │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## API Requirements

**Analysis operations**:
- List standardization candidates (with summary stats, sorting)
- Get maturity gap analysis for enterprise capability
- List unlinked capabilities (with domain filter, search)

**Derived calculations**:
- Target maturity = max maturity of linked capabilities
- Gap = target - current maturity
- Investment priority = High (gap > 40), Medium (15-40), Low (1-14), None (0)
- Maturity distribution = count per maturity section

---

## Checklist

- [ ] Specification approved
- [ ] Standardization candidates query
- [ ] Maturity gap analysis query
- [ ] Unlinked capabilities query
- [ ] Dashboard UI
- [ ] Gap detail view
- [ ] Unlinked capabilities view
- [ ] Tests passing
- [ ] User sign-off
