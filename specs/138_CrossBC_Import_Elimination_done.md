# Cross-BC Go Import Elimination

## Description
Eliminate all remaining Go import violations between bounded contexts by introducing proper interfaces, adapters, and dependency injection. After this spec, every BC imports only from other BCs' `publishedlanguage` packages, shared kernel, or shared infrastructure.

## Purpose
After specs 135-137 resolve the published language gaps and SQL violations, several Go-level import violations remain. These are structural couplings where one BC directly imports another's internal packages (application commands, read models, infrastructure adapters). This spec resolves them all.

## Dependencies
- Spec 135: Published Language Expansion (for auth published language and StrategyPillarsGateway relocation)

## Remaining Violations After Spec 135

| # | Importing BC | Imported Package | Usage | Solution |
|---|---|---|---|---|
| 1 | accessdelegation | `architecturemodeling/application/readmodels` | ArtifactNameResolver reads component names | Interface + DI |
| 2 | accessdelegation | `architectureviews/application/readmodels` | ArtifactNameResolver reads view names | Interface + DI |
| 3 | accessdelegation | `capabilitymapping/application/readmodels` | ArtifactNameResolver reads capability/domain names | Interface + DI |
| 4 | accessdelegation | `auth/application/readmodels` | UserReadModel for user name lookup | Interface + DI |
| 5 | architectureviews | `auth/application/readmodels` | UserReadModel for view visibility check | Interface + DI |
| 6 | importing | `architecturemodeling/application/commands` | Import orchestrator dispatches component creation | Command bus + published language |
| 7 | importing | `capabilitymapping/application/commands` | Import orchestrator dispatches capability creation | Command bus + published language |

## Part 1: AccessDelegation — ArtifactNameResolver Decoupling

### Problem

`accessdelegation/infrastructure/services/artifact_name_resolver.go` imports read models from 3 other BCs to look up artifact names by ID:
- `architecturemodeling/application/readmodels.ApplicationComponentReadModel`
- `architectureviews/application/readmodels.ArchitectureViewReadModel`
- `capabilitymapping/application/readmodels.CapabilityReadModel` (and BusinessDomainReadModel)

It also imports `auth/application/readmodels.UserReadModel` for user name lookup.

### Solution: Define interfaces in AccessDelegation

AccessDelegation defines the interfaces it needs. The consuming BCs' read models happen to satisfy them, but the import direction is inverted: AccessDelegation depends on its own interfaces, and the wiring code (router initialization) injects the concrete implementations.

**File:** `backend/internal/accessdelegation/application/ports/artifact_name_ports.go`

```go
package ports

import "context"

// ArtifactNameLookup resolves an artifact ID to its human-readable name.
// Each bounded context provides an implementation for its own artifact types.
type ArtifactNameLookup interface {
    GetArtifactName(ctx context.Context, artifactID string) (string, error)
}

// UserNameLookup resolves a user ID to their display name.
type UserNameLookup interface {
    GetUserDisplayName(ctx context.Context, userID string) (string, error)
}
```

**File:** `backend/internal/accessdelegation/infrastructure/services/artifact_name_resolver.go` (updated)

```go
package services

import "easi/backend/internal/accessdelegation/application/ports"

type ArtifactNameResolver struct {
    componentLookup  ports.ArtifactNameLookup
    viewLookup       ports.ArtifactNameLookup
    capabilityLookup ports.ArtifactNameLookup
    domainLookup     ports.ArtifactNameLookup
    userLookup       ports.UserNameLookup
}
```

No imports from other BCs. The resolver depends only on its own interfaces.

### Adapter Implementations

Each BC creates a thin adapter in its own infrastructure layer that implements the AccessDelegation port:

**File:** `backend/internal/architecturemodeling/infrastructure/adapters/artifact_name_adapter.go`

```go
package adapters

import (
    "context"
    "easi/backend/internal/architecturemodeling/application/readmodels"
)

type ComponentNameAdapter struct {
    readModel *readmodels.ApplicationComponentReadModel
}

func NewComponentNameAdapter(rm *readmodels.ApplicationComponentReadModel) *ComponentNameAdapter {
    return &ComponentNameAdapter{readModel: rm}
}

func (a *ComponentNameAdapter) GetArtifactName(ctx context.Context, artifactID string) (string, error) {
    return a.readModel.GetComponentName(ctx, artifactID)
}
```

Similar adapters for:
- `architectureviews/infrastructure/adapters/view_name_adapter.go`
- `capabilitymapping/infrastructure/adapters/capability_name_adapter.go`
- `capabilitymapping/infrastructure/adapters/domain_name_adapter.go`
- `auth/infrastructure/adapters/user_name_adapter.go`

### Wiring

The central router (`backend/internal/infrastructure/api/router.go`) creates the adapters and injects them into AccessDelegation's setup:

```go
componentNameAdapter := archAdapters.NewComponentNameAdapter(archReadModels.Component)
viewNameAdapter := viewAdapters.NewViewNameAdapter(viewReadModels.View)
capabilityNameAdapter := cmAdapters.NewCapabilityNameAdapter(cmReadModels.Capability)
domainNameAdapter := cmAdapters.NewDomainNameAdapter(cmReadModels.BusinessDomain)
userNameAdapter := authAdapters.NewUserNameAdapter(authReadModels.User)

accessdelegationAPI.SetupAccessDelegationRoutes(accessdelegationAPI.AccessDelegationRoutesDeps{
    // ... existing deps ...
    ComponentNameLookup:  componentNameAdapter,
    ViewNameLookup:       viewNameAdapter,
    CapabilityNameLookup: capabilityNameAdapter,
    DomainNameLookup:     domainNameAdapter,
    UserNameLookup:       userNameAdapter,
})
```

### Import Direction After Fix

```
router.go (shared infrastructure) → imports adapters from each BC
    → passes adapters to accessdelegation as interface implementations
accessdelegation → imports ONLY its own ports package (interfaces)
adapters in each BC → import their own BC's read models
```

No BC imports another BC's internal packages.

## Part 2: ArchitectureViews — UserReadModel Decoupling

### Problem

`architectureviews/application/handlers/change_view_visibility_handler.go` imports `auth/application/readmodels.UserReadModel` to look up user existence for private view visibility.

### Solution: Define interface in ArchitectureViews

**File:** `backend/internal/architectureviews/application/ports/user_lookup_port.go`

```go
package ports

import "context"

type UserExistenceChecker interface {
    UserExists(ctx context.Context, userID string) (bool, error)
}
```

Update the handler to depend on this interface instead of auth's concrete read model.

Create an adapter in auth:

**File:** `backend/internal/auth/infrastructure/adapters/user_existence_adapter.go`

```go
package adapters

type UserExistenceAdapter struct {
    readModel *readmodels.UserReadModel
}

func (a *UserExistenceAdapter) UserExists(ctx context.Context, userID string) (bool, error) {
    user, err := a.readModel.GetByID(ctx, userID)
    return user != nil, err
}
```

Wire via the router, same pattern as Part 1.

## Part 3: Import Saga (Process Manager)

### Problem

`importing/application/orchestrator/import_orchestrator.go` is an ad-hoc orchestrator that directly imports command structs from 3 other BCs:
- `architecturemodeling/application/commands` (CreateApplicationComponent, CreateComponentRelation)
- `capabilitymapping/application/commands` (CreateCapability, UpdateCapabilityMetadata, LinkSystemToCapability, AssignCapabilityToDomain)
- `valuestreams/application/commands` (CreateValueStream, AddStage, AddStageCapability)

This IS cross-BC commanding, but it's not arbitrary — it's an import process that coordinates entity creation across bounded contexts. That's a **saga** (process manager). The problem isn't that cross-BC commanding happens; it's that the code doesn't follow the pattern that legitimizes it.

### Why a Saga Is Correct Here

In DDD, a **process manager (saga)** is the one pattern explicitly designed to coordinate work across bounded contexts by dispatching commands. From Vaughn Vernon's *Implementing Domain-Driven Design*: a process manager tracks the state of a long-running process, listens for events, and issues commands to participants.

The import process fits this pattern exactly:
- It is a long-running, multi-step process (8 phases, runs in a background goroutine)
- It coordinates entity creation across 3 bounded contexts
- It maintains state across phases (ID mappings from source → created entities)
- Later phases depend on results from earlier phases
- It tracks progress and reports results

What the current code does wrong: it's a saga that doesn't know it's a saga. No formal steps, no compensation, no participant abstraction — just a method that directly dispatches other BCs' commands.

### Solution: Textbook Orchestration-Based Saga

Restructure the import process as a proper saga with:
1. **Saga steps** — formal definition of each phase with forward action + compensating action
2. **Saga state** — explicit state object tracking ID mappings, created entities, and results
3. **Participant gateways** — interfaces the saga uses to communicate with other BCs (the saga never touches their commands directly)
4. **Compensation** — on critical failure, previously completed steps are undone in reverse order

### 3.1: Participant Gateways

The saga communicates with other BCs through **participant gateways** — interfaces defined in the importing BC. Each gateway exposes forward actions (create/link) and their corresponding compensating actions (delete/unlink).

**File:** `backend/internal/importing/application/ports/import_gateways.go`

```go
package ports

import "context"

type ComponentGateway interface {
    // Forward actions
    CreateComponent(ctx context.Context, name, description string) (string, error)
    CreateRelation(ctx context.Context, sourceID, targetID, relationType, name, description string) (string, error)
    // Compensating actions
    DeleteComponent(ctx context.Context, id string) error
    DeleteRelation(ctx context.Context, id string) error
}

type CapabilityGateway interface {
    // Forward actions
    CreateCapability(ctx context.Context, name, description, parentID, level string) (string, error)
    UpdateMetadata(ctx context.Context, id, eaOwner, status string) error
    LinkSystem(ctx context.Context, capabilityID, componentID, realizationLevel, notes string) (string, error)
    AssignToDomain(ctx context.Context, capabilityID, businessDomainID string) error
    // Compensating actions
    DeleteCapability(ctx context.Context, id string) error
    UnlinkSystem(ctx context.Context, realizationID string) error
}

type ValueStreamGateway interface {
    // Forward actions
    CreateValueStream(ctx context.Context, name, description string) (string, error)
    AddStage(ctx context.Context, valueStreamID, name, description string) (string, error)
    MapCapabilityToStage(ctx context.Context, valueStreamID, stageID, capabilityID string) error
    // Compensating actions
    DeleteValueStream(ctx context.Context, id string) error
}
```

Forward actions that create entities return `(createdID, error)` so the saga can track them for compensation. Compensating actions are the reverse operations.

### 3.2: Gateway Adapters

Each adapter lives in the **target BC's infrastructure layer**. It dispatches the BC's own commands internally — no cross-BC imports.

**File:** `backend/internal/architecturemodeling/infrastructure/adapters/import_component_gateway.go`

```go
package adapters

import (
    "context"
    "easi/backend/internal/architecturemodeling/application/commands"
    "easi/backend/internal/shared/cqrs"
)

type ImportComponentGateway struct {
    commandBus cqrs.CommandBus
}

func NewImportComponentGateway(bus cqrs.CommandBus) *ImportComponentGateway {
    return &ImportComponentGateway{commandBus: bus}
}

func (g *ImportComponentGateway) CreateComponent(ctx context.Context, name, description string) (string, error) {
    result, err := g.commandBus.Dispatch(ctx, &commands.CreateApplicationComponent{
        Name: name, Description: description,
    })
    if err != nil {
        return "", err
    }
    return result.CreatedID, nil
}

func (g *ImportComponentGateway) CreateRelation(ctx context.Context, sourceID, targetID, relationType, name, description string) (string, error) {
    result, err := g.commandBus.Dispatch(ctx, &commands.CreateComponentRelation{
        SourceComponentID: sourceID, TargetComponentID: targetID,
        RelationType: relationType, Name: name, Description: description,
    })
    if err != nil {
        return "", err
    }
    return result.CreatedID, nil
}

func (g *ImportComponentGateway) DeleteComponent(ctx context.Context, id string) error {
    _, err := g.commandBus.Dispatch(ctx, &commands.DeleteApplicationComponent{ID: id})
    return err
}

func (g *ImportComponentGateway) DeleteRelation(ctx context.Context, id string) error {
    _, err := g.commandBus.Dispatch(ctx, &commands.DeleteComponentRelation{ID: id})
    return err
}
```

Similar adapters for capability mapping and value streams. Each wraps its own BC's commands — forward and compensating.

### 3.3: Saga Step Definition

**File:** `backend/internal/importing/application/saga/saga_step.go`

```go
package saga

import (
    "context"
    "easi/backend/internal/importing/domain/aggregates"
    "easi/backend/internal/importing/domain/valueobjects"
)

type StepAction func(ctx context.Context, state *SagaState, data aggregates.ParsedData) []valueobjects.ImportError

type Step struct {
    Phase      string
    Forward    StepAction
    Compensate func(ctx context.Context, state *SagaState) error // nil = no compensation
}
```

### 3.4: Saga State

**File:** `backend/internal/importing/application/saga/saga_state.go`

```go
package saga

import "easi/backend/internal/importing/domain/aggregates"

type SagaState struct {
    // ID mappings: source file ID → created entity ID (used by later phases)
    SourceToComponentID   map[string]string
    SourceToCapabilityID  map[string]string
    SourceToValueStreamID map[string]string
    SourceToStageID       map[string]string

    // Compensation tracking: all IDs created by each phase (used for rollback)
    CreatedComponentIDs   []string
    CreatedCapabilityIDs  []string
    CreatedRelationIDs    []string
    CreatedRealizationIDs []string
    CreatedValueStreamIDs []string

    Result aggregates.ImportResult
}
```

The source-to-ID maps serve the forward direction (later phases look up IDs created by earlier phases). The created-ID slices serve compensation (delete everything created if we need to roll back).

### 3.5: Import Saga

**File:** `backend/internal/importing/application/saga/import_saga.go`

```go
package saga

import (
    "context"
    "easi/backend/internal/importing/application/ports"
    "easi/backend/internal/importing/domain/aggregates"
    "easi/backend/internal/importing/domain/valueobjects"
)

type ImportSaga struct {
    steps      []Step
    components ports.ComponentGateway
    capabilities ports.CapabilityGateway
    valueStreams ports.ValueStreamGateway
    repository  ImportSessionRepository
}

func New(
    comp ports.ComponentGateway,
    cap ports.CapabilityGateway,
    vs ports.ValueStreamGateway,
    repo ImportSessionRepository,
) *ImportSaga {
    s := &ImportSaga{
        components: comp, capabilities: cap,
        valueStreams: vs, repository: repo,
    }
    s.steps = s.defineSteps()
    return s
}

func (s *ImportSaga) defineSteps() []Step {
    return []Step{
        {Phase: valueobjects.PhaseCreatingComponents,
            Forward: s.createComponents, Compensate: s.deleteComponents},
        {Phase: valueobjects.PhaseCreatingCapabilities,
            Forward: s.createCapabilities, Compensate: s.deleteCapabilities},
        {Phase: valueobjects.PhaseAssigningCapabilityMetadata,
            Forward: s.updateMetadata, Compensate: nil},
        {Phase: valueobjects.PhaseCreatingValueStreams,
            Forward: s.createValueStreams, Compensate: s.deleteValueStreams},
        {Phase: valueobjects.PhaseCreatingRealizations,
            Forward: s.createRealizations, Compensate: s.deleteRealizations},
        {Phase: valueobjects.PhaseCreatingComponentRelations,
            Forward: s.createRelations, Compensate: s.deleteRelations},
        {Phase: valueobjects.PhaseAssigningDomains,
            Forward: s.assignDomains, Compensate: nil},
        {Phase: valueobjects.PhaseMappingCapabilitiesToStages,
            Forward: s.mapCapabilityStages, Compensate: nil},
    }
}

func (s *ImportSaga) Execute(ctx context.Context, session *aggregates.ImportSession) {
    state := newSagaState()
    parsedData := session.GetParsedData()

    for i, step := range s.steps {
        s.updateProgress(ctx, session, step.Phase, parsedData)
        errors := step.Forward(ctx, &state, parsedData)
        state.Result.AccumulateErrors(errors)
    }

    session.Complete(state.Result)
}

func (s *ImportSaga) compensate(ctx context.Context, state *SagaState, failedStepIndex int) {
    // Walk completed steps in reverse, executing each compensation
    for i := failedStepIndex - 1; i >= 0; i-- {
        if s.steps[i].Compensate != nil {
            _ = s.steps[i].Compensate(ctx, state)
        }
    }
}
```

The `Execute` method runs steps sequentially. Each step's forward action returns item-level errors that are accumulated (best-effort: individual item failures don't stop the import). The `compensate` method walks completed steps in reverse — it's available for future use when a phase-level abort/cancel is needed.

### 3.6: Step Implementations (Example)

Each step method lives on `ImportSaga` and uses the gateway interfaces:

```go
func (s *ImportSaga) createComponents(ctx context.Context, state *SagaState, data aggregates.ParsedData) []valueobjects.ImportError {
    var errors []valueobjects.ImportError
    for _, comp := range data.Components {
        createdID, err := s.components.CreateComponent(ctx, comp.Name, comp.Description)
        if err != nil {
            errors = append(errors, valueobjects.NewImportError(comp.SourceID, comp.Name, err.Error(), "skipped"))
            continue
        }
        state.SourceToComponentID[comp.SourceID] = createdID
        state.CreatedComponentIDs = append(state.CreatedComponentIDs, createdID)
    }
    return errors
}

func (s *ImportSaga) deleteComponents(ctx context.Context, state *SagaState) error {
    for _, id := range state.CreatedComponentIDs {
        _ = s.components.DeleteComponent(ctx, id)
    }
    return nil
}
```

Every create step tracks the IDs it creates. Every compensate step deletes them. The saga never imports from another BC.

### 3.7: Wiring

```go
// router.go
componentGateway := archAdapters.NewImportComponentGateway(commandBus)
capabilityGateway := cmAdapters.NewImportCapabilityGateway(commandBus)
valueStreamGateway := vsAdapters.NewImportValueStreamGateway(commandBus)

importSaga := saga.New(componentGateway, capabilityGateway, valueStreamGateway, importRepo)
```

### 3.8: Structural Result

```
importing/application/saga/        → defines steps, state, and execution
importing/application/ports/       → defines participant gateway interfaces
archmod/infrastructure/adapters/   → implements ComponentGateway (dispatches own commands)
capmap/infrastructure/adapters/    → implements CapabilityGateway (dispatches own commands)
valstr/infrastructure/adapters/    → implements ValueStreamGateway (dispatches own commands)
router.go                          → creates adapters, injects into saga
```

The importing BC imports only its own packages. Each adapter imports only its own BC's commands. No cross-BC Go imports. Published language stays events-only. The cross-BC commanding is architecturally justified: this is a saga, and dispatching commands to participants is what sagas do.

## Verification

After implementing all 3 parts, the only allowed cross-BC imports should be:
- `<bc>/publishedlanguage` (events and contracts — NO commands)
- `shared/*` (shared kernel)
- `infrastructure/*` (shared infrastructure)
- `platform/infrastructure/api` (shared middleware)

### Architecture Guard Test Update

Remove ALL remaining entries from the `allowedCrossBCImports` map in the architecture guard test (spec 134). The test should pass with an **empty allowlist**.

Run:
```bash
go test ./internal/ -run TestNoCrossBoundedContextImports
```

This must pass with zero allowlist entries.

## Files to Create

```
# AccessDelegation ports
backend/internal/accessdelegation/application/ports/artifact_name_ports.go

# ArchitectureViews ports
backend/internal/architectureviews/application/ports/user_lookup_port.go

# Importing saga
backend/internal/importing/application/ports/import_gateways.go
backend/internal/importing/application/saga/import_saga.go
backend/internal/importing/application/saga/saga_state.go
backend/internal/importing/application/saga/saga_step.go

# Adapters in each BC (read model adapters — Parts 1 & 2)
backend/internal/architecturemodeling/infrastructure/adapters/artifact_name_adapter.go
backend/internal/architectureviews/infrastructure/adapters/view_name_adapter.go
backend/internal/capabilitymapping/infrastructure/adapters/capability_name_adapter.go
backend/internal/capabilitymapping/infrastructure/adapters/domain_name_adapter.go
backend/internal/auth/infrastructure/adapters/user_name_adapter.go
backend/internal/auth/infrastructure/adapters/user_existence_adapter.go

# Gateway adapters (saga participant proxies — Part 3)
backend/internal/architecturemodeling/infrastructure/adapters/import_component_gateway.go
backend/internal/capabilitymapping/infrastructure/adapters/import_capability_gateway.go
backend/internal/valuestreams/infrastructure/adapters/import_value_stream_gateway.go
```

## Files to Modify

```
# AccessDelegation
backend/internal/accessdelegation/infrastructure/services/artifact_name_resolver.go  # Use interfaces
backend/internal/accessdelegation/infrastructure/api/routes.go  # Accept interface deps

# ArchitectureViews
backend/internal/architectureviews/application/handlers/change_view_visibility_handler.go  # Use interface

# Importing
backend/internal/importing/application/handlers/confirm_import_handler.go  # Use saga instead of orchestrator
backend/internal/importing/infrastructure/api/routes.go  # Wire saga with gateways

# Central wiring
backend/internal/infrastructure/api/router.go  # Create adapters, create saga, inject
```

## Files to Delete

```
backend/internal/importing/application/orchestrator/import_orchestrator.go  # Replaced by saga
```

## Success Criteria

- Architecture guard test `allowedCrossBCImports` map is EMPTY
- `go test ./internal/ -run TestNoCrossBoundedContextImports` passes with zero allowlist entries
- No BC imports another BC's `application/readmodels`, `application/commands`, `domain/*`, or `infrastructure/*` packages
- Published language contains events and contracts only — no command structs
- Import orchestrator replaced by saga with formal steps + compensation
- All existing import functionality works identically
- `go test ./...` passes

## Checklist

- [x] Specification approved
- [x] Part 1: AccessDelegation ports defined
- [x] Part 1: Adapters created in architecturemodeling, architectureviews, capabilitymapping, auth
- [x] Part 1: ArtifactNameResolver refactored to use interfaces
- [x] Part 1: Wiring updated in router.go
- [x] Part 2: ArchitectureViews UserExistenceChecker port defined
- [x] Part 2: Auth UserExistenceAdapter created
- [x] Part 2: change_view_visibility_handler refactored
- [x] Part 3: Participant gateway interfaces defined (forward + compensating actions)
- [x] Part 3: Gateway adapters created in architecturemodeling, capabilitymapping, valuestreams
- [x] Part 3: Saga step, state, and executor implemented
- [x] Part 3: Old orchestrator replaced by saga
- [x] Part 3: confirm_import_handler updated to use saga
- [x] Part 3: Importing BC has zero imports from other BCs' internal packages
- [x] Architecture guard test allowlist completely emptied
- [x] All tests passing
- [x] User sign-off
