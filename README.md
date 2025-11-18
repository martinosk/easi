# easi - Enterprise Architecture - Simple 
Simple, modern tool for modelling, documenting and analysing enterprise architecture.

## Spec-Driven Development
All specs are in /specs. Code and documentation must follow existing specifications.

### Spec Format
All specs must contain a description and checklist:
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off

If there's no check-mark in "Specification ready", do not implement, but ask user to verify the spec first.
Update spec checklist when contributing.

### Spec Naming
- `001_ShortDescription_pending.md` - not yet implemented
- `001_ShortDescription_ongoing.md` - in development
- `001_ShortDescription_done.md` - fully implemented

## Architecture
Domain-Driven Design with CQRS and Event Sourcing for core domains.
Supporting domains can use CRUD or whatever other architecture makes sense for their purpose.

### Current Architecture Summary
The system uses event sourcing for core aggregates, REST Level 3 APIs with HATEOAS, and a clean separation between domain models and infrastructure.

**Implemented Aggregates:**
- **ApplicationComponent** - Architecture components with relations
- **ComponentRelation** - Component relationships (Triggers, Serves)
- **ArchitectureView** - Graphical view layouts with positions
- **Capability** - Business capabilities with hierarchical structure (L1-L4)
- **CapabilityDependency** - Dependencies between capabilities (Requires, Enables, Supports)

**Features:**
- Event-sourced domain models with read model projections
- Multi-tenant data isolation
- REST Level 3 APIs with HATEOAS links
- Capability hierarchy modeling (4 levels)
- Capability metadata (strategy alignment, maturity, ownership, experts, tags)
- Capability dependency tracking

### Bounded contexts
#### ArchitectureModeling
This is the core domain that supports and enforces best practices for architecture modelling and documentation.
Focus is on enterprise architecture modelling in the style of ArchiMate, but an opinionated limited subset.

#### ArchitectureViews
This is a supporting domain that allows for visualisations of the architecture model.
A key trait of Easi is that views are separate from the model.
It is considered supporting, because the API and event first approach of Easi allows for complete freedom of creating views using other tools (COTS reporting solutions, OSS libraries etc)

#### CapabilityMapping
Core domain for enterprise capability modeling. Uses CQRS with event sourcing.

**Implemented (Specs 023-025):**
- **Capability Hierarchy** - 4-level capability model (L1: Domain, L2: Area, L3: Capability, L4: Sub-capability)
- **Capability Metadata** - Strategy pillars, maturity levels, ownership model, experts, tags
- **Capability Dependencies** - Model relationships between capabilities (Requires, Enables, Supports)

**API Endpoints:**
- `POST/GET/PUT /api/capabilities` - Manage capabilities
- `GET /api/capabilities/{id}/children` - Get child capabilities
- `PUT /api/capabilities/{id}/metadata` - Update metadata
- `POST /api/capabilities/{id}/experts` - Add experts
- `POST /api/capabilities/{id}/tags` - Add tags
- `POST/GET/DELETE /api/capability-dependencies` - Manage dependencies
- `GET /api/capabilities/{id}/dependencies/outgoing` - Dependencies this capability has
- `GET /api/capabilities/{id}/dependencies/incoming` - Dependencies on this capability

**Planned (Specs 026-029):**
- System realization links to architecture components
- Custom perspectives and viewpoints
- Capability versioning

### ArchitectureAnalysis
Core domain that allows the gathering and analysis of architecture knowledge. It supports the architecture modelling process.

┌─────────────────────────────────────────────────────────┐
│                      Browser                            │
│  ┌─────────────────────────────────────────────────┐   │
│  │         React Frontend (Port 5173)              │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │   │
│  │  │ Canvas   │  │ Dialogs  │  │ Details  │     │   │
│  │  │ (React   │  │ (Create) │  │ (View)   │     │   │
│  │  │  Flow)   │  └──────────┘  └──────────┘     │   │
│  │  └────┬─────┘                                  │   │
│  │       │ API Client (Axios)                     │   │
│  └───────┼────────────────────────────────────────┘   │
│          │                                             │
└──────────┼─────────────────────────────────────────────┘
           │ HTTP/JSON
           │
┌──────────▼─────────────────────────────────────────────┐
│              Go Backend (Port 8080)                    │
│  ┌─────────────────────────────────────────────────┐  │
│  │            RESTful API Layer                    │  │
│  │  (Chi Router, CORS, Middleware)                 │  │
│  └────────┬────────────────────────────────────────┘  │
│           │                                            │
│  ┌────────▼────────────────────────────────────────┐  │
│  │         CQRS Command/Query Buses                │  │
│  └────────┬────────────────────────────────────────┘  │
│           │                                            │
│  ┌────────▼────────────────────────────────────────┐  │
│  │        Bounded Contexts (DDD)                   │  │
│  │  ┌──────────────┐  ┌──────────────┐            │  │
│  │  │Architecture  │  │Architecture  │            │  │
│  │  │Modeling      │  │Views         │            │  │
│  │  │(Components,  │  │(Positions)   │            │  │
│  │  │ Relations)   │  │              │            │  │
│  │  └──────┬───────┘  └──────┬───────┘            │  │
│  │                                                 │  │
│  │  ┌──────────────────────────────────┐          │  │
│  │  │CapabilityMapping                 │          │  │
│  │  │(Capabilities, Dependencies,      │          │  │
│  │  │ Metadata, Experts, Tags)         │          │  │
│  │  └──────┬───────────────────────────┘          │  │
│  └─────────┼──────────────────┼───────────────────┘  │
│            │                  │                       │
│  ┌─────────▼──────────────────▼───────────────────┐  │
│  │         Event Store (PostgreSQL)               │  │
│  │  - All events (audit trail)                    │  │
│  │  - Event sourcing                              │  │
│  └────────────────────────────────────────────────┘  │
│                                                       │
│  ┌────────────────────────────────────────────────┐  │
│  │         Read Models (PostgreSQL)               │  │
│  │  - Components, Relations                       │  │
│  │  - Views, Positions                            │  │
│  │  - Capabilities, Dependencies                  │  │
│  │  - Capability Metadata, Experts, Tags          │  │
│  └────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────┘


### Structure
- Bounded contexts organize the codebase
- RESTful APIs (maturity level 3)

## Tech Stack
- **Backend**: Go
- **Frontend**: React, TypeScript
- **API**: OpenAPI specifications
- **Containers**: Docker/Podman

## Prerequisites
- Docker or Podman

## Setup

### First-Time Setup
```bash
# Set up environment variables
./setup-local-env.sh

# Start database and services
docker compose up -d
```

### Environment Configuration
The project uses environment variables for configuration. On first setup, run `./setup-local-env.sh` to create a `.env` file with default development values.

## Database
PostgreSQL 16

## Testing

### Running E2E Tests
```bash
# Start the test environment
docker compose -f docker-compose.e2e.yml up -d

# Run the e2e tests
cd frontend
npm run test:e2e

# Clean up
docker compose -f docker-compose.e2e.yml down
```

**Note**: Some tests may fail if the database is not clean between runs. This will be addressed in a future update when tests run as isolated tenants.
