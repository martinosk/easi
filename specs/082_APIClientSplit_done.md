# API Client Split by Bounded Context

## Description
Split the monolithic API client into domain-specific modules aligned with bounded contexts. This improves maintainability and enables feature-level code ownership.

## Current State
- Single `client.ts` file contains all API methods (~560 lines)
- Mixes concerns from components, capabilities, relations, views, business domains, auth, and metadata
- Difficult to locate specific API methods
- Changes to one domain risk affecting others

## Target State
- Core API client handles only HTTP configuration and error handling
- Each bounded context has its own API module
- Modules are co-located with their feature directories
- Clear ownership and reduced merge conflicts

## Requirements

### Phase 1: Core Client Extraction
- [x] Extract Axios instance configuration to dedicated client module
- [x] Extract shared error handling and response processing
- [x] Extract shared type utilities (branded types, response wrappers)
- [x] Maintain backward compatibility during migration

### Phase 2: Feature API Modules
- [x] Create components API module in features/components/api/
- [x] Create capabilities API module in features/capabilities/api/
- [x] Create relations API module in features/relations/api/
- [x] Create views API module in features/views/api/
- [x] Create business-domains API module in features/business-domains/api/
- [x] Create canvas API module in features/canvas/api/ (layout endpoints)

### Phase 3: Shared API Modules
- [x] Create metadata API module for reference data (maturity levels, statuses, etc.)
- [x] Keep auth API in features/auth/api/ (already exists)

### Phase 4: Migration
- [x] Update all imports to use feature-specific API modules
- [x] Update store slices to import from new locations
- [x] Remove methods from original client.ts (now delegates to feature APIs)
- [x] Original client.ts now acts as backward-compatible facade

## Incremental Delivery
1. First: Extract core client infrastructure (no breaking changes)
2. Second: Create feature API modules (export from both old and new)
3. Third: Migrate imports one feature at a time
4. Fourth: Remove deprecated exports

## Checklist
- [x] Specification ready
- [x] Core client extracted
- [x] Feature API modules created
- [x] All imports migrated
- [x] Original client refactored to facade
- [x] Tests passing
- [x] User sign-off
