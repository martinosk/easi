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

### Bounded contexts
#### ArchitectureModeling
This is the core domain that supports and enforces best practices for architecture modelling and documentation.
Focus is on enterprise architecture modelling in the style of ArchiMate, but an opinionated limited subset.

#### ArchitectureViews
This is a supporting domain that allows for visualisations of the architecture model.
A key trait of Easi is that views are separate from the model.
It is considered supporting, because the API and event first approach of Easi allows for complete freedom of creating views using other tools (COTS reporting solutions, OSS libraries etc)

### ArchitectureAnalysis
Core domain that allows the gathering and analysis of architecture knowledge. It supports the architecture modelling process.

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
```bash
# Start database
docker compose up -d

```

## Database
PostgreSQL 16
