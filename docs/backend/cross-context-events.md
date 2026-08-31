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
| `eaPL` | `enterprisearchitecture/publishedlanguage` |
| `authPL` | `auth/publishedlanguage` |

## Delivery Semantics

`PostgresEventStore.SaveEvents` publishes to every subscriber on the same database transaction that inserts the event row, before that transaction commits (`backend/internal/infrastructure/eventstore/event_store.go`). A subscriber that returns an error — including one raised by a command it dispatches, which nests into the same transaction via `database.WithTx`/`TxFromContext` — rolls back the whole write: the event row, every cache write any subscriber made, and, when the producer itself joined a caller's transaction (as the merged create-tenant handler does via `TenantAwareDB.RunInTx`), the caller's own relational writes too. The caller gets the error back; nothing is left half-committed to retry against.

This means every subscriber (projector, reactor, deletion handler) must stay **database-only**. No outbound HTTP call, no email, no file I/O, no LLM/AI-provider call, no long-running computation — any such side effect would now execute while holding the producing write's transaction open, and would not automatically retry or compensate on a later rollback. A subscriber that legitimately needs a non-database side effect belongs outside this synchronous path (e.g. dispatched from a projected read model after commit), not inside an event handler.

## Complete Event Constants Catalogue

### Architecture Modeling (`archPL`)

```go
const (
    ApplicationComponentCreated       = "ApplicationComponentCreated"
    ApplicationComponentUpdated       = "ApplicationComponentUpdated"
    ApplicationComponentDeleted       = "ApplicationComponentDeleted"
    ApplicationComponentExpertAdded   = "ApplicationComponentExpertAdded"
    ApplicationComponentExpertRemoved = "ApplicationComponentExpertRemoved"

    ComponentRelationCreated = "ComponentRelationCreated"
    ComponentRelationUpdated = "ComponentRelationUpdated"
    ComponentRelationDeleted = "ComponentRelationDeleted"

    AcquiredEntityCreated = "AcquiredEntityCreated"
    AcquiredEntityUpdated = "AcquiredEntityUpdated"
    AcquiredEntityDeleted = "AcquiredEntityDeleted"

    VendorCreated = "VendorCreated"
    VendorUpdated = "VendorUpdated"
    VendorDeleted = "VendorDeleted"

    InternalTeamCreated = "InternalTeamCreated"
    InternalTeamUpdated = "InternalTeamUpdated"
    InternalTeamDeleted = "InternalTeamDeleted"

    OriginLinkSet          = "OriginLinkSet"
    OriginLinkReplaced     = "OriginLinkReplaced"
    OriginLinkNotesUpdated = "OriginLinkNotesUpdated"
    OriginLinkCleared      = "OriginLinkCleared"
    OriginLinkDeleted      = "OriginLinkDeleted"
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

Cross-context constants live in `events.go`; intra-context constants live in `internal_events.go` — both are listed below since spec 209 added cross-context consumers (OnePagers) for several events that were originally intra-context only.

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

    SystemLinkedToCapability = "SystemLinkedToCapability"
    SystemRealizationDeleted = "SystemRealizationDeleted"

    BusinessDomainCreated = "BusinessDomainCreated"
    BusinessDomainUpdated = "BusinessDomainUpdated"
    BusinessDomainDeleted = "BusinessDomainDeleted"

    EffectiveImportanceRecalculated = "EffectiveImportanceRecalculated"
    ApplicationFitScoreSet          = "ApplicationFitScoreSet"
    ApplicationFitScoreUpdated      = "ApplicationFitScoreUpdated"
    ApplicationFitScoreRemoved      = "ApplicationFitScoreRemoved"

    CapabilityExpertAdded             = "CapabilityExpertAdded"
    CapabilityExpertRemoved           = "CapabilityExpertRemoved"
    CapabilityTagAdded                = "CapabilityTagAdded"
    CapabilityDependencyCreated       = "CapabilityDependencyCreated"
    CapabilityDependencyDeleted       = "CapabilityDependencyDeleted"
    SystemRealizationUpdated          = "SystemRealizationUpdated"
    CapabilityRealizationsInherited   = "CapabilityRealizationsInherited"
    CapabilityRealizationsUninherited = "CapabilityRealizationsUninherited"
    StrategyImportanceSet             = "StrategyImportanceSet"
    StrategyImportanceUpdated         = "StrategyImportanceUpdated"
    StrategyImportanceRemoved         = "StrategyImportanceRemoved"
)
```

### Architecture Views (`avPL`)

```go
const (
    ViewCreated              = "ViewCreated"
    ViewRenamed              = "ViewRenamed"
    ViewDeleted              = "ViewDeleted"
    ViewVisibilityChanged    = "ViewVisibilityChanged"
    DefaultViewChanged       = "DefaultViewChanged"
    ComponentAddedToView     = "ComponentAddedToView"
    ComponentRemovedFromView = "ComponentRemovedFromView"
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

    TimeAssessmentRecorded = "TimeAssessmentRecorded"
    TimeAssessmentRemoved  = "TimeAssessmentRemoved"

    RealizationRoleAssigned = "RealizationRoleAssigned"
    RealizationRoleCleared  = "RealizationRoleCleared"

    JourneyPlanned                   = "JourneyPlanned"
    JourneyStarted                   = "JourneyStarted"
    JourneyCompleted                 = "JourneyCompleted"
    JourneyAbandoned                 = "JourneyAbandoned"
    JourneyProgressUpdated           = "JourneyProgressUpdated"
    JourneyDetailsUpdated            = "JourneyDetailsUpdated"
    JourneyMilestoneAdded            = "JourneyMilestoneAdded"
    JourneyMilestoneUpdated          = "JourneyMilestoneUpdated"
    JourneyMilestoneRemoved          = "JourneyMilestoneRemoved"
    JourneyMilestonesReordered       = "JourneyMilestonesReordered"
    JourneySourceApplicationsChanged = "JourneySourceApplicationsChanged"
)
```

### Enterprise Architecture (`eaPL`)

```go
const (
    EnterpriseCapabilityCreated           = "EnterpriseCapabilityCreated"
    EnterpriseCapabilityUpdated           = "EnterpriseCapabilityUpdated"
    EnterpriseCapabilityDeleted           = "EnterpriseCapabilityDeleted"
    EnterpriseCapabilityTargetMaturitySet = "EnterpriseCapabilityTargetMaturitySet"

    EnterpriseStrategicImportanceSet     = "EnterpriseStrategicImportanceSet"
    EnterpriseStrategicImportanceUpdated = "EnterpriseStrategicImportanceUpdated"
    EnterpriseStrategicImportanceRemoved = "EnterpriseStrategicImportanceRemoved"
)
```

### Auth (`authPL`)

```go
const (
    UserCreated     = "UserCreated"
    UserRoleChanged = "UserRoleChanged"
    UserDisabled    = "UserDisabled"
    UserEnabled     = "UserEnabled"

    InvitationCreated  = "InvitationCreated"
    InvitationAccepted = "InvitationAccepted"
    InvitationRevoked  = "InvitationRevoked"
    InvitationExpired  = "InvitationExpired"

    TenantCreated = "TenantCreated"
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
| `BusinessDomainCreated` / `Updated` / `Deleted` | `BusinessDomainNameCacheProjector` | same | Business-domain name cache read by the metadata projector at assignment time (spec 209) |

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

**Artifact name cache** (spec 209) — `ArtifactNameCacheProjector`, wired in `registerArtifactNameSubscriptions()`:

| Supplier | Events | Purpose |
|----------|--------|---------|
| Capability Mapping | `CapabilityCreated/Updated/Deleted`, `BusinessDomainCreated/Updated/Deleted` | Display names of capability and business-domain grant artifacts |
| Architecture Modeling | `ApplicationComponentCreated/Updated/Deleted`, `VendorCreated/Updated/Deleted`, `AcquiredEntityCreated/Updated/Deleted`, `InternalTeamCreated/Updated/Deleted` | Display names of component, vendor, acquired-entity and team grant artifacts |
| Architecture Views | `ViewCreated`, `ViewRenamed`, `ViewDeleted` | Display names of view grant artifacts |

### Arch Assistant consumes from:

**Auth** (`authPL`):

| Event | Handler | Wired In | Purpose |
|-------|---------|----------|---------|
| `TenantCreated` | `TenantCreatedHandler` | `archassistant/infrastructure/api/routes.go` `SetupArchAssistantRoutes()` | Provision default AI configuration for the new tenant |

Auth publishes `TenantCreated` itself (Platform was merged into Auth, spec 209 amendment 2026-08-30) and reacts to it in-context with `TenantCreatedReactor` to create the first-admin invitation; tenant, domain and OIDC reads are live queries against `auth.tenants`, `auth.tenant_domains` and `auth.tenant_oidc_configs`.

Access Delegation invites a grantee without an account by dispatching Auth's published command `EnsureInvitation` (see [Published Commands](#published-commands)); `EditGrantForNonUserCreated` is published when an invitation was created, for audit.

### Architecture Direction consumes from:

**Enterprise Architecture** (`eaPL`):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `EnterpriseCapabilityCreated` / `Updated` / `Deleted` / `TargetMaturitySet` | `EnterpriseCapabilityCacheProjector` | `architecturedirection/infrastructure/api/routes.go` `subscribeCacheEvents()` | Local enterprise capability cache for composition, existence checks and maturity analysis (spec 207) |
| `EnterpriseCapabilityDeleted` | `EnterpriseCapabilityDeletedReactor` | `SetupRoutes()` | Reject the active direction of a deleted enterprise capability |

**Capability Mapping** (`cmPL`) — capability node cache (spec 207):

| Event | Projector | Wired In | Purpose |
|-------|-----------|----------|---------|
| `CapabilityCreated` / `Updated` / `Deleted` / `ParentChanged` / `LevelChanged` / `AssignedToDomain` / `UnassignedFromDomain` / `MetadataUpdated`, `BusinessDomainUpdated` | `CapabilityNodeCacheProjector` | `subscribeCacheEvents()` | Local capability tree with L1 ancestor, effective business domain and maturity, used by composition, source eligibility, maturity analysis and capability existence / effective-domain checks |
| `SystemLinkedToCapability`, `SystemRealizationDeleted` | `RealizationCacheProjector` | `subscribeReferenceCacheEvents()` | Direct-realization cache behind TIME assessments and realization roles (spec 209) |
| `BusinessDomainDeleted` | `ReferenceCacheProjector` | same | Remove the domain from the reference-name cache so existence checks fail (spec 209) |

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
| `ApplicationComponentDeleted` | `StaleApplicationProjector`, `ReferenceCacheProjector` | same, `subscribeReferenceCacheEvents()` | Mark standard applications as stale; remove the component from the reference-name cache so existence checks fail (spec 209) |

### OnePagers consumes from:

All subscriptions are wired in `onepagers/infrastructure/api/routes.go` `SetupOnePagersRoutes()`; every cache is backfilled by migration 148 (spec 209).

| Supplier | Events | Projector | Cache | Purpose |
|----------|--------|-----------|-------|---------|
| Architecture Modeling (`archPL`) | `ApplicationComponentCreated/Updated/Deleted`, `ApplicationComponentExpertAdded/Removed`, `AcquiredEntityCreated/Updated/Deleted`, `VendorCreated/Updated/Deleted`, `InternalTeamCreated/Updated/Deleted` | `SubjectIndexProjector` | `one_pager_subject_index` (name, existence, completeness counters, `built_in_fields` = the complete published attribute set) | Subject header, built-in field values, subject existence, completeness |
| Capability Mapping (`cmPL`) | `CapabilityCreated/Updated/Deleted`, `CapabilityMetadataUpdated`, `CapabilityExpertAdded/Removed`, `CapabilityParentChanged`, `CapabilityLevelChanged` | `SubjectIndexProjector` | same | same, plus `parentId` / `level` built-in attributes |
| Enterprise Architecture (`eaPL`) | `EnterpriseCapabilityCreated/Updated/Deleted`, `EnterpriseCapabilityTargetMaturitySet` | `SubjectIndexProjector` | same | same, plus `targetMaturity` built-in attribute |
| Capability Mapping (`cmPL`) | `CapabilityCreated`, `SystemLinkedToCapability`, `SystemRealizationDeleted`, `CapabilityRealizationsInherited/Uninherited`, `CapabilityDependencyCreated/Deleted`, `CapabilityAssignedToDomain/UnassignedFromDomain`, `CapabilityParentChanged`, `BusinessDomainCreated/Updated/Deleted` | `SubjectRelationProjector` | `subject_relation_cache`, `business_domain_name_cache` | Relation built-in fields (realizations, dependencies, domains, parent/children) and domain labels |
| Architecture Modeling (`archPL`) | `ComponentRelationCreated/Deleted`, `OriginLinkSet/Replaced/Cleared/Deleted` | `SubjectRelationProjector` | `subject_relation_cache` | Relation built-in fields (component relations, built-by / purchased-from / acquired-via and their reverse entries) |
| MetaModel (`mmPL`) | `MetaModelConfigurationCreated`, `MaturityScaleConfigUpdated/Reset` | `MaturityScaleProjector` | `maturity_scale_cache` | Maturity-scale sections for rendering maturity fields |

Expert names arrive on the expert events themselves (`expertName`, `expertRole`, `contactInfo`), so no user cache is needed.

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

## Published Commands

The second and only other integration channel (spec 209): a supplier declares a command struct with `CommandName()` in its `publishedlanguage` package, aliases it internally (`type CreateInvitation = publishedlanguage.CreateInvitation`) and registers its handler under that name; a consumer imports the published struct and dispatches it through the shared command bus. A result carries at most the created ID. Command handlers never return supplier read-model data.

| Supplier | Command | Consumer | Purpose |
|----------|---------|----------|---------|
| Auth | `CreateInvitation` | Auth's own `TenantCreatedReactor` | First-admin invitation on tenant creation |
| Auth | `EnsureInvitation` | Access Delegation | Invite a grantee without an account; no-op when the e-mail has a user or a pending invitation; error when the domain is not allowed; `CreatedID` set only when an invitation was created |
| Architecture Modeling | `CreateApplicationComponent`, `CreateComponentRelation` | Importing | Import gateway |
| Capability Mapping | `CreateCapability`, `UpdateCapabilityMetadata`, `LinkSystemToCapability`, `AssignCapabilityToDomain` | Importing | Import gateway |
| Value Streams | `CreateValueStream`, `AddStage`, `AddStageCapability` | Importing | Import gateway |

## Integration Rules (spec 209)

A context depends on another context only through that context's published language — its events (above) and its published commands. Nothing else crosses a boundary:

- **No composition-root bridges.** `TestCompositionRootOnlyRegistersRoutes`: every production file under `backend/internal/infrastructure/api/` imports from a context only its `infrastructure/api` package; the router passes shared infrastructure and registers routes. Cross-context wiring, adapters and lookups do not exist.
- **No cross-schema SQL.** `TestSQLSchemaOwnership` fails on any runtime SQL referencing another context's schema — there is no allowlist. `TestNewMigrationsCrossSchemasOnlyInBackfills` lets a migration read another schema only when its filename contains `backfill`.
- **`shared/` and `infrastructure/` import no context** (`TestSharedAndInfrastructureImportNoContext`); a context imports only another context's `publishedlanguage` (`TestNoCrossBoundedContextImports`); published languages import only the standard library (`TestPublishedLanguageContractsPurity`).
- **The dependency graph is the import graph and it is acyclic** (`TestContextDependencyGraphIsAcyclic`).
- **Every cache of upstream data is backfilled** by a `*backfill*` migration from the supplier's tables and kept current by a projector on the supplier's published events, wired inside the consuming context.
- **Actor identity and role come from the request context**, never from Auth's read models.

When a context needs another context's data at request time, the fix is a local cache fed by the supplier's events (as Access Delegation, Architecture Direction, Enterprise Architecture, OnePagers and Auth do), or moving the derived read to the context that owns its inputs (as spec 207 did for composition), or serving the data from its owner (as spec 208 did for one-pager completeness).
