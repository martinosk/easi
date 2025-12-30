# Create Enterprise Capability Groupings

**Status**: done

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
- `enterprise-arch:read`: View enterprise capabilities
- `enterprise-arch:write`: Create/update/link
- `enterprise-arch:delete`: Delete enterprise capabilities

---

## Implemented in This Spec

### Backend
- Bounded context scaffolding (`internal/enterprisearchitecture/`)
- EnterpriseCapability aggregate with TDD tests
- EnterpriseCapabilityLink aggregate with TDD tests
- EnterpriseStrategicImportance aggregate with TDD tests
- All command handlers and projectors
- Database migrations for all tables
- REST API endpoints with HATEOAS links
- New permissions added to role system

### Frontend
- Enterprise Architecture navigation item (permission-gated)
- Basic list view with create/delete capabilities
- Route `/enterprise-architecture`

---

## Deferred to Spec 101

The following UX features are deferred to a follow-up spec:
- Drag/drop linking interface
- Enterprise capability detail view
- Linked capabilities display with unlink
- Strategic importance UI
- Domain capability details enhancement (show enterprise link)
- Parent/child conflict validation for hierarchical linking

---

## Checklist

- [x] Specification approved
- [x] Bounded context scaffolding
- [x] EnterpriseCapability aggregate
- [x] EnterpriseCapabilityLink aggregate
- [x] EnterpriseCapabilityStrategicImportance aggregate
- [x] API implemented
- [x] Enterprise Architecture page (basic list view)
- [x] New permissions added
- [x] Tests passing
- [ ] Drag/drop linking UI (see Spec 101)
- [ ] Enterprise capability detail view (see Spec 101)
- [ ] Domain capability details enhancement (see Spec 101)
