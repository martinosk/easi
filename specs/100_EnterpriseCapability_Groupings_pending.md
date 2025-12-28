# Create Enterprise Capability Groupings

**Status**: pending

## User Value

> "As an enterprise architect, I want to create canonical capability groupings (e.g., 'Payroll') and link domain capabilities that represent the same logical capability, so I can discover overlap and track how many implementations exist across the organization."

### The Problem

Organizations often have siloed capability mapping per business domain with no standard naming:
- IT Support calls it "Payroll Management"
- Customer Service calls it "Salary Processing"
- Finance calls it "Compensation Admin"
- HR calls it "Pay & Benefits"

All four are the same logical capability, but there's no visibility into this overlap.

### The Solution

Enterprise Capabilities provide a **bottom-up discovery** mechanism:
1. Architect notices similar capabilities across domains
2. Creates an Enterprise Capability ("Payroll") as the canonical name
3. Links domain capabilities to it
4. System shows: "4 implementations across 4 domains"

## Dependencies

- Spec 098: Strategy Pillars (for enterprise-level importance)
- Spec 053: Business Domain Aggregate
- Spec 023: Capability Model

## Bounded Context

**Enterprise Architecture** - A new bounded context for cross-domain capability analysis.

See [Bounded Context Canvas](/docs/bounded-contexts/EnterpriseArchitecture.md) for full definition.

---

## Domain Model

### EnterpriseCapability Aggregate

Represents a logical capability that may exist across multiple business domains.

**Properties**:
- Name (required, unique per tenant, max 200 chars) - the canonical name
- Description (optional, max 1000 chars)
- Category (optional, max 100 chars) - for grouping
- Active flag (for soft delete)

**Commands**:
- Create enterprise capability
- Update enterprise capability
- Delete enterprise capability (soft delete)

**Business Rules**:
- Name must be unique within tenant (case-insensitive)
- Soft delete preserves links for historical analysis

### EnterpriseCapabilityLink Aggregate

Links a domain capability to an enterprise capability.

**Properties**:
- Enterprise Capability reference
- Domain Capability reference
- Linked timestamp
- Linked by (user)

**Commands**:
- Link capability to enterprise capability
- Unlink capability from enterprise capability

**Business Rules**:
- Enterprise capability must exist and be active
- Domain capability must exist
- **A domain capability can only be linked to ONE enterprise capability** (prevents confusion)

**Cascade Behavior**:
- When domain capability deleted → remove its link

### EnterpriseCapabilityStrategicImportance Aggregate

Rates how important an enterprise capability is for a strategy pillar (enterprise-wide perspective).

**Properties**:
- Enterprise Capability reference
- Pillar reference
- Importance (1-5 scale)
- Rationale (optional, max 500 chars)

**Commands**:
- Set importance
- Update importance
- Remove importance

**Business Rules**:
- Enterprise capability must exist
- Pillar must exist and be active
- One rating per (enterprise capability + pillar) combination

---

## User Experience

### Entry Point

New top-level navigation: **Enterprise Architecture**

### Main View: Enterprise Capabilities List

```
┌─────────────────────────────────────────────────────────────┐
│  Enterprise Architecture                                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Enterprise Capabilities                                    │
│  Discover and manage capability overlap across domains      │
│                                                             │
│  [+ New Enterprise Capability]       [Search...         🔍] │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Payroll                                             │    │
│  │ HR Operations                                       │    │
│  │ 4 implementations · 4 domains                       │    │
│  │ Maturity: Genesis → Product (spread: 50)            │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Customer Identity                                   │    │
│  │ Security                                            │    │
│  │ 2 implementations · 2 domains                       │    │
│  │ Maturity: Custom Built → Custom Built (spread: 10)  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Enterprise Capability Detail View

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back                                                     │
├─────────────────────────────────────────────────────────────┤
│  Payroll                                      [Edit] [🗑]   │
│  HR Operations                                              │
│                                                             │
│  Employee compensation processing across all domains        │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Summary                                                    │
│  ─────────────────────────────────────────────────────────  │
│  Implementations: 4  ·  Domains: 4                          │
│  Maturity Range: Genesis (15) → Product (65)                │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Linked Capabilities                          [+ Link]      │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ IT Support                                          │    │
│  │ └─ Payroll Management              [Genesis] [Unlink]   │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Customer Service                                    │    │
│  │ └─ Salary Processing               [Product] [Unlink]   │
│  ├─────────────────────────────────────────────────────┤    │
│  │ Finance                                             │    │
│  │ └─ Compensation Admin              [Custom]  [Unlink]   │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Strategic Importance                    [+ Rate Pillar]    │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Standardization       ★★★★★ Critical        [Edit]  │    │
│  │ Multiple implementations with varying quality.      │    │
│  │ Consolidation would reduce TCO by ~40%.             │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Link Capability Dialog

```
┌─────────────────────────────────────────────────────────────┐
│  Link Capability to Payroll                          [X]    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Select a domain capability that represents "Payroll"       │
│  in its local domain context.                               │
│                                                             │
│  [Search capabilities...                               🔍]  │
│                                                             │
│  Available Capabilities (not yet linked):                   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ ○ Operations                                        │    │
│  │   └─ Wage Administration          [Genesis]         │    │
│  │ ○ Support Services                                  │    │
│  │   └─ Employee Compensation        [Custom]          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  Note: A domain capability can only be linked to ONE        │
│  enterprise capability.                                     │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                              [Cancel]  [Link]               │
└─────────────────────────────────────────────────────────────┘
```

### Domain Capability Details Enhancement

Show enterprise link in domain capability details:

```
┌─────────────────────────────────────────────────────────────┐
│  Payroll Management                                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  Enterprise Capability                                      │
│  ─────────────────────────────────────────────────────────  │
│                                                             │
│  Linked to: Payroll ↗                                       │
│  (4 implementations across 4 domains)                       │
│                                                             │
│         - or if not linked -                                │
│                                                             │
│  Not linked to any enterprise capability                    │
│  [Link to Enterprise Capability]                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## API Requirements

**Enterprise Capability operations**:
- List all enterprise capabilities (with counts)
- Get enterprise capability with linked capabilities
- Create enterprise capability
- Update enterprise capability
- Delete enterprise capability (soft)

**Link operations**:
- List links for an enterprise capability
- Create link
- Delete link

**Strategic importance operations**:
- List importance ratings for enterprise capability
- Set/update/remove importance

**Discovery operations**:
- Check if a domain capability is linked (and to which enterprise capability)

**Permissions** (new):
- `PermEnterpriseArchitectureRead`: View enterprise capabilities
- `PermEnterpriseArchitectureWrite`: Create/update/delete/link

---

## Checklist

- [ ] Specification approved
- [ ] Bounded context scaffolding
- [ ] EnterpriseCapability aggregate
- [ ] EnterpriseCapabilityLink aggregate
- [ ] EnterpriseCapabilityStrategicImportance aggregate
- [ ] API implemented
- [ ] Enterprise Architecture page
- [ ] Capability details enhancement
- [ ] New permissions added
- [ ] Tests passing
- [ ] User sign-off
