# EASI Development Guidelines

## Documentation

Find detailed patterns and conventions in `/docs/`:

| Working on | Read |
|------------|------|
| Quick lookup | [docs/INDEX.md](docs/INDEX.md) |
| Backend | [docs/backend/README.md](docs/backend/README.md) |
| Frontend | [docs/frontend/README.md](docs/frontend/README.md) |
| Architecture | [docs/architecture/README.md](docs/architecture/README.md) |

## Core Rules

### Code Style
- Never add comments unless explicitly asked
- Always verify build and tests after modifying files

### Architecture
- Strategic DDD: structure code by bounded contexts with business meaning
- No direct coupling between bounded contexts - use events
- Tactical DDD: aggregates as transactional boundaries, value objects for all properties
- Aggregates link by ID only, never by reference
- API first: all functionality via API calls

### Spec Management
- **NEVER modify a spec file with "done" status**
- Specs contain only what is to be implemented NOW
- Status workflow: `pending` → `ongoing` → `done`

### Database Migrations
- **NEVER modify a committed migration file**
- Forward-only, sequential (001, 002, 003...)
- To fix issues: create a new migration
- No foreign keys - use domain model and events

### API Principles
- REST Level 3 with HATEOAS
- All routes resolve to `/api/v1/` prefix
- Swagger annotations use relative paths (no `/api/v1/` prefix)
- Validation in domain model only - API translates exceptions to status codes

## Running Tests

```bash
# Backend
go test ./...

# Frontend
npm test -- --run
```
