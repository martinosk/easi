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
backend/internal/architecturedirection/publishedlanguage/events.go
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
| `adirPL` | `architecturedirection/publishedlanguage` |

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

Cross-context constants live in `events.go`; intra-context constants live in `internal_events.go`.

```go
const (
    CapabilityCreated              = "CapabilityCreated"
    CapabilityUpdated              = "CapabilityUpdated"
    CapabilityDeleted              = "CapabilityDeleted"
    CapabilityMetadataUpdated      = "CapabilityMetadataUpdated"
    CapabilityParentChanged        = "CapabilityParentChanged"
    CapabilityLevelChanged         = "CapabilityLevelChanged"
    CapabilityAssignedToDomain     = "CapabilityAssignedToDomain"
    CapabilityUnassignedFromDomain = "CapabilityUnassignedFromDomain"
    SystemLinkedToCapability       = "SystemLinkedToCapability"
    SystemRealizationDeleted       = "SystemRealizationDeleted"
    BusinessDomainCreated          = "BusinessDomainCreated"
    BusinessDomainUpdated          = "BusinessDomainUpdated"
    BusinessDomainDeleted          = "BusinessDomainDeleted"
    EffectiveImportanceRecalculated = "EffectiveImportanceRecalculated"
    ApplicationFitScoreSet          = "ApplicationFitScoreSet"
    ApplicationFitScoreRemoved      = "ApplicationFitScoreRemoved"
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

### Architecture Direction (`adirPL`)

```go
const (
    DirectionDrafted                   = "DirectionDrafted"
    DirectionProposed                  = "DirectionProposed"
    DirectionAgreed                    = "DirectionAgreed"
    DirectionRejected                  = "DirectionRejected"
    DirectionNarrativeUpdated          = "DirectionNarrativeUpdated"
    DirectionHorizonChanged            = "DirectionHorizonChanged"
    DirectionPlacementsChanged         = "DirectionPlacementsChanged"
    DirectionSourceCapabilitiesChanged = "DirectionSourceCapabilitiesChanged"

    StandardApplicationSet = "StandardApplicationSet"
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

### Value Streams consumes from:

**Capability Mapping** (`cmPL`):

| Event | Projector/Handler | Wired In | Purpose |
|-------|-------------------|----------|---------|
| `CapabilityCreated` | `CapabilityProjector` | `valuestreams/infrastructure/api/routes.go` `SetupValueStreamsRoutes()` | Cache capability ID-to-name mapping |
| `CapabilityUpdated` | `CapabilityProjector`, `CapabilityNameSyncProjector` | same | Update cached name; update denormalized capability name in stage mappings |
| `CapabilityDeleted` | `CapabilityProjector`, `CapabilityDeletedHandler` | same | Remove cache entry; remove all stage-capability mappings referencing the deleted capability |

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

### Arch Assistant consumes from:

**Platform** (`platformPL`):

| Event | Handler | Wired In | Purpose |
|-------|---------|----------|---------|
| `TenantCreated` | `TenantCreatedHandler` | `archassistant/infrastructure/api/routes.go` `SetupArchAssistantRoutes()` | Provision default AI configuration for the new tenant |

Auth consumes no cross-context events. Access Delegation asks Auth to invite a grantee without an account through its `InvitationRequester` port (a declared bridge, spec 206); `EditGrantForNonUserCreated` is still published for audit.

### Architecture Direction consumes from:

**Enterprise Architecture** (`eaPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `EnterpriseCapabilityCreated` / `Updated` / `Deleted` / `TargetMaturitySet` | `EnterpriseCapabilityCacheProjector` | `architecturedirection/infrastructure/api/routes.go` `subscribeCacheEvents()` | Local enterprise capability cache for composition, existence checks and maturity analysis (spec 207) |
| `EnterpriseCapabilityDeleted` | `EnterpriseCapabilityDeletedReactor` | `SetupRoutes()` | Reject the active direction of a deleted enterprise capability |

**Capability Mapping** (`cmPL`) — capability node cache (spec 207):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `CapabilityCreated` / `Updated` / `Deleted` / `ParentChanged` / `LevelChanged` / `AssignedToDomain` / `UnassignedFromDomain` / `MetadataUpdated`, `BusinessDomainUpdated` | `CapabilityNodeCacheProjector` | `subscribeCacheEvents()` | Local capability tree with L1 ancestor, effective business domain and maturity, used by composition, source eligibility and maturity analysis |

**Capability Mapping** (`cmPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `CapabilityCreated` | `StaleReferenceProjector` | `architecturedirection/infrastructure/api/routes.go` `subscribeEvents()` | Cache capability name for stale detection |
| `CapabilityUpdated` | `StaleReferenceProjector` | same | Update cached capability name |
| `CapabilityDeleted` | `StaleReferenceProjector` | same | Mark source capabilities as stale |
| `BusinessDomainCreated` | `StaleReferenceProjector` | same | Cache business domain name |
| `BusinessDomainUpdated` | `StaleReferenceProjector` | same | Update cached business domain name |
| `CapabilityAssignedToDomain` | `StaleReferenceProjector` | same | Update source capability's business domain |
| `CapabilityUnassignedFromDomain` | `StaleReferenceProjector` | same | Clear source capability's business domain |

**Architecture Modeling** (`amPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `ApplicationComponentCreated` | `StaleApplicationProjector` | `architecturedirection/infrastructure/api/routes.go` `subscribeStandardApplicationEvents()` | Cache application component name |
| `ApplicationComponentUpdated` | `StaleApplicationProjector` | same | Update cached application component name |
| `ApplicationComponentDeleted` | `StaleApplicationProjector` | same | Mark standard applications as stale when component is deleted |

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
- [ ] Verify all downstream read models that reference the artifact are cleaned up

## Composition-Root Bridges (Query-Time Integration)

Some cross-context dependencies are synchronous: the composition root (`backend/internal/infrastructure/api/`) constructs an adapter over one context's read model or service and passes it into another context's port. Each such file is a **bridge** and must be declared in `declaredBridges` in `backend/internal/architecture_bridges_test.go` with every consumer → supplier edge and a reason. The tests enforce:

- `router.go` imports only a context's `infrastructure/api` package — all wiring lives in `*_bridges.go` / adapter files.
- A composition-root file reaching two or more contexts must be declared, and the declaration must name exactly the contexts the file imports. A file reaching one context belongs inside that context.
- The graph of published-language imports plus declared bridges is acyclic.
- `shared/` and `infrastructure/` packages import only a context's `publishedlanguage`; production code never imports `internal/testing`.

Declared bridges today (consumer → supplier):

| File | Consumer | Suppliers | Purpose |
|------|----------|-----------|---------|
| `accessdelegation_bridges.go` | Access Delegation | Capability Mapping, Architecture Modeling, Architecture Views, Auth | Artifact display names; user existence, pending invitations, allowed domains and invitation requests for grantees without an account |
| `architecturedirection_bridges.go` | Architecture Direction | Capability Mapping, Architecture Modeling | Capability / domain / component existence, direct realization lookup, effective-domain check |
| `architectureviews_bridges.go` | Architecture Views | Auth | User role check for view visibility |
| `enterprisearchitecture_bridges.go` | Enterprise Architecture | Capability Mapping | Business-domain name at capability assignment time |
| `importing_bridges.go` | Importing | Architecture Modeling, Capability Mapping, Value Streams | Import gateways |
| `onepager_builtin_field_adapters.go` | OnePagers | Architecture Modeling, Capability Mapping, Enterprise Architecture, MetaModel | Built-in field values and maturity scale sections |
| `onepager_relation_adapters.go` | OnePagers | Architecture Modeling, Capability Mapping, Enterprise Architecture, Architecture Direction | Relation fields, including enterprise capability composition |
| `onepager_subject_adapters.go` | OnePagers | Architecture Modeling, Capability Mapping, Enterprise Architecture | Subject existence |

Prefer the local cache projector pattern for data that must stay consistent over time; a bridge is acceptable for write-time validation and display enrichment. A bridge may never close a cycle — the fix is to move the derived read to the context that owns its inputs (as spec 207 did for composition) or to serve the data from its owner (as spec 208 did for one-pager completeness).
