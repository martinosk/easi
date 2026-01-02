# Spec 107: CapabilityMapping Domain Model Refactoring

## Status: DONE

## Overview
Refactored the capabilitymapping bounded context to address anemic domain model patterns by moving business logic from handlers into aggregates and domain services.

## Problem Statement
The capabilitymapping context exhibited transaction script patterns within an event-sourced CQRS framework:

**Previous State:**
- Business rules validated in handlers instead of aggregates
- Handlers queried read models to enforce invariants
- Cross-aggregate validation leaked into handlers

**Specific Anti-Patterns Addressed:**
1. **DeleteCapabilityHandler** - validated "no children" rule in handler
2. **AssignCapabilityToDomainHandler** - validated "L1 only" rule in handler
3. **DeleteBusinessDomainHandler** - validated "no assignments" rule in handler

## Implementation

### Pattern Applied
Following proper DDD tactical patterns as advised by strategic DDD architect:

**Single-Aggregate Invariants → Aggregate Methods**
- Aggregates validate rules they can enforce using their own state

**Cross-Aggregate Invariants → Domain Services**
- Domain services query read models and enforce cross-aggregate rules
- Handlers call domain services before executing aggregate operations

**Handler Pattern:**
```
1. Parse command into value objects (validation happens here)
2. Domain service validates cross-aggregate constraints (if any)
3. Load aggregate from repository
4. Call aggregate method (single-aggregate business logic)
5. Save aggregate
6. Return result
```

### Changes Made

#### Domain Services Created
- `CapabilityDeletionService` - validates capability has no children before deletion
- `BusinessDomainDeletionService` - validates domain has no assignments before deletion

#### Aggregate Methods Added
- `Capability.CanBeAssignedToDomain()` - validates only L1 capabilities can be assigned (aggregate owns level property)

#### Aggregate Methods Removed (Anti-Pattern)
- Removed `Capability.CanDelete(hasChildren bool)` - passing validation state to aggregate was wrong pattern
- Removed `BusinessDomain.CanDelete(hasAssignments bool)` - same anti-pattern

#### BusinessDomainAssignment Simplified
- Removed `capabilityLevel` parameter from constructor - L1 validation moved to Capability aggregate

#### Handlers Refactored
- `DeleteCapabilityHandler` - now uses `CapabilityDeletionService` for cross-aggregate validation
- `DeleteBusinessDomainHandler` - now uses `BusinessDomainDeletionService` for cross-aggregate validation
- `AssignCapabilityToDomainHandler` - loads Capability aggregate and calls `CanBeAssignedToDomain()`

#### Infrastructure Adapters Created
- `CapabilityChildrenCheckerAdapter` - implements domain service interface using read model
- `BusinessDomainAssignmentCheckerAdapter` - implements domain service interface using read model

## Files Changed

### Created
- `backend/internal/capabilitymapping/domain/services/capability_deletion_service.go`
- `backend/internal/capabilitymapping/domain/services/business_domain_deletion_service.go`
- `backend/internal/capabilitymapping/infrastructure/adapters/capability_children_checker_adapter.go`
- `backend/internal/capabilitymapping/infrastructure/adapters/business_domain_assignment_checker_adapter.go`

### Modified
- `backend/internal/capabilitymapping/domain/aggregates/capability.go`
- `backend/internal/capabilitymapping/domain/aggregates/business_domain.go`
- `backend/internal/capabilitymapping/domain/aggregates/business_domain_assignment.go`
- `backend/internal/capabilitymapping/application/handlers/delete_capability_handler.go`
- `backend/internal/capabilitymapping/application/handlers/delete_business_domain_handler.go`
- `backend/internal/capabilitymapping/application/handlers/assign_capability_to_domain_handler.go`
- `backend/internal/capabilitymapping/infrastructure/api/routes.go`
- `backend/internal/capabilitymapping/infrastructure/api/error_registration.go`
- `backend/internal/capabilitymapping/infrastructure/api/capability_handlers.go`

### Tests Updated
- `backend/internal/capabilitymapping/domain/aggregates/capability_test.go`
- `backend/internal/capabilitymapping/domain/aggregates/business_domain_test.go`
- `backend/internal/capabilitymapping/domain/aggregates/business_domain_assignment_test.go`
- `backend/internal/capabilitymapping/application/handlers/delete_capability_handler_test.go`
- `backend/internal/capabilitymapping/application/handlers/delete_business_domain_handler_test.go`
- `backend/internal/capabilitymapping/application/handlers/assign_capability_to_domain_handler_test.go`

## Success Criteria Met
- [x] Business rules defined in exactly one place (aggregates or domain services)
- [x] Handlers contain no conditional business logic for cross-aggregate validation
- [x] Domain services handle cross-aggregate invariants
- [x] Aggregates validate rules using their own state only
- [x] All tests pass
