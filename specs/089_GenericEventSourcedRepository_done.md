# Spec 089: Generic Event-Sourced Repository

## Overview
Create a generic base repository for event-sourced aggregates to eliminate ~1,142 lines of duplicate code across 9 repositories.

## Problem Statement
All event-sourced repositories contain nearly identical implementations:
- `Save()` method: same pattern for getting uncommitted events, calling event store, marking committed
- `GetByID()` method: same pattern for fetching events, checking empty, deserializing, loading from history
- Constructor pattern: same dependency injection of event store
- Error variable definitions: same sentinel error pattern

**Repositories affected**:
- `ApplicationComponentRepository`, `ComponentRelationRepository` (architecturemodeling)
- `ArchitectureViewRepository` (architectureviews)
- `CapabilityRepository`, `BusinessDomainRepository`, `BusinessDomainAssignmentRepository`, `DependencyRepository`, `RealizationRepository` (capabilitymapping)

Note: `LayoutContainerRepository` (viewlayouts) is NOT event-sourced - it uses traditional SQL persistence.

**Additional issue**: Two different event deserialization patterns exist (switch statements vs map-based), causing inconsistency.

## Requirements

### R1: Aggregate Interface
Define a formal interface that all event-sourced aggregates must implement:
- `ID() string` - returns aggregate identifier
- `Version() int` - returns current version
- `GetUncommittedChanges() []domain.DomainEvent` - returns pending events
- `MarkChangesAsCommitted()` - clears pending events

### R2: Generic Repository Base
Create a generic repository base using Go generics:
- Must handle `Save()` with optimistic concurrency (version checking)
- Must handle `GetByID()` with event reconstruction
- Must support custom `LoadFromHistory` function per aggregate type
- Must support custom event deserializer per aggregate type

### R3: Event Deserializer Interface
Standardize event deserialization:
- Define common interface for event deserializers
- Support map-based deserializer registration (the cleaner pattern from `architecture_view_repository.go`)
- Provide helper for common JSON unmarshaling with error logging

### R4: Error Handling
Create shared sentinel errors:
- `ErrAggregateNotFound` as base error
- Context-specific errors should wrap this base error
- Enable `errors.Is(err, ErrAggregateNotFound)` checking

### R5: Migration Path
- Existing repositories become thin wrappers around generic base
- Aggregate-specific logic (deserializers, error types) remains in context
- No changes to repository interfaces used by handlers

## Affected Files

### Files to Create
- `backend/internal/shared/domain/aggregate_interface.go`
- `backend/internal/shared/infrastructure/repository/event_sourced_repository.go`
- `backend/internal/shared/infrastructure/repository/event_deserializer.go`
- `backend/internal/shared/infrastructure/repository/errors.go`

### Files to Modify
- `backend/internal/architecturemodeling/infrastructure/repositories/application_component_repository.go`
- `backend/internal/architecturemodeling/infrastructure/repositories/component_relation_repository.go`
- `backend/internal/architectureviews/infrastructure/repositories/architecture_view_repository.go`
- `backend/internal/capabilitymapping/infrastructure/repositories/*.go` (5 files)

## Checklist
- [x] Specification approved
- [x] Aggregate interface defined
- [x] Generic repository base created
- [x] Event deserializer interface created
- [x] Shared error types created
- [x] architecturemodeling repositories migrated
- [x] architectureviews repositories migrated
- [x] capabilitymapping repositories migrated
- [x] All existing tests pass
- [ ] User sign-off
