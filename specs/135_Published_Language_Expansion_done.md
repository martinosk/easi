# Published Language Expansion

## Description
Expand the published language packages across all bounded contexts to expose the event constants and shared contracts needed for proper cross-context integration. This enables specs 136-138 to eliminate internal package imports.

## Purpose
Several BCs import internal packages from other BCs because no published language contract exists. The auth BC has no `publishedlanguage` package at all, forcing every other BC to import `auth/domain/valueobjects` and `auth/infrastructure/session`. The CapabilityMapping BC publishes only structural events but not data-change events (metadata updates, realizations, fit scores), preventing EnterpriseArchitecture from building proper ACL caches.

## Dependencies
- None (can be implemented in parallel with spec 134)

## Part 1: Auth Published Language

### Problem
Six bounded contexts import `auth/domain/valueobjects` for permission constants and `auth/infrastructure/session` for the SessionManager. These are internal packages.

### Solution
Create `backend/internal/auth/publishedlanguage/` with contracts that other BCs can import.

### Package: `auth/publishedlanguage/contracts.go`

```go
package publishedlanguage

// Permission represents an authorization permission.
type Permission string

const (
    PermArchModelRead    Permission = "arch:model:read"
    PermArchModelWrite   Permission = "arch:model:write"
    PermArchModelDelete  Permission = "arch:model:delete"
    PermArchViewRead     Permission = "arch:view:read"
    PermArchViewWrite    Permission = "arch:view:write"
    PermArchViewDelete   Permission = "arch:view:delete"
    PermCapMapRead       Permission = "cap:map:read"
    PermCapMapWrite      Permission = "cap:map:write"
    PermCapMapDelete     Permission = "cap:map:delete"
    PermEnterpriseArchRead  Permission = "ea:read"
    PermEnterpriseArchWrite Permission = "ea:write"
    PermEnterpriseArchDelete Permission = "ea:delete"
    PermMetaModelRead    Permission = "metamodel:read"
    PermMetaModelWrite   Permission = "metamodel:write"
    PermUserManage       Permission = "user:manage"
    PermSettingsManage   Permission = "settings:manage"
    PermImport           Permission = "import:execute"
    PermValueStreamsRead  Permission = "vs:read"
    PermValueStreamsWrite Permission = "vs:write"
    PermValueStreamsDelete Permission = "vs:delete"
)
```

**Note:** Copy the exact permission constants from `auth/domain/valueobjects/permission.go`. The published language becomes the **source of truth** — after creating this file, delete the `Permission` type and constants from `auth/domain/valueobjects/permission.go` entirely. Update all imports within the auth BC to use `auth/publishedlanguage` directly. Do NOT alias or re-export from the internal package, as that creates a circular import risk (`publishedlanguage` → `domain/valueobjects` → `publishedlanguage`).

### Package: `auth/publishedlanguage/session.go`

```go
package publishedlanguage

import (
    "context"
    "net/http"
)

// SessionProvider provides access to the current user's session.
// Implemented by auth/infrastructure/session.SessionManager.
type SessionProvider interface {
    GetCurrentUserID(ctx context.Context) (string, error)
    GetCurrentUserName(ctx context.Context) (string, error)
}

// AuthMiddleware provides HTTP middleware for permission-based access control.
type AuthMiddleware interface {
    RequirePermission(permission Permission) func(http.Handler) http.Handler
}
```

### Migration Steps

1. Create `backend/internal/auth/publishedlanguage/contracts.go` with Permission type and constants
2. Create `backend/internal/auth/publishedlanguage/session.go` with SessionProvider and AuthMiddleware interfaces
3. Delete the `Permission` type and constants from `auth/domain/valueobjects/permission.go`; update all auth-internal imports to use `auth/publishedlanguage` directly
4. Update `auth/infrastructure/session/session_manager.go` to implement the published language interface
5. Update all consuming BCs to import from `auth/publishedlanguage` instead of internal packages:
   - `enterprisearchitecture/infrastructure/api/routes.go`
   - `capabilitymapping/infrastructure/api/routes.go`
   - `architecturemodeling/infrastructure/api/routes.go`
   - `architectureviews/infrastructure/api/routes.go`
   - `metamodel/infrastructure/api/routes.go`
   - `valuestreams/infrastructure/api/routes.go`
   - `accessdelegation/infrastructure/api/routes.go`
   - `importing/infrastructure/api/routes.go`
   - `platform/infrastructure/api/routes.go` (if applicable)

### Verification
After migration, the following imports should NOT appear anywhere outside `auth/`:
- `"easi/backend/internal/auth/domain/valueobjects"`
- `"easi/backend/internal/auth/infrastructure/session"`

Remove the corresponding entries from the architecture guard test allowlist (spec 134).

## Part 2: CapabilityMapping Published Language Expansion

### Problem
EnterpriseArchitecture needs to build ACL cache tables (spec 136) by subscribing to CapabilityMapping events. Currently, CM's published language only publishes structural events (create/update/delete capabilities, parent/level changes, domain assignments). It does NOT publish:
- `CapabilityMetadataUpdated` — carries `maturityValue`, which EA needs for maturity analysis
- Realization events — EA needs to know when a component is linked to a capability
- Fit score events — EA needs application fit scores for TIME suggestions
- Effective importance events — EA needs importance scores for TIME suggestions

### Solution
Add missing event constants to `capabilitymapping/publishedlanguage/events.go`.

### Updated `capabilitymapping/publishedlanguage/events.go`

```go
package publishedlanguage

const (
    // Capability structure events (existing)
    CapabilityCreated              = "CapabilityCreated"
    CapabilityUpdated              = "CapabilityUpdated"
    CapabilityDeleted              = "CapabilityDeleted"
    CapabilityParentChanged        = "CapabilityParentChanged"
    CapabilityLevelChanged         = "CapabilityLevelChanged"
    CapabilityAssignedToDomain     = "CapabilityAssignedToDomain"
    CapabilityUnassignedFromDomain = "CapabilityUnassignedFromDomain"
    BusinessDomainDeleted          = "BusinessDomainDeleted"

    // Capability data events (NEW — needed by EA for ACL caches)
    CapabilityMetadataUpdated = "CapabilityMetadataUpdated"

    // Realization events (NEW — needed by EA for TIME suggestion cache)
    SystemRealizationCreated = "SystemRealizationCreated"
    SystemRealizationRemoved = "SystemRealizationRemoved"

    // Fit score events (NEW — needed by EA for TIME suggestion cache)
    ApplicationFitScoreSet     = "ApplicationFitScoreSet"
    ApplicationFitScoreRemoved = "ApplicationFitScoreRemoved"

    // Importance events (NEW — needed by EA for TIME suggestion cache)
    EffectiveImportanceRecalculated = "EffectiveImportanceRecalculated"
)
```

### Verification
- Ensure the event type strings match exactly what the CM aggregates produce. Check the `EventType()` method of each domain event in `capabilitymapping/domain/events/`.
- If some events don't exist yet as domain events (e.g., `EffectiveImportanceRecalculated` is a projection-level event), the projector should publish a synthetic event via the EventBus after recalculating. See Part 3.

## Part 3: Event Publication for Projection-Level Changes

### Problem
Some data that EA needs (effective_capability_importance, application_fit_scores) is calculated by CM projectors, not by domain aggregates. These projectors update read model tables but don't publish events. Without events, EA cannot build its own cache.

### Solution
CM projectors that update tables EA depends on must publish events after their updates. This follows the "projection change notification" pattern.

### Implementation

1. **CapabilityProjector** — After handling `CapabilityMetadataUpdated` and updating the `capabilities` table (which includes `maturity_value`), the event is already published by the aggregate. No additional work needed — EA just needs to subscribe to it.

2. **RealizationProjector** — After handling `SystemRealizationCreated`/`SystemRealizationRemoved` and updating `capability_realizations`, the events are already published by the aggregate. EA subscribes to these.

3. **Importance Projector** — The `effective_capability_importance` table is updated by a projector that calculates inherited importance. After recalculating, this projector should publish an `EffectiveImportanceRecalculated` event:
   ```go
   type EffectiveImportanceRecalculated struct {
       CapabilityID     string `json:"capabilityId"`
       BusinessDomainID string `json:"businessDomainId"`
       PillarID         string `json:"pillarId"`
       Importance       int    `json:"importance"`
   }
   ```
   The projector needs access to the EventBus to publish this.

4. **Fit Score Projector** — The `application_fit_scores` table is updated when fit scores are set. The events `ApplicationFitScoreSet` / `ApplicationFitScoreRemoved` are already published by the aggregate. EA subscribes to these.

### Cascading Event Safeguard

Publishing events from projectors is a pattern shift — normally only aggregates publish events. To prevent cascading projection loops (projector A publishes event → projector B handles it → publishes another event → ...):

**Rule:** Projection-published events MUST only be consumed by projectors in OTHER bounded contexts, never by projectors within the same BC. To enforce this:

1. **Naming convention:** All projection-notification events MUST use the suffix `Recalculated` (e.g., `EffectiveImportanceRecalculated`). This distinguishes them from domain events and signals "do not subscribe within this BC."

2. **Architecture guard test rule (spec 134):** Add a subtest to the Go import boundary test that scans event subscription wiring in each BC's `routes.go`. If a BC subscribes to a `*Recalculated` event that originates from its own published language, the test fails. This makes the rule compile-time enforceable, consistent with the overall approach.

3. **Stack overflow backstop:** The event bus is synchronous — a cascading loop would cause an immediate stack overflow during development. This provides a runtime safety net even if the test is bypassed.

### Projector EventBus Access

Projectors that need to publish events should receive the EventBus as a dependency:

```go
type EffectiveImportanceProjector struct {
    readModel *readmodels.EffectiveCapabilityImportanceReadModel
    eventBus  events.EventBus  // NEW dependency
}
```

The projector publishes after completing its update:

```go
func (p *EffectiveImportanceProjector) handleRecalculation(ctx context.Context, ...) error {
    // 1. Update the read model table
    err := p.readModel.UpsertEffectiveImportance(ctx, dto)
    if err != nil {
        return err
    }

    // 2. Publish notification for downstream consumers (EA)
    return p.eventBus.Publish(ctx, []domain.DomainEvent{
        NewEffectiveImportanceRecalculatedEvent(dto),
    })
}
```

## Part 4: StrategyPillarsGateway Interface Relocation

### Problem
EnterpriseArchitecture imports `capabilitymapping/infrastructure/metamodel` for the `StrategyPillarsGateway` interface. This is an infrastructure-level import across BC boundaries.

### Solution
The `StrategyPillarsGateway` interface and its DTOs describe a contract for accessing MetaModel configuration. This belongs in `metamodel/publishedlanguage/`, not in `capabilitymapping/infrastructure/`.

### Steps

1. Move the interface and DTOs to `metamodel/publishedlanguage/contracts.go`:
   ```go
   package publishedlanguage

   type StrategyPillarsConfigDTO struct { ... }
   type StrategyPillarDTO struct { ... }

   type StrategyPillarsGateway interface {
       GetStrategyPillars(ctx context.Context) (*StrategyPillarsConfigDTO, error)
       GetActivePillar(ctx context.Context, pillarID string) (*StrategyPillarDTO, error)
       InvalidateCache(tenantID string)
   }
   ```

2. Update all implementers:
   - `capabilitymapping/infrastructure/metamodel/local_strategy_pillars_gateway.go` — implements `mmPL.StrategyPillarsGateway`
   - `enterprisearchitecture/infrastructure/metamodel/local_strategy_pillars_gateway.go` — implements `mmPL.StrategyPillarsGateway`

3. Update all consumers to import from `metamodel/publishedlanguage` instead of `capabilitymapping/infrastructure/metamodel`.

4. Remove the interface and DTOs from `capabilitymapping/infrastructure/metamodel/strategy_pillars_gateway.go`. Keep only the implementation.

### Verification
After migration, the following import should NOT appear outside `capabilitymapping/`:
- `"easi/backend/internal/capabilitymapping/infrastructure/metamodel"`

Remove the corresponding entry from the architecture guard test allowlist.

## Files to Create

```
backend/internal/auth/publishedlanguage/contracts.go     # Permission type + constants
backend/internal/auth/publishedlanguage/session.go        # SessionProvider, AuthMiddleware interfaces
```

## Files to Modify

```
# Auth published language source of truth
backend/internal/auth/domain/valueobjects/permission.go   # Alias from publishedlanguage
backend/internal/auth/infrastructure/session/*.go          # Implement publishedlanguage interface

# CM published language expansion
backend/internal/capabilitymapping/publishedlanguage/events.go  # Add new event constants

# StrategyPillarsGateway relocation
backend/internal/metamodel/publishedlanguage/events.go     # Add contracts (or new contracts.go)
backend/internal/capabilitymapping/infrastructure/metamodel/strategy_pillars_gateway.go  # Remove interface, keep impl
backend/internal/capabilitymapping/infrastructure/metamodel/local_strategy_pillars_gateway.go  # Update import
backend/internal/enterprisearchitecture/infrastructure/metamodel/local_strategy_pillars_gateway.go  # Update import

# All consuming BCs — update imports
backend/internal/enterprisearchitecture/infrastructure/api/routes.go
backend/internal/capabilitymapping/infrastructure/api/routes.go
backend/internal/architecturemodeling/infrastructure/api/routes.go
backend/internal/architectureviews/infrastructure/api/routes.go
backend/internal/metamodel/infrastructure/api/routes.go
backend/internal/valuestreams/infrastructure/api/routes.go
backend/internal/accessdelegation/infrastructure/api/routes.go
backend/internal/importing/infrastructure/api/routes.go

# Projectors that need EventBus access
backend/internal/capabilitymapping/application/projectors/effective_importance_projector.go
```

## Success Criteria

- Auth's `publishedlanguage` package exists and exports Permission constants + SessionProvider interface
- No BC imports `auth/domain/valueobjects` or `auth/infrastructure/session` (except auth itself)
- CM's published language includes all event constants needed for EA ACL caches
- StrategyPillarsGateway interface lives in `metamodel/publishedlanguage`
- No BC imports `capabilitymapping/infrastructure/metamodel` (except capabilitymapping itself)
- Projection-level events are published for effective importance changes
- Architecture guard test (spec 134) allowlist entries for these imports are removed
- `go test ./...` passes
- `npm run build` passes (no frontend impact expected)

## Checklist

- [x] Specification approved
- [x] Auth publishedlanguage package created (contracts.go, session.go)
- [x] Auth internals refactored to use/alias published language
- [x] All BCs migrated to import from auth/publishedlanguage
- [x] CM publishedlanguage expanded with new event constants
- [x] Event type strings verified against domain event implementations
- [x] EffectiveImportanceRecalculated event publication added to projector
- [x] StrategyPillarsGateway interface moved to metamodel/publishedlanguage
- [x] All implementers and consumers updated
- [x] Architecture guard test allowlist entries removed for resolved imports
- [x] All tests passing
- [x] User sign-off
