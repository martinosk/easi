# Capability Mapping - Developer Guide

## Overview

The Capability Mapping bounded context enables modeling of business capabilities with hierarchy, dependencies, and metadata. It uses CQRS with event sourcing.

## Core Concepts

### Capability Hierarchy (4 Levels)
- **L1 (Domain)** - Top-level business domains (e.g., "Customer Engagement")
- **L2 (Area)** - Business areas within a domain (e.g., "Digital Experience")
- **L3 (Capability)** - Specific capabilities (e.g., "Online Chat Support")
- **L4 (Sub-capability)** - Detailed sub-capabilities (e.g., "Chat Routing")

**Rules:**
- L1 capabilities cannot have parents
- L2-L4 capabilities must have a parent exactly one level above

### Dependency Types
- **Requires** - Hard dependency (source cannot function without target)
- **Enables** - Enabling relationship (source is enabled by target)
- **Supports** - Supportive relationship (source is supported by target)

### Metadata
- Strategy pillars, maturity levels, ownership model
- Primary owner, EA owner
- Status (Active, Deprecated, Planned)
- Experts (name, role, contact)
- Tags (flexible categorization)

## Code Structure

```
backend/internal/capabilitymapping/
├── domain/
│   ├── aggregates/
│   │   ├── capability.go                 # Capability aggregate root
│   │   ├── capability_dependency.go      # Dependency aggregate root
│   │   └── *_test.go                     # Unit tests
│   ├── valueobjects/
│   │   ├── capability_id.go
│   │   ├── capability_name.go
│   │   ├── capability_level.go
│   │   ├── dependency_type.go
│   │   └── ... (all value objects)
│   └── events/
│       ├── capability_created.go
│       ├── capability_dependency_created.go
│       └── ... (all domain events)
├── application/
│   ├── commands/                         # Command DTOs
│   ├── handlers/                         # Command handlers
│   ├── readmodels/                       # Query side read models
│   └── projectors/                       # Event → Read Model projectors
└── infrastructure/
    ├── api/                              # HTTP handlers & routes
    └── repositories/                     # Event sourcing repositories
```

## Quick Examples

### Creating a Capability

```bash
# Create L1 capability
curl -X POST http://localhost:8080/api/capabilities \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: dev" \
  -d '{
    "name": "Customer Engagement",
    "description": "All customer-facing capabilities",
    "level": "L1"
  }'

# Create L2 capability (with parent)
curl -X POST http://localhost:8080/api/capabilities \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: dev" \
  -d '{
    "name": "Digital Experience",
    "description": "Online customer touchpoints",
    "parentId": "<parent-capability-id>",
    "level": "L2"
  }'
```

### Creating a Dependency

```bash
curl -X POST http://localhost:8080/api/capability-dependencies \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: dev" \
  -d '{
    "sourceCapabilityId": "<source-id>",
    "targetCapabilityId": "<target-id>",
    "dependencyType": "Requires",
    "description": "Payment processing requires customer management"
  }'
```

### Updating Metadata

```bash
curl -X PUT http://localhost:8080/api/capabilities/{id}/metadata \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: dev" \
  -d '{
    "strategyPillar": "CustomerExcellence",
    "pillarWeight": 80,
    "maturityLevel": "Defined",
    "ownershipModel": "Centralized",
    "primaryOwner": "Product Team",
    "eaOwner": "Jane Smith",
    "status": "Active"
  }'
```

## Development Workflow

### Adding a New Feature

1. **Create/Update Spec** - In `/specs/` directory
2. **Domain Model** - Add value objects, aggregates, events
3. **Write Tests** - Unit tests for aggregates (TDD approach)
4. **Application Layer** - Commands, handlers, projectors
5. **Infrastructure** - API handlers, routes, migrations
6. **Run Tests** - `go test ./internal/capabilitymapping/...`

### Testing

```bash
# Run all capability mapping tests
go test ./internal/capabilitymapping/... -v

# Run specific test
go test ./internal/capabilitymapping/domain/aggregates -v -run TestNewCapability

# Build
go build ./...
```

## Key Design Patterns

### Event Sourcing
- All state changes captured as events
- Events stored in event store
- Aggregates reconstructed from event history

### CQRS
- Commands → Write side → Events → Event Store
- Events → Projectors → Read Models → Queries

### Value Objects
- All aggregate properties are value objects
- Encapsulate business invariants and validation
- Immutable

### HATEOAS
- All API responses include hypermedia links
- Clients discover available actions through links

## Database Migrations

Located in `backend/migrations/`:
- `008_add_capabilities_table.sql` - Capability read model
- `009_add_capability_metadata.sql` - Metadata tables
- `010_add_capability_dependencies_table.sql` - Dependency read model

Apply migrations:
```bash
cd backend
go run cmd/migrate/main.go
```

## Common Tasks

### Add a New Value Object
1. Create file in `domain/valueobjects/`
2. Implement: constructor with validation, `Value()`, `Equals()`, `String()`
3. Add tests

### Add a New Event
1. Create file in `domain/events/`
2. Implement: `EventType()`, `EventData()`, constructor
3. Update aggregate's `apply()` method
4. Update repository's `deserializeEvents()`
5. Update projector

### Add API Endpoint
1. Add handler method in `infrastructure/api/`
2. Add route in `routes.go`
3. Update HATEOAS links if needed

## Troubleshooting

### Tests Failing
- Ensure value objects validate inputs correctly
- Check aggregate invariants are enforced
- Verify event deserialization matches event data structure

### API Returns 500
- Check domain validation errors are caught and returned as 400
- Verify tenant context is set (X-Tenant-ID header)

### Read Model Out of Sync
- Check projector is subscribed to events
- Verify event data keys match projector expectations
- Check database migrations applied
