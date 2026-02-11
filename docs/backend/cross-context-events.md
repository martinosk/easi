# Cross-Context Event Integration

## Published Language

Each publishing bounded context exposes a `publishedlanguage/events.go` package with typed string constants for its event types:

```
backend/internal/architecturemodeling/publishedlanguage/events.go
backend/internal/architectureviews/publishedlanguage/events.go
backend/internal/capabilitymapping/publishedlanguage/events.go
backend/internal/metamodel/publishedlanguage/events.go
backend/internal/accessdelegation/publishedlanguage/events.go
backend/internal/valuestreams/publishedlanguage/events.go
```

These packages contain **only constants**. No structs, no constructors, no logic.

```go
package publishedlanguage

const (
    ApplicationComponentCreated = "ApplicationComponentCreated"
    ApplicationComponentUpdated = "ApplicationComponentUpdated"
    ApplicationComponentDeleted = "ApplicationComponentDeleted"
)
```

### When to add a constant

When a bounded context publishes an event that another context subscribes to. Intra-context subscriptions do not need published language constants (they use local event type strings).

### When to create a new package

When a bounded context becomes a publisher for the first time (i.e., another context needs to subscribe to its events).

## Anti-Corruption Layer (ACL)

Consuming contexts **never import domain event structs** from the publishing context. Instead:

1. Import the **published language constants** for subscription and handler dispatch
2. Define **local deserialization structs** with only the fields the consumer needs

```go
import archPL "easi/backend/internal/architecturemodeling/publishedlanguage"

type componentDeletedEvent struct {
    ID string `json:"id"`
}

func (p *Projector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
    switch eventType {
    case archPL.ApplicationComponentDeleted:
        var event componentDeletedEvent
        json.Unmarshal(eventData, &event)
        // handle locally
    }
}
```

### Import alias convention

| Alias | Package |
|-------|---------|
| `archPL` | `architecturemodeling/publishedlanguage` |
| `avPL` | `architectureviews/publishedlanguage` |
| `cmPL` | `capabilitymapping/publishedlanguage` |
| `mmPL` | `metamodel/publishedlanguage` |
| `adPL` | `accessdelegation/publishedlanguage` |
| `vsPL` | `valuestreams/publishedlanguage` |

## Complete Event Constants Catalogue

### Architecture Modeling (`archPL`)

```go
const (
    ApplicationComponentCreated = "ApplicationComponentCreated"
    ApplicationComponentUpdated = "ApplicationComponentUpdated"
    ApplicationComponentDeleted = "ApplicationComponentDeleted"
    ComponentRelationDeleted    = "ComponentRelationDeleted"
    AcquiredEntityDeleted       = "AcquiredEntityDeleted"
    VendorDeleted               = "VendorDeleted"
    InternalTeamDeleted         = "InternalTeamDeleted"
)
```

### MetaModel (`mmPL`)

```go
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

### Capability Mapping (`cmPL`)

```go
const (
    CapabilityCreated              = "CapabilityCreated"
    CapabilityUpdated              = "CapabilityUpdated"
    CapabilityDeleted              = "CapabilityDeleted"
    CapabilityParentChanged        = "CapabilityParentChanged"
    CapabilityAssignedToDomain     = "CapabilityAssignedToDomain"
    CapabilityUnassignedFromDomain = "CapabilityUnassignedFromDomain"
    BusinessDomainDeleted          = "BusinessDomainDeleted"
)
```

### Architecture Views (`avPL`)

```go
const (
    ViewDeleted = "ViewDeleted"
)
```

### Access Delegation (`adPL`)

```go
const (
    EditGrantActivated         = "EditGrantActivated"
    EditGrantRevoked           = "EditGrantRevoked"
    EditGrantExpired           = "EditGrantExpired"
    EditGrantForNonUserCreated = "EditGrantForNonUserCreated"
)
```

### Value Streams (`vsPL`)

```go
const (
    ValueStreamCreated                = "ValueStreamCreated"
    ValueStreamUpdated                = "ValueStreamUpdated"
    ValueStreamDeleted                = "ValueStreamDeleted"
    ValueStreamStageAdded             = "ValueStreamStageAdded"
    ValueStreamStageUpdated           = "ValueStreamStageUpdated"
    ValueStreamStageRemoved           = "ValueStreamStageRemoved"
    ValueStreamStagesReordered        = "ValueStreamStagesReordered"
    ValueStreamStageCapabilityAdded   = "ValueStreamStageCapabilityAdded"
    ValueStreamStageCapabilityRemoved = "ValueStreamStageCapabilityRemoved"
)
```

## Cross-Context Subscription Registry

Every event subscription that crosses a bounded context boundary is documented below, organized by consuming context. Use this registry when adding new events to ensure all consumers are accounted for.

### Architecture Views consumes from:

**Architecture Modeling** (`archPL`):

| Event | Handler | Wired In | Purpose |
|-------|---------|----------|---------|
| `ApplicationComponentDeleted` | `ApplicationComponentDeletedHandler` | `architectureviews/infrastructure/api/routes.go` `SubscribeEvents()` | Remove deleted component from all views via command dispatch |
| `ComponentRelationDeleted` | `ComponentRelationDeletedHandler` | same | Clean up relation visualization data |

### Capability Mapping consumes from:

**Architecture Modeling** (`archPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `ApplicationComponentCreated` | `ComponentCacheProjector` | `capabilitymapping/infrastructure/api/routes.go` `subscribeComponentCacheEvents()` | Cache component ID-to-name mapping |
| `ApplicationComponentUpdated` | `ComponentCacheProjector`, `RealizationProjector` | same + `subscribeRealizationEvents()` | Update cached name; update component name in realization read model |
| `ApplicationComponentDeleted` | `ComponentCacheProjector`, `RealizationProjector` | same | Remove cache entry; remove all realizations for that component |

**MetaModel** (`mmPL`):

| Event | Projector/Handler | Wired In | Purpose |
|-------|-------------------|----------|---------|
| `MetaModelConfigurationCreated` | `StrategyPillarCacheProjector` | `subscribeMetaModelEvents()` | Seed all pillars into local cache on initial tenant setup |
| `StrategyPillarAdded` | `StrategyPillarCacheProjector` | same | Insert new pillar into local cache |
| `StrategyPillarUpdated` | `StrategyPillarCacheProjector` | same | Update pillar name and description |
| `StrategyPillarRemoved` | `StrategyPillarCacheProjector` | same | Remove pillar from cache |
| `PillarFitConfigurationUpdated` | `StrategyPillarCacheProjector` | same | Update fit scoring configuration |
| `MaturityScaleConfigUpdated` | `MaturityScaleConfigUpdatedHandler` | `setupMetaModelEventHandlers()` | Invalidate maturity scale gateway cache |
| `MaturityScaleConfigReset` | `MaturityScaleConfigUpdatedHandler` | same | Reset maturity scale gateway to defaults |

### Enterprise Architecture consumes from:

**Capability Mapping** (`cmPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `CapabilityCreated` | `DomainCapabilityMetadataProjector` | `enterprisearchitecture/infrastructure/api/routes.go` `subscribeCapabilityMappingEvents()` | Track new capability with level, parent, L1 root |
| `CapabilityUpdated` | `DomainCapabilityMetadataProjector` | same | Update capability name in metadata |
| `CapabilityDeleted` | `DomainCapabilityMetadataProjector` | same | Remove metadata, clean up links and blocking records |
| `CapabilityParentChanged` | `DomainCapabilityMetadataProjector`, `EnterpriseCapabilityLinkProjector` | same + `subscribeLinkEvents()` | Recalculate L1 ancestry; recompute blocking relationships for subtree |
| `CapabilityAssignedToDomain` | `DomainCapabilityMetadataProjector` | `subscribeCapabilityMappingEvents()` | Update business domain for L1 subtree, recalculate enterprise domain counts |
| `CapabilityUnassignedFromDomain` | `DomainCapabilityMetadataProjector` | same | Clear business domain for L1 subtree, recalculate counts |

**MetaModel** (`mmPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `MetaModelConfigurationCreated` | `StrategyPillarCacheProjector` | `subscribePillarCacheEvents()` | Seed pillar cache on initial configuration |
| `StrategyPillarAdded` | `StrategyPillarCacheProjector` | same | Add pillar to local cache |
| `StrategyPillarUpdated` | `StrategyPillarCacheProjector` | same | Update pillar name and description |
| `StrategyPillarRemoved` | `StrategyPillarCacheProjector` | same | Remove pillar from cache |
| `PillarFitConfigurationUpdated` | `StrategyPillarCacheProjector` | same | Update fit scoring config in cache |

### View Layouts consumes from:

**Architecture Modeling** (`archPL`):

| Event | Handler | Wired In | Purpose |
|-------|---------|----------|---------|
| `ApplicationComponentDeleted` | `ComponentDeletedHandler` | `viewlayouts/infrastructure/api/routes.go` `SubscribeEvents()` | Remove element positions for deleted component |

**Capability Mapping** (`cmPL`):

| Event | Handler | Wired In | Purpose |
|-------|---------|----------|---------|
| `CapabilityDeleted` | `CapabilityDeletedHandler` | same | Remove element positions for deleted capability |
| `BusinessDomainDeleted` | `BusinessDomainDeletedHandler` | same | Remove layout container for deleted domain |

**Architecture Views** (`avPL`):

| Event | Handler | Wired In | Purpose |
|-------|---------|----------|---------|
| `ViewDeleted` | `ViewDeletedHandler` | same | Remove layout container for deleted view |

### Access Delegation consumes from:

**Architecture Modeling** (`archPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `ApplicationComponentDeleted` | `ArtifactDeletionProjector` (component) | `accessdelegation/infrastructure/api/routes.go` `registerArtifactDeletionSubscriptions()` | Revoke all edit grants for deleted component |
| `AcquiredEntityDeleted` | `ArtifactDeletionProjector` (acquired_entity) | same | Revoke grants for deleted acquired entity |
| `VendorDeleted` | `ArtifactDeletionProjector` (vendor) | same | Revoke grants for deleted vendor |
| `InternalTeamDeleted` | `ArtifactDeletionProjector` (internal_team) | same | Revoke grants for deleted internal team |

**Capability Mapping** (`cmPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `CapabilityDeleted` | `ArtifactDeletionProjector` (capability) | same | Revoke all edit grants for deleted capability |
| `BusinessDomainDeleted` | `ArtifactDeletionProjector` (domain) | same | Revoke all edit grants for deleted domain |

**Architecture Views** (`avPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `ViewDeleted` | `ArtifactDeletionProjector` (view) | same | Revoke all edit grants for deleted view |

### Auth consumes from:

**Access Delegation** (`adPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `EditGrantForNonUserCreated` | `InvitationAutoCreateProjector` | `infrastructure/api/router.go` `wireAutoInvitationProjector()` | Auto-create platform invitation for non-user grantee |

## Adding a New Cross-Context Event

Follow this checklist when a bounded context needs to publish an event for another context:

1. **Add the constant** to the publisher's `publishedlanguage/events.go`
2. **Create or update the projector/handler** in the consuming context using local deserialization structs
3. **Wire the subscription** in the consumer's route setup function (`SubscribeEvents` or `setupEventSubscriptions`)
4. **Update this document** -- add entries to both the catalogue and the subscription registry
5. **Update the architecture README** -- add the event to the Published Language Catalogue table and update the context map if a new integration path is introduced

### Checklist for deletion events

When adding a new deletable artifact type:

- [ ] Add `<ArtifactType>Deleted` constant to the publisher's published language
- [ ] Subscribe `ArtifactDeletionProjector` in Access Delegation for grant cleanup
- [ ] Subscribe cleanup handler in View Layouts if the artifact can appear in layouts
- [ ] Verify all downstream read models that reference the artifact are cleaned up

## Query-Based Integration (Non-Event)

Some cross-context dependencies use synchronous queries rather than events:

| Consumer | Provider | Mechanism | Purpose |
|----------|----------|-----------|---------|
| Value Streams | Capability Mapping | `CapabilityGateway` (reads `CapabilityReadModel`) | Verify capability existence before adding to stage |
| Capability Mapping | Architecture Modeling | `ComponentGateway` (reads `ApplicationComponentReadModel`) | Look up component name during realization linking |
| Access Delegation | Multiple contexts | `ArtifactNameResolver` (reads from multiple read models) | Resolve display names for edit grant artifacts |

These query-based integrations are acceptable for validation-at-write-time and display enrichment. For data that must remain consistent over time (e.g., component names in realization read models), prefer the local cache projector pattern instead.
