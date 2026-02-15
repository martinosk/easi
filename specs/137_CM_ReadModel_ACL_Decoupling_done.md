# CM Read Model ACL Decoupling

## Description
Eliminate the cross-bounded-context SQL join in CapabilityMapping's `strategic_fit_analysis_read_model.go` which references EnterpriseArchitecture's `domain_capability_metadata` table.

## Purpose
The CapabilityMapping BC's strategic fit analysis joins EnterpriseArchitecture's `domain_capability_metadata` table to retrieve the effective business domain for each capability. This is a reverse cross-BC dependency: CM reaches into EA's ACL cache to get data that originates from CM's own domain. CM should maintain its own denormalized view of this data.

## Dependencies
- None (can be implemented independently of other specs)

## Current Violation

### File: `strategic_fit_analysis_read_model.go`

**Lines 88-91:**
```sql
LEFT JOIN domain_capability_metadata dcm ON r.tenant_id = dcm.tenant_id AND r.capability_id = dcm.capability_id
```

**Data used from `domain_capability_metadata`:**
- `dcm.business_domain_id` — the effective business domain this capability belongs to
- `dcm.business_domain_name` — the human-readable name

### Why CM Has This Data Already

The business domain assignment is a CM concept. CM owns:
- `capabilities` — the capabilities themselves
- `domain_capability_assignments` — which L1 capabilities are assigned to which business domains
- `business_domains` — the business domains

The `domain_capability_metadata` table in EA is a denormalized view that propagates business domain assignments from L1 capabilities down to their L2/L3/L4 descendants. CM doesn't have this propagation — it stores the raw assignment at L1 only.

## Solution: `cm_effective_business_domain` Cache Table

### Approach

Create a new CM-owned projection table that stores the effective business domain for every capability (not just L1). This table is populated by CM's own projectors reacting to CM's own events. It replaces the need to join EA's `domain_capability_metadata`.

### Migration

```sql
-- Migration: NNN_add_cm_effective_business_domain.sql (use next sequential number)

CREATE TABLE IF NOT EXISTS cm_effective_business_domain (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255),
    business_domain_name VARCHAR(500),
    l1_capability_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (tenant_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_cm_eff_bd_tenant_cap
    ON cm_effective_business_domain(tenant_id, capability_id);

ALTER TABLE cm_effective_business_domain ENABLE ROW LEVEL SECURITY;

CREATE POLICY cm_effective_business_domain_tenant_isolation ON cm_effective_business_domain
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

### Data Model

| Column | Type | Description |
|---|---|---|
| `tenant_id` | VARCHAR(50) | Tenant isolation |
| `capability_id` | VARCHAR(255) | The capability this row describes |
| `business_domain_id` | VARCHAR(255) | The effective business domain (inherited from L1 ancestor) |
| `business_domain_name` | VARCHAR(500) | Human-readable name |
| `l1_capability_id` | VARCHAR(255) | The L1 ancestor from which the BD is inherited |

### Read Model

**File:** `backend/internal/capabilitymapping/application/readmodels/cm_effective_business_domain_read_model.go`

```go
type CMEffectiveBusinessDomainReadModel struct {
    db *database.TenantAwareDB
}

func (rm *CMEffectiveBusinessDomainReadModel) Upsert(ctx context.Context, dto CMEffectiveBusinessDomainDTO) error { ... }
func (rm *CMEffectiveBusinessDomainReadModel) Delete(ctx context.Context, capabilityID string) error { ... }
func (rm *CMEffectiveBusinessDomainReadModel) UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, bd BusinessDomainRef) error { ... }
func (rm *CMEffectiveBusinessDomainReadModel) RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error { ... }
```

This follows the EXACT same logic as EA's `DomainCapabilityMetadataReadModel` for business domain propagation, but it lives in CM and operates on CM's own table.

### Projector

**File:** `backend/internal/capabilitymapping/application/projectors/effective_business_domain_projector.go`

Subscribes to CM's OWN domain events (not cross-BC):
- `CapabilityCreated` → Insert row; inherit BD from parent's L1 ancestor
- `CapabilityDeleted` → Delete row
- `CapabilityParentChanged` → Recalculate L1 and BD for subtree
- `CapabilityLevelChanged` → Update level (affects L1 calculation)
- `CapabilityAssignedToDomain` → Update BD for L1 subtree
- `CapabilityUnassignedFromDomain` → Clear BD for L1 subtree

The projector logic mirrors `DomainCapabilityMetadataProjector` in EA but is scoped to only BD propagation (no capability names, no maturity values).

### Fix `strategic_fit_analysis_read_model.go`

**Before:**
```sql
LEFT JOIN domain_capability_metadata dcm ON r.tenant_id = dcm.tenant_id AND r.capability_id = dcm.capability_id
```

**After:**
```sql
LEFT JOIN cm_effective_business_domain cmbd ON r.tenant_id = cmbd.tenant_id AND r.capability_id = cmbd.capability_id
```

Replace all references:
- `dcm.business_domain_id` → `cmbd.business_domain_id`
- `dcm.business_domain_name` → `cmbd.business_domain_name`

Remove the import of any EA packages (there should be none — the current file doesn't import EA Go packages, only references the EA table via SQL).

### Data Backfill

The backfill computes the effective business domain from CM's own source tables (`capabilities`, `domain_capability_assignments`, `business_domains`) rather than copying from EA's `domain_capability_metadata` cache. This avoids a dependency on EA's derived data, which could be stale or incomplete.

```sql
-- Backfill cm_effective_business_domain from CM's own source tables (one-time migration)
-- Step 1: For each capability, find its L1 ancestor by walking up the parent chain.
-- Step 2: Look up the L1's business domain assignment from domain_capability_assignments.
-- Step 3: Insert the effective business domain for every capability.

WITH RECURSIVE capability_tree AS (
    -- Base case: all capabilities with their direct parent
    SELECT id, tenant_id, parent_id, id AS root_l1_id, level
    FROM capabilities
    WHERE level = 'L1'

    UNION ALL

    -- Recursive: children inherit their L1 ancestor
    SELECT c.id, c.tenant_id, c.parent_id, ct.root_l1_id, c.level
    FROM capabilities c
    JOIN capability_tree ct ON c.parent_id = ct.id AND c.tenant_id = ct.tenant_id
    WHERE c.level != 'L1'
)
INSERT INTO cm_effective_business_domain (tenant_id, capability_id, business_domain_id, business_domain_name, l1_capability_id)
SELECT
    ct.tenant_id,
    ct.id,
    dca.business_domain_id,
    dca.business_domain_name,
    ct.root_l1_id
FROM capability_tree ct
LEFT JOIN domain_capability_assignments dca
    ON dca.capability_id = ct.root_l1_id AND dca.tenant_id = ct.tenant_id
ON CONFLICT DO NOTHING;
```

### Event Subscription Wiring

**File:** `backend/internal/capabilitymapping/infrastructure/api/routes.go`

Add the new projector to CM's own event subscriptions:

```go
effectiveBDProjector := projectors.NewEffectiveBusinessDomainProjector(rm.effectiveBusinessDomain)

for _, event := range []string{
    "CapabilityCreated",
    "CapabilityDeleted",
    "CapabilityParentChanged",
    "CapabilityLevelChanged",
    "CapabilityAssignedToDomain",
    "CapabilityUnassignedFromDomain",
} {
    eventBus.Subscribe(event, effectiveBDProjector)
}
```

Note: These are CM's own events, NOT cross-BC subscriptions. CM subscribes to its own domain events.

## Alternative Considered: Join Through `domain_capability_assignments`

An alternative would be to join `domain_capability_assignments` directly instead of creating a new table. However, `domain_capability_assignments` only stores assignments at the L1 level. To get the business domain for an L2/L3/L4 capability, we'd need a recursive CTE to find the L1 ancestor, then join to `domain_capability_assignments`. This is complex, slow, and fragile. The denormalized cache table is the proper solution — it's the same pattern EA uses for the same reason.

## Files to Create

```
backend/internal/capabilitymapping/application/readmodels/cm_effective_business_domain_read_model.go
backend/internal/capabilitymapping/application/projectors/effective_business_domain_projector.go
```

## Files to Modify

```
backend/internal/capabilitymapping/application/readmodels/strategic_fit_analysis_read_model.go   # Replace dcm join with cmbd
backend/internal/capabilitymapping/infrastructure/api/routes.go   # Wire new projector
backend/deploy-scripts/migrations/NNN_add_cm_effective_business_domain.sql   # New migration
```

## Testing Strategy

### Unit Tests
- Test `EffectiveBusinessDomainProjector` event handling for all 6 event types
- Test business domain propagation: L1 assignment propagates to L2/L3/L4 descendants
- Test parent change recalculates BD for entire subtree

### Integration Tests (require `integration` build tag + PostgreSQL)
- End-to-end: Create L1 → L2 → L3 hierarchy, assign L1 to domain → verify all have BD populated
- End-to-end: Query strategic fit analysis → verify results match pre-refactoring behavior
- End-to-end: Move L2 to different L1 → verify BD recalculated

### Verification
- Remove `capabilitymapping/strategic_fit_analysis_read_model.go -> domain_capability_metadata` from architecture SQL test allowlist (spec 134)
- Run `go test ./internal/ -run TestReadModelsOnlyReferenceOwnedTables` — must pass

## Success Criteria

- `strategic_fit_analysis_read_model.go` joins ZERO tables from EnterpriseArchitecture
- Business domain propagation works identically to EA's `domain_capability_metadata` behavior
- Strategic fit analysis API returns identical results before and after the change
- Architecture SQL guard test passes with CM allowlist entry removed
- Backfill migration correctly populates `cm_effective_business_domain` for all existing tenants

## Checklist

- [x] Specification approved
- [x] Migration: create `cm_effective_business_domain` table
- [x] Migration: backfill from CM's own source tables (capabilities, domain_capability_assignments)
- [x] Read model: `CMEffectiveBusinessDomainReadModel` implemented
- [x] Projector: `EffectiveBusinessDomainProjector` implemented with all 6 event handlers
- [x] `strategic_fit_analysis_read_model.go` rewritten (no cross-BC joins)
- [x] Event subscriptions wired in routes.go
- [x] Unit tests for projector
- [x] Integration tests for BD propagation
- [x] Architecture SQL guard test allowlist entry removed for CM
- [x] API behavior verified identical before/after
- [x] User sign-off
