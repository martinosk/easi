# Published Language Event Constants

**Status**: done

## Description

Introduce a `publishedlanguage` package per publishing bounded context containing event type string constants. Consuming contexts import these constants for event subscription and handler dispatch instead of duplicating magic strings. Consumers keep their own local deserialization structs (the existing Anti-Corruption Layer pattern).

## Purpose

Event type strings like `"StrategyPillarAdded"` are duplicated across 5+ files with no compile-time safety. A typo silently breaks event routing. Meanwhile, the local deserialization structs in consuming projectors are a correct ACL pattern and should be preserved.

This spec formalizes the Published Language relationship already documented in the architecture README.

## Scope

### In scope
- Create `publishedlanguage/events.go` for each publishing context with event type constants
- Migrate all cross-context event subscriptions and handler dispatch maps to use these constants
- Migrate Pattern A usages (direct domain event imports) to Pattern B (local structs + published constants)

### Out of scope
- Shared event DTOs/structs (explicitly avoided to prevent Shared Kernel coupling)
- A single shared events package (would couple every context to every other)
- Intra-context usage (contexts can continue using their own domain event types internally)

## Architecture

### Package Structure

```
backend/internal/metamodel/publishedlanguage/
    events.go

backend/internal/capabilitymapping/publishedlanguage/
    events.go

backend/internal/architecturemodeling/publishedlanguage/
    events.go
```

Each file contains only typed string constants. No structs, no constructors, no logic.

### Example: MetaModel Published Language

```go
package publishedlanguage

const (
    MetaModelConfigurationCreated = "MetaModelConfigurationCreated"
    StrategyPillarAdded           = "StrategyPillarAdded"
    StrategyPillarUpdated         = "StrategyPillarUpdated"
    StrategyPillarRemoved         = "StrategyPillarRemoved"
    PillarFitConfigurationUpdated = "PillarFitConfigurationUpdated"
    MaturityScaleConfigUpdated    = "MaturityScaleConfigUpdated"
    MaturityScaleConfigReset      = "MaturityScaleConfigReset"
)
```

### Consumer Usage

Subscription in routes:
```go
import mmEvents "easi/backend/internal/metamodel/publishedlanguage"

func subscribeMetaModelEvents(eventBus events.EventBus, projector *projectors.StrategyPillarCacheProjector) {
    for _, event := range []string{
        mmEvents.MetaModelConfigurationCreated,
        mmEvents.StrategyPillarAdded,
        mmEvents.StrategyPillarUpdated,
        mmEvents.StrategyPillarRemoved,
        mmEvents.PillarFitConfigurationUpdated,
    } {
        eventBus.Subscribe(event, projector)
    }
}
```

Handler dispatch in projectors:
```go
import mmEvents "easi/backend/internal/metamodel/publishedlanguage"

func (p *StrategyPillarCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
    handlers := map[string]func(context.Context, []byte) error{
        mmEvents.MetaModelConfigurationCreated: p.handleConfigurationCreated,
        mmEvents.StrategyPillarAdded:           p.handlePillarAdded,
        // ...
    }
    // ...
}
```

Local deserialization structs remain unchanged in each consumer.

### Pattern A Cleanup

The `component_cache_projector.go` and `realization_projector.go` in CapabilityMapping currently import `archEvents "easi/backend/internal/architecturemodeling/domain/events"`. Replace this direct domain coupling with:
1. Import constants from `architecturemodeling/publishedlanguage`
2. Define local deserialization structs in each projector (matching the existing pattern in pillar cache projectors)

## Files to Create

```
backend/internal/metamodel/publishedlanguage/events.go
backend/internal/capabilitymapping/publishedlanguage/events.go
backend/internal/architecturemodeling/publishedlanguage/events.go
```

## Files to Modify

```
backend/internal/capabilitymapping/infrastructure/api/routes.go
backend/internal/capabilitymapping/application/projectors/strategy_pillar_cache_projector.go
backend/internal/capabilitymapping/application/projectors/component_cache_projector.go
backend/internal/capabilitymapping/application/projectors/realization_projector.go
backend/internal/enterprisearchitecture/infrastructure/api/routes.go
backend/internal/enterprisearchitecture/application/projectors/strategy_pillar_cache_projector.go
backend/internal/enterprisearchitecture/application/projectors/domain_capability_metadata_projector.go
backend/internal/enterprisearchitecture/application/projectors/enterprise_capability_link_projector.go
```

## Checklist

- [x] Specification approved
- [x] MetaModel publishedlanguage package created
- [x] CapabilityMapping publishedlanguage package created
- [x] ArchitectureModeling publishedlanguage package created
- [x] ArchitectureViews publishedlanguage package created
- [x] CM routes.go migrated to use constants
- [x] EA routes.go migrated to use constants
- [x] AV routes.go migrated to use constants
- [x] VL routes.go migrated to use constants (fixed dead "ComponentDeleted" subscription)
- [x] CM pillar cache projector migrated to use constants
- [x] EA pillar cache projector migrated to use constants
- [x] CM component_cache_projector Pattern A removed (local structs + constants)
- [x] CM realization_projector Pattern A removed (local structs + constants)
- [x] EA domain_capability_metadata_projector migrated to use constants
- [x] EA enterprise_capability_link_projector migrated to use constants
- [x] All tests passing
- [x] Pattern documented in docs/backend/cross-context-events.md
