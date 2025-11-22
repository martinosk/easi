# Capability API Client and Types

## Description
Add TypeScript interfaces and API client methods to support capability management in the frontend. This establishes the foundation for all capability features by defining types matching backend DTOs and providing API methods for CRUD operations.

## User Need
As a user, I need the frontend to communicate with the capability backend API so that I can manage capabilities through the UI.

## API Endpoints to Support

Based on existing backend routes:

**Capabilities:**
- `POST /api/v1/capabilities` - Create capability
- `GET /api/v1/capabilities` - Get all capabilities
- `GET /api/v1/capabilities/:id` - Get capability by ID
- `GET /api/v1/capabilities/:id/children` - Get children
- `PUT /api/v1/capabilities/:id` - Update capability
- `PUT /api/v1/capabilities/:id/metadata` - Update metadata
- `POST /api/v1/capabilities/:id/experts` - Add expert
- `POST /api/v1/capabilities/:id/tags` - Add tag

**Dependencies:**
- `POST /api/v1/capability-dependencies` - Create dependency
- `GET /api/v1/capability-dependencies` - Get all dependencies
- `GET /api/v1/capabilities/:id/dependencies/outgoing` - Get outgoing
- `GET /api/v1/capabilities/:id/dependencies/incoming` - Get incoming
- `DELETE /api/v1/capability-dependencies/:id` - Delete dependency

**Realizations:**
- `POST /api/v1/capabilities/:id/systems` - Link system to capability
- `GET /api/v1/capabilities/:id/systems` - Get systems by capability
- `PUT /api/v1/capability-realizations/:id` - Update realization
- `DELETE /api/v1/capability-realizations/:id` - Delete realization
- `GET /api/v1/capability-realizations/by-component/:componentId` - Get capabilities by component

## Types Required

**Capability**: id, name, description, parentId, level (L1-L4), strategyPillar, pillarWeight, maturityLevel, ownershipModel, primaryOwner, eaOwner, status, experts, tags, createdAt, _links

**Expert**: name, role, contact, addedAt

**CapabilityDependency**: id, sourceCapabilityId, targetCapabilityId, dependencyType (Requires/Enables/Supports), description, createdAt, _links

**CapabilityRealization**: id, capabilityId, componentId, realizationLevel, notes, linkedAt, _links

**Request types** for create/update operations matching each endpoint

## Store Requirements

Extend Zustand store with:
- Capability state (list, dependencies, realizations)
- Actions: load, create, update, delete for each entity type
- Cache invalidation on mutations

## Acceptance Criteria
- [x] All TypeScript interfaces defined matching backend DTOs
- [x] All API client methods implemented with proper error handling
- [x] Zustand store extended with capability state and actions
- [x] API client follows existing patterns (axios, error interception)
- [x] Types properly exported for use in components

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [x] User sign-off
