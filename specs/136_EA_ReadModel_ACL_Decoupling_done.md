# EA Read Model ACL Decoupling

## Description
Eliminate all cross-bounded-context SQL joins from EnterpriseArchitecture read models by replacing direct table access with event-driven ACL cache tables. After this spec, every EA read model queries only tables owned by the EA bounded context.

## Purpose
The EnterpriseArchitecture BC currently contains the most severe cross-BC coupling in the codebase: its read models directly join tables from CapabilityMapping (`capabilities`, `capability_realizations`, `effective_capability_importance`, `application_fit_scores`, `domain_capability_assignments`). This violates bounded context isolation and means schema changes in CM can silently break EA.

## Dependencies
- Spec 135: Published Language Expansion (CM must publish `CapabilityMetadataUpdated`, `SystemRealizationCreated/Removed`, `ApplicationFitScoreSet/Removed`, `EffectiveImportanceRecalculated` events)

## Current State: Violations

### Violation 1: `maturity_analysis_read_model.go`
**Foreign tables:** `capabilities` (CM)
**Data needed:** `maturity_value` per linked domain capability
**Lines:** 114, 207, 235-237, 291-296

### Violation 2: `time_suggestion_read_model.go`
**Foreign tables:** `capability_realizations`, `capabilities`, `effective_capability_importance`, `application_fit_scores` (all CM)
**Data needed:** Realization links, capability names, importance scores, fit scores
**Lines:** 155-189

### Violation 3: `domain_capability_metadata_read_model.go`
**Foreign tables:** `domain_capability_assignments` (CM)
**Data needed:** Business domain name lookup
**Lines:** 394-399

## Solution Architecture

### Principle
EA already follows the ACL cache pattern for some data: `domain_capability_metadata` is a cache of CM's capability structure, populated by the `DomainCapabilityMetadataProjector`. We extend this pattern to cover ALL data that EA needs from CM.

### New/Extended ACL Cache Tables

| Table | Purpose | Source Events | Existing? |
|---|---|---|---|
| `domain_capability_metadata` | Capability structure + maturity | CM: CapabilityCreated, Updated, Deleted, ParentChanged, LevelChanged, AssignedToDomain, UnassignedFromDomain, **CapabilityMetadataUpdated** | YES (extend with `maturity_value`) |
| `ea_realization_cache` | Component-capability links | CM: SystemRealizationCreated, SystemRealizationRemoved | NEW |
| `ea_importance_cache` | Effective importance per capability per pillar | CM: EffectiveImportanceRecalculated | NEW |
| `ea_fit_score_cache` | Fit scores per component per pillar | CM: ApplicationFitScoreSet, ApplicationFitScoreRemoved | NEW |

## Part 1: Extend `domain_capability_metadata` with `maturity_value`

### Migration

```sql
-- Migration: NNN_add_maturity_to_domain_capability_metadata.sql (use next sequential number)

ALTER TABLE domain_capability_metadata
    ADD COLUMN IF NOT EXISTS maturity_value INTEGER NOT NULL DEFAULT 0;
```

### Projector Update

Update `DomainCapabilityMetadataProjector` to:
1. Subscribe to `CapabilityMetadataUpdated` event
2. Update `maturity_value` when a capability's maturity changes

**File:** `backend/internal/enterprisearchitecture/application/projectors/domain_capability_metadata_projector.go`

Add to the `ProjectEvent` handlers map:
```go
cmPL.CapabilityMetadataUpdated: p.handleCapabilityMetadataUpdated,
```

Add handler:
```go
type capabilityMetadataUpdatedEvent struct {
    ID            string `json:"id"`
    MaturityValue int    `json:"maturityValue"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityMetadataUpdated(ctx context.Context, eventData []byte) error {
    var event capabilityMetadataUpdatedEvent
    if err := json.Unmarshal(eventData, &event); err != nil {
        return err
    }
    return p.metadataReadModel.UpdateMaturityValue(ctx, event.ID, event.MaturityValue)
}
```

**File:** `backend/internal/enterprisearchitecture/application/readmodels/domain_capability_metadata_read_model.go`

Add method:
```go
func (rm *DomainCapabilityMetadataReadModel) UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error {
    return rm.execForTenant(ctx,
        `UPDATE domain_capability_metadata SET maturity_value = $2 WHERE tenant_id = $1 AND capability_id = $3`,
        maturityValue, capabilityID,
    )
}
```

### Event Subscription

**File:** `backend/internal/enterprisearchitecture/infrastructure/api/routes.go`

Add `cmPL.CapabilityMetadataUpdated` to the `subscribeCapabilityMappingEvents` list.

### Fix `maturity_analysis_read_model.go`

Replace all joins to `capabilities c` with joins to `domain_capability_metadata dcm` using the new `maturity_value` column.

**Before (line 114):**
```sql
LEFT JOIN capabilities c ON ecl.domain_capability_id = c.id AND ecl.tenant_id = c.tenant_id
```

**After:**
```sql
-- No separate join needed; dcm already joined, use dcm.maturity_value
```

Update all references from `c.maturity_value` to `dcm.maturity_value` throughout the file:
- `buildCandidatesQuery` (lines 107-109): `MAX(c.maturity_value)` → `MAX(dcm.maturity_value)`, etc.
- `getMaturityDistribution` (lines 200-208): `c.maturity_value` → `dcm.maturity_value`
- `GetMaturityGapDetail` (lines 234-237): `c.maturity_value` → `dcm.maturity_value`
- `getImplementations` (lines 291-296): `c.maturity_value` → `dcm.maturity_value`

Remove the `LEFT JOIN capabilities c` clause entirely.

### Fix `LookupBusinessDomainName` in `domain_capability_metadata_read_model.go`

**Before (line 397):**
```sql
SELECT business_domain_name FROM domain_capability_assignments WHERE tenant_id = $1 AND business_domain_id = $2 LIMIT 1
```

**After:**
```sql
SELECT business_domain_name FROM domain_capability_metadata WHERE tenant_id = $1 AND business_domain_id = $2 AND business_domain_name IS NOT NULL LIMIT 1
```

This queries EA's own cache table instead of CM's `domain_capability_assignments` table. The `domain_capability_metadata` table already has `business_domain_name` populated by the projector when capabilities are assigned to domains.

## Part 2: New `ea_realization_cache` Table

### Migration

```sql
-- In the same migration file: NNN_add_ea_acl_cache_tables.sql (use next sequential number)

CREATE TABLE IF NOT EXISTS ea_realization_cache (
    tenant_id VARCHAR(50) NOT NULL,
    realization_id VARCHAR(255) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    component_name VARCHAR(500) NOT NULL,
    origin VARCHAR(50) NOT NULL DEFAULT 'Direct',
    PRIMARY KEY (tenant_id, realization_id)
);

CREATE INDEX IF NOT EXISTS idx_ea_realization_cache_tenant_cap
    ON ea_realization_cache(tenant_id, capability_id);
CREATE INDEX IF NOT EXISTS idx_ea_realization_cache_tenant_comp
    ON ea_realization_cache(tenant_id, component_id);

ALTER TABLE ea_realization_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_realization_cache_tenant_isolation ON ea_realization_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

### Read Model

**File:** `backend/internal/enterprisearchitecture/application/readmodels/ea_realization_cache_read_model.go`

```go
type EARealizationCacheReadModel struct {
    db *database.TenantAwareDB
}

func (rm *EARealizationCacheReadModel) Upsert(ctx context.Context, dto EARealizationCacheDTO) error { ... }
func (rm *EARealizationCacheReadModel) Delete(ctx context.Context, realizationID string) error { ... }
func (rm *EARealizationCacheReadModel) DeleteByCapabilityID(ctx context.Context, capabilityID string) error { ... }
```

### Projector

**File:** `backend/internal/enterprisearchitecture/application/projectors/ea_realization_cache_projector.go`

Subscribes to:
- `cmPL.SystemRealizationCreated` → Insert into `ea_realization_cache`
- `cmPL.SystemRealizationRemoved` → Delete from `ea_realization_cache`
- `cmPL.CapabilityDeleted` → Delete all realizations for that capability
- `amPL.ApplicationComponentUpdated` → Update `component_name` for all rows matching `component_id`

**Important:** The cache stores `component_name` for query convenience. When a component is renamed in Architecture Modeling, the cached name would go stale. The projector subscribes to AM's published language `ApplicationComponentUpdated` event to keep the name current:

```go
func (p *EARealizationCacheProjector) handleComponentUpdated(ctx context.Context, eventData []byte) error {
    var event struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }
    if err := json.Unmarshal(eventData, &event); err != nil {
        return err
    }
    return p.readModel.UpdateComponentName(ctx, event.ID, event.Name)
}
```

The read model method:
```go
func (rm *EARealizationCacheReadModel) UpdateComponentName(ctx context.Context, componentID string, name string) error {
    return rm.execForTenant(ctx,
        `UPDATE ea_realization_cache SET component_name = $2 WHERE tenant_id = $1 AND component_id = $3`,
        name, componentID,
    )
}
```

Local deserialization structs:
```go
type systemRealizationCreatedEvent struct {
    ID            string `json:"id"`
    CapabilityID  string `json:"capabilityId"`
    ComponentID   string `json:"componentId"`
    ComponentName string `json:"componentName"`
    Origin        string `json:"origin"`
}
```

## Part 3: New `ea_importance_cache` Table

### Migration

```sql
CREATE TABLE IF NOT EXISTS ea_importance_cache (
    tenant_id VARCHAR(50) NOT NULL,
    capability_id VARCHAR(255) NOT NULL,
    business_domain_id VARCHAR(255) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    effective_importance INTEGER NOT NULL DEFAULT 0,
    importance_label VARCHAR(100) NOT NULL DEFAULT '',
    source_capability_id VARCHAR(255) NOT NULL DEFAULT '',
    source_capability_name VARCHAR(500) NOT NULL DEFAULT '',
    is_inherited BOOLEAN NOT NULL DEFAULT false,
    rationale TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, capability_id, business_domain_id, pillar_id)
);

CREATE INDEX IF NOT EXISTS idx_ea_importance_cache_tenant_cap
    ON ea_importance_cache(tenant_id, capability_id);

ALTER TABLE ea_importance_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_importance_cache_tenant_isolation ON ea_importance_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

### Projector

**File:** `backend/internal/enterprisearchitecture/application/projectors/ea_importance_cache_projector.go`

Subscribes to:
- `cmPL.EffectiveImportanceRecalculated` → Upsert into `ea_importance_cache`

The event payload must carry all fields needed by the cache (capability_id, business_domain_id, pillar_id, effective_importance, label, source info, is_inherited, rationale).

## Part 4: New `ea_fit_score_cache` Table

### Migration

```sql
CREATE TABLE IF NOT EXISTS ea_fit_score_cache (
    tenant_id VARCHAR(50) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    pillar_id VARCHAR(255) NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,
    score_label VARCHAR(100) NOT NULL DEFAULT '',
    rationale TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, component_id, pillar_id)
);

CREATE INDEX IF NOT EXISTS idx_ea_fit_score_cache_tenant_comp
    ON ea_fit_score_cache(tenant_id, component_id);

ALTER TABLE ea_fit_score_cache ENABLE ROW LEVEL SECURITY;

CREATE POLICY ea_fit_score_cache_tenant_isolation ON ea_fit_score_cache
    FOR ALL TO easi_app
    USING (tenant_id = current_setting('app.current_tenant', true))
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true));
```

### Projector

**File:** `backend/internal/enterprisearchitecture/application/projectors/ea_fit_score_cache_projector.go`

Subscribes to:
- `cmPL.ApplicationFitScoreSet` → Upsert into `ea_fit_score_cache`
- `cmPL.ApplicationFitScoreRemoved` → Delete from `ea_fit_score_cache`

## Part 5: Rewrite `time_suggestion_read_model.go`

### New Query (replaces `buildGapsQuery`)

The rewritten query joins ONLY EA-owned tables:

```sql
SELECT
    dcm.capability_id,
    dcm.capability_name,
    erc.component_id,
    erc.component_name,
    eic.pillar_id,
    eic.effective_importance as importance,
    efsc.score as fit_score
FROM ea_realization_cache erc
JOIN domain_capability_metadata dcm
    ON dcm.capability_id = erc.capability_id AND dcm.tenant_id = erc.tenant_id
JOIN ea_importance_cache eic
    ON eic.capability_id = erc.capability_id AND eic.tenant_id = erc.tenant_id
    AND eic.business_domain_id = dcm.business_domain_id
JOIN ea_fit_score_cache efsc
    ON efsc.component_id = erc.component_id AND efsc.tenant_id = erc.tenant_id
    AND efsc.pillar_id = eic.pillar_id
WHERE erc.tenant_id = $1
    AND erc.origin = 'Direct'
    AND eic.effective_importance > 0
    AND efsc.score > 0
```

All tables referenced (`ea_realization_cache`, `domain_capability_metadata`, `ea_importance_cache`, `ea_fit_score_cache`) are owned by EA.

### Structural Changes
- Remove import of `capabilitymapping/infrastructure/metamodel` — the `StrategyPillarsGateway` is now imported from `metamodel/publishedlanguage` (spec 135)
- The rest of the read model logic (calculating suggestions, separating gaps by fit type) remains unchanged

## Eventual Consistency

The ACL cache pattern introduces eventual consistency between CM source data and EA caches. When a user creates or modifies data in CM (e.g., sets a fit score, creates a realization), there is a brief window before the corresponding EA cache is updated via the event bus.

**Impact assessment:** This is acceptable for this codebase because:
1. The event bus is in-process and synchronous — events are dispatched and handled within the same HTTP request transaction. There is no message queue or async delay.
2. Users interact with CM and EA through separate UI pages. By the time a user navigates from the CM page to the EA TIME suggestions page, the projectors have already processed the events.
3. The pre-existing behavior already had this characteristic: CM projectors that populate `capability_realizations` and `effective_capability_importance` were already eventually consistent with the aggregate state.

**If the event bus becomes asynchronous in the future**, a cache-warming strategy should be considered (e.g., a "last projected" timestamp check on read, with a synchronous fallback to source tables if the cache is stale). This is NOT needed now.

## Part 6: Data Migration (Backfill)

### Problem
Existing tenants have data in CM's tables but the new EA cache tables are empty. A one-time backfill is needed.

### Approach: Event Replay

Create a migration that backfills the new cache tables from the existing CM source tables. This is acceptable as a one-time migration; going forward, the projectors keep the caches in sync.

```sql
-- Backfill ea_realization_cache from capability_realizations
INSERT INTO ea_realization_cache (tenant_id, realization_id, capability_id, component_id, component_name, origin)
SELECT tenant_id, id, capability_id, component_id, component_name, origin
FROM capability_realizations
ON CONFLICT DO NOTHING;

-- Backfill maturity_value in domain_capability_metadata from capabilities
UPDATE domain_capability_metadata dcm
SET maturity_value = COALESCE(
    (SELECT c.maturity_value FROM capabilities c WHERE c.id = dcm.capability_id AND c.tenant_id = dcm.tenant_id),
    0
);

-- Backfill ea_importance_cache from effective_capability_importance
INSERT INTO ea_importance_cache (tenant_id, capability_id, business_domain_id, pillar_id, effective_importance, importance_label, source_capability_id, source_capability_name, is_inherited, rationale)
SELECT tenant_id, capability_id, business_domain_id, pillar_id, effective_importance, COALESCE(importance_label, ''), COALESCE(source_capability_id, ''), COALESCE(source_capability_name, ''), COALESCE(is_inherited, false), COALESCE(rationale, '')
FROM effective_capability_importance
ON CONFLICT DO NOTHING;

-- Backfill ea_fit_score_cache from application_fit_scores
INSERT INTO ea_fit_score_cache (tenant_id, component_id, pillar_id, score, score_label, rationale)
SELECT tenant_id, component_id, pillar_id, score, COALESCE(score_label, ''), COALESCE(rationale, '')
FROM application_fit_scores
ON CONFLICT DO NOTHING;
```

**Important:** This migration runs ONCE. After deployment, the projectors maintain the caches via events.

**Migration ordering:** This backfill reads directly from CM's source tables (`capability_realizations`, `effective_capability_importance`, `application_fit_scores`). The migration file number must be sequenced so it runs while all tables are still accessible in the same schema. If schema-level isolation is ever introduced in the future, this migration must have already executed.

## Part 7: Event Subscription Wiring

**File:** `backend/internal/enterprisearchitecture/infrastructure/api/routes.go`

### Updated `initializeReadModels`

```go
func initializeReadModels(db *database.TenantAwareDB) *routeReadModels {
    pillarCache := readmodels.NewStrategyPillarCacheReadModel(db)
    pillarsGateway := metamodel.NewLocalStrategyPillarsGateway(pillarCache)
    realizationCache := readmodels.NewEARealizationCacheReadModel(db)   // NEW
    importanceCache := readmodels.NewEAImportanceCacheReadModel(db)     // NEW
    fitScoreCache := readmodels.NewEAFitScoreCacheReadModel(db)         // NEW
    return &routeReadModels{
        // ... existing read models ...
        realizationCache: realizationCache,  // NEW
        importanceCache:  importanceCache,   // NEW
        fitScoreCache:    fitScoreCache,     // NEW
    }
}
```

### New Event Subscriptions

```go
func subscribeRealizationCacheEvents(eventBus events.EventBus, projector *projectors.EARealizationCacheProjector) {
    for _, event := range []string{
        cmPL.SystemRealizationCreated,
        cmPL.SystemRealizationRemoved,
        cmPL.CapabilityDeleted,
        amPL.ApplicationComponentUpdated, // Keep cached component_name fresh
    } {
        eventBus.Subscribe(event, projector)
    }
}

func subscribeImportanceCacheEvents(eventBus events.EventBus, projector *projectors.EAImportanceCacheProjector) {
    eventBus.Subscribe(cmPL.EffectiveImportanceRecalculated, projector)
}

func subscribeFitScoreCacheEvents(eventBus events.EventBus, projector *projectors.EAFitScoreCacheProjector) {
    for _, event := range []string{
        cmPL.ApplicationFitScoreSet,
        cmPL.ApplicationFitScoreRemoved,
    } {
        eventBus.Subscribe(event, projector)
    }
}
```

## Files to Create

```
backend/internal/enterprisearchitecture/application/readmodels/ea_realization_cache_read_model.go
backend/internal/enterprisearchitecture/application/readmodels/ea_importance_cache_read_model.go
backend/internal/enterprisearchitecture/application/readmodels/ea_fit_score_cache_read_model.go
backend/internal/enterprisearchitecture/application/projectors/ea_realization_cache_projector.go
backend/internal/enterprisearchitecture/application/projectors/ea_importance_cache_projector.go
backend/internal/enterprisearchitecture/application/projectors/ea_fit_score_cache_projector.go
backend/deploy-scripts/migrations/NNN_add_ea_acl_cache_tables.sql
```

## Files to Modify

```
backend/internal/enterprisearchitecture/application/readmodels/domain_capability_metadata_read_model.go  # Add maturity_value, fix LookupBusinessDomainName
backend/internal/enterprisearchitecture/application/readmodels/maturity_analysis_read_model.go  # Remove join to capabilities
backend/internal/enterprisearchitecture/application/readmodels/time_suggestion_read_model.go  # Rewrite to use cache tables
backend/internal/enterprisearchitecture/application/projectors/domain_capability_metadata_projector.go  # Add CapabilityMetadataUpdated handler
backend/internal/enterprisearchitecture/infrastructure/api/routes.go  # Wire new projectors and subscriptions
```

## Testing Strategy

### Unit Tests
- Test each new projector's event handling (given event → expected cache state)
- Test rewritten read model queries return correct results with cache data

### Integration Tests (require `integration` build tag + PostgreSQL)
- End-to-end: Create a capability in CM with metadata → verify `domain_capability_metadata.maturity_value` is populated in EA
- End-to-end: Create a realization in CM → verify `ea_realization_cache` is populated
- End-to-end: Set fit scores in CM → verify `ea_fit_score_cache` is populated
- End-to-end: Query TIME suggestions → verify results match pre-refactoring behavior
- End-to-end: Query maturity analysis → verify results match pre-refactoring behavior
- End-to-end: Rename a component in AM → verify `ea_realization_cache.component_name` is updated

### Verification
After implementation:
- Remove ALL entries from architecture SQL test allowlist (spec 134) for `enterprisearchitecture`
- Run `go test ./internal/ -run TestReadModelsOnlyReferenceOwnedTables` — must pass with no allowlist entries for EA

## Success Criteria

- `maturity_analysis_read_model.go` joins ZERO tables from CapabilityMapping
- `time_suggestion_read_model.go` joins ZERO tables from CapabilityMapping
- `domain_capability_metadata_read_model.go` references ZERO tables from CapabilityMapping
- All EA read models query only tables in the `enterprisearchitecture` ownership group
- TIME suggestion API returns identical results before and after the change
- Maturity analysis API returns identical results before and after the change
- Data backfill migration populates all cache tables for existing tenants
- Architecture SQL guard test passes with EA allowlist entries removed

## Implementation Notes (deviations from spec)

1. **Simplified `ea_importance_cache` schema**: Spec called for `importance_label`, `source_capability_id`, `source_capability_name`, `is_inherited`, `rationale` columns. Implementation only includes `effective_importance` (the only column consumed by queries). Extra columns can be added later if needed.

2. **Simplified `ea_fit_score_cache` schema**: Spec called for `score_label` column. Implementation omits it — not consumed by any current query.

3. **RLS policies**: Implementation uses separate SELECT + INSERT policies (matching existing project pattern) instead of the single FOR ALL policy suggested in spec.

4. **Realization event naming**: Spec referenced `SystemRealizationCreated`/`SystemRealizationRemoved` but implementation uses `SystemLinkedToCapability`/`SystemRealizationDeleted` (the actual published language event names from spec 135).

## Checklist

- [x] Specification approved
- [x] Migration: extend `domain_capability_metadata` with `maturity_value`
- [x] Migration: create `ea_realization_cache` table
- [x] Migration: create `ea_importance_cache` table
- [x] Migration: create `ea_fit_score_cache` table
- [x] Migration: backfill all cache tables from existing data
- [x] Read model: `EARealizationCacheReadModel` implemented
- [x] Read model: `EAImportanceCacheReadModel` implemented
- [x] Read model: `EAFitScoreCacheReadModel` implemented
- [x] Projector: `EARealizationCacheProjector` implemented
- [x] Projector: `EAImportanceCacheProjector` implemented
- [x] Projector: `EAFitScoreCacheProjector` implemented
- [x] Projector: `DomainCapabilityMetadataProjector` extended with CapabilityMetadataUpdated
- [x] `maturity_analysis_read_model.go` rewritten (no cross-BC joins)
- [x] `time_suggestion_read_model.go` rewritten (no cross-BC joins)
- [x] `domain_capability_metadata_read_model.go` LookupBusinessDomainName fixed
- [x] Event subscriptions wired in routes.go
- [x] Unit tests for all new projectors
- [x] Integration tests for end-to-end data flow
- [x] Architecture SQL guard test allowlist entries removed for EA
- [x] API behavior verified identical before/after (integration tests confirm read models return correct data from EA cache tables)
- [x] User sign-off
