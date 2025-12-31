# Enterprise Capability Linking UI

**Status**: pending

## User Value

> "As an enterprise architect, I want to drag domain capabilities onto enterprise capabilities to link them, with visual feedback showing what's already linked and what's blocked by hierarchy rules."

## Dependencies

- Spec 100: Enterprise Capability Groupings (done)
- Business Domains page components (for reuse)

---

## Domain Rules for Hierarchical Linking

### Linking Any Level

Domain capabilities are hierarchical (L1 → L2 → L3 → L4). Enterprise capabilities are flat. The system allows linking at **any hierarchy level** because:

- Different domains model at different granularities
- "Payroll" might be L2 in Finance but L3 in IT Support
- The architect's choice of link level is a business decision

### Parent-Child Conflict Rule

**Invariant**: If a capability is linked to an enterprise capability, its ancestors and descendants cannot be linked to a *different* enterprise capability.

```
ALLOWED:
  L1: Finance Ops
  └─ L2: Payroll Processing  [LINKED to "Payroll"]
     └─ L3: Tax Calculation  (available, not linked)

NOT ALLOWED:
  L1: Finance Ops           [LINKED to "Finance"]
  └─ L2: Payroll Processing [CANNOT link to "Payroll" - parent linked elsewhere]

ALSO NOT ALLOWED:
  L1: Finance Ops           (available)
  └─ L2: Payroll Processing [LINKED to "Payroll"]
     └─ L3: Tax Calculation [CANNOT link to "Tax Services" - ancestor linked elsewhere]
```

### Counting Rules

- `link_count`: Number of explicit links (not children)
- `domain_count`: `COUNT(DISTINCT business_domain_id)` from linked capabilities

---

## User Experience

### Two-Panel Drag/Drop Layout

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Enterprise Architecture                                                │
├─────────────────────────────────┬───────────────────────────────────────┤
│  Enterprise Capabilities        │  Domain Capabilities                  │
│  Drop zone for linking          │  Drag source                          │
│                                 │                                       │
│  [+ New Enterprise Capability]  │  Filter: [All Domains ▼] [Unlinked ○] │
│                                 │                                       │
│  ┌───────────────────────────┐  │  ┌─────────────────────────────────┐  │
│  │ 📦 Payroll            (3) │  │  │ IT Support                      │  │
│  │ HR Operations             │  │  │ └─ L1: Operations         [≡]   │  │
│  │                           │  │  │    └─ L2: Payroll Mgmt    [≡]   │  │
│  │ Drop capability here      │  │  │       └─ L3: Tax Calc           │  │
│  └───────────────────────────┘  │  │                                 │  │
│                                 │  │ Finance                         │  │
│  ┌───────────────────────────┐  │  │ └─ L1: Finance Ops        [≡]   │  │
│  │ 📦 Customer Identity  (2) │  │  │    └─ L2: Compensation    [≡]   │  │
│  │ Security                  │  │  │                                 │  │
│  └───────────────────────────┘  │  │ ════════════════════════════    │  │
│                                 │  │ Already Linked                  │  │
│  ┌───────────────────────────┐  │  │ ════════════════════════════    │  │
│  │ 📦 Order Management   (0) │  │  │ HR Domain                       │  │
│  │ Commerce                  │  │  │ └─ L2: Payroll Admin ──► Payroll│  │
│  │                           │  │  │                                 │  │
│  │ No linked capabilities    │  │  └─────────────────────────────────┘  │
│  └───────────────────────────┘  │                                       │
│                                 │  Legend:                              │
│                                 │  [≡] = Draggable                      │
│                                 │  Grayed = Blocked by hierarchy rule   │
│                                 │  ──► = Already linked to              │
└─────────────────────────────────┴───────────────────────────────────────┘
```

### Visual States for Domain Capabilities

| State | Visual Treatment |
|-------|------------------|
| **Available** | Normal color, drag handle visible `[≡]` |
| **Already linked** | Shows `──► {Enterprise Capability Name}`, not draggable |
| **Blocked by parent** | Grayed out, tooltip: "Parent linked to {name}" |
| **Blocked by child** | Grayed out, tooltip: "Child linked to {name}" |

### Drop Zone Feedback

When dragging over an enterprise capability card:
- Card highlights with dashed border
- Shows "Drop to link {capability name}"
- Invalid drops (already linked elsewhere) show error state

### Interaction Flow

1. User drags a capability (any level) from right panel
2. Hovers over enterprise capability card on left
3. Card shows drop highlight
4. User drops → API call to create link
5. Success: Capability moves to "Already Linked" section, counts update
6. Error: Toast notification with reason

### Click to View Details

Clicking an enterprise capability card opens a detail panel:

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Payroll                                                    [Edit] [🗑] │
│  HR Operations                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│  Employee compensation processing across all domains                    │
│                                                                         │
│  Summary                                                                │
│  ────────────────────────────────────────────────                       │
│  Implementations: 3  ·  Domains: 3                                      │
│  Maturity Range: Genesis (15) → Product (65)                            │
│                                                                         │
│  Linked Capabilities                                                    │
│  ────────────────────────────────────────────────                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │ IT Support                                                      │    │
│  │ └─ Payroll Management (L2)           [Genesis]        [Unlink]  │    │
│  ├─────────────────────────────────────────────────────────────────┤    │
│  │ Customer Service                                                │    │
│  │ └─ Salary Processing (L3)            [Product]        [Unlink]  │    │
│  ├─────────────────────────────────────────────────────────────────┤    │
│  │ Finance                                                         │    │
│  │ └─ Compensation Admin (L2)           [Custom]         [Unlink]  │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Strategic Importance                              [+ Rate Pillar]      │
│  ────────────────────────────────────────────────                       │
│  (No ratings yet)                                                       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Component Reuse from Business Domains

### Reusable Components

| Component | Current Location | Reuse Strategy |
|-----------|-----------------|----------------|
| `CapabilityExplorer` | `business-domains/components/` | **Extend**: Add `dragLevels` prop to control which levels are draggable (currently L1 only) |
| `CapabilityExplorerSidebar` | `business-domains/components/` | **Wrap**: Create `EnterpriseCapabilityExplorer` that configures it for all-level dragging |
| `DomainFilter` | `business-domains/components/` | **Reuse directly**: Filter by business domain |
| `StrategicImportanceSection` | `business-domains/components/` | **Adapt**: Change from domain capability to enterprise capability context |
| `SetImportanceDialog` | `business-domains/components/` | **Adapt**: Same as above |

### New Components Needed

| Component | Purpose |
|-----------|---------|
| `EnterpriseCapabilityCard` | Drop target card showing name, category, counts |
| `EnterpriseCapabilityList` | Left panel with cards and new capability button |
| `EnterpriseCapabilityDetail` | Detail panel with linked capabilities and importance |
| `LinkedCapabilityItem` | Row in detail panel showing domain, capability name, maturity, unlink button |

### CapabilityExplorer Extension

Current behavior: Only L1 items are draggable.

Proposed change:
```tsx
interface CapabilityExplorerProps {
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  isLoading: boolean;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
  // NEW PROPS:
  draggableLevels?: ('L1' | 'L2' | 'L3' | 'L4')[]; // Default: ['L1']
  linkedCapabilities?: Map<CapabilityId, string>; // capId -> enterprise capability name
  blockedCapabilities?: Set<CapabilityId>; // blocked by parent/child rule
}
```

---

## API Requirements

### New Endpoints

**Check link eligibility** (for visual feedback):
```
GET /api/v1/capabilities/{id}/enterprise-link-status
Response:
{
  "capabilityId": "...",
  "status": "available" | "linked" | "blocked_by_parent" | "blocked_by_child",
  "linkedTo": { "id": "...", "name": "Payroll" } | null,
  "blockingCapability": { "id": "...", "name": "..." } | null
}
```

**Batch check** (for initial load):
```
GET /api/v1/capabilities/enterprise-link-status?domainId={domainId}
Response:
{
  "data": [
    { "capabilityId": "...", "status": "available" },
    { "capabilityId": "...", "status": "linked", "linkedTo": { "id": "...", "name": "Payroll" } },
    { "capabilityId": "...", "status": "blocked_by_parent", "blockingCapability": { "id": "...", "name": "..." } }
  ]
}
```

### Existing Endpoints (from Spec 100)

- `POST /api/v1/enterprise-capabilities/{id}/links` - Create link
- `DELETE /api/v1/enterprise-capabilities/{id}/links/{linkId}` - Remove link
- `GET /api/v1/enterprise-capabilities/{id}/links` - List links

---

## Backend Changes

### Parent-Child Validation

Add validation in `LinkDomainCapability` handler:

1. Query capability's ancestors (via `parentId` chain)
2. Query capability's descendants (via recursive query)
3. Check if any ancestor or descendant is linked to a *different* enterprise capability
4. Reject with appropriate error if conflict found

### Read Model Enhancement

Update `EnterpriseCapabilityLinkDTO` to include:
- `businessDomainId`
- `businessDomainName`
- `capabilityLevel` (L1/L2/L3/L4)
- `capabilityMaturity`

---

## Checklist

- [ ] Specification approved
- [ ] Extend `CapabilityExplorer` with `draggableLevels` prop
- [ ] Add link status indicators (linked, blocked)
- [ ] Create `EnterpriseCapabilityCard` component
- [ ] Create `EnterpriseCapabilityList` with drop zones
- [ ] Implement two-panel layout
- [ ] Create `EnterpriseCapabilityDetail` panel
- [ ] Add parent-child validation to backend
- [ ] Add link eligibility API endpoint
- [ ] Add strategic importance UI (reuse from business domains)
- [ ] Tests passing
- [ ] User sign-off
