# EASI — Project Conventions

## Stack

| Layer | Technology |
|-------|-----------|
| Frontend | TypeScript, React 19, Vite, Vitest, Playwright |
| Backend | Go 1.26, chi router, PostgreSQL, Swagger/swag |
| Container | Docker, docker-compose |
| Package manager | npm (frontend) |

## Build & Test Commands

### Frontend (`frontend/`)

```bash
npm run build          # tsc + vite build
npm run lint           # eslint
npm test -- --run      # vitest (single run)
npm test               # vitest (watch)
npm run test:e2e       # playwright
npm run test:ci        # coverage + e2e
```

### Backend (`backend/`)

```bash
go build -o bin/api cmd/api/main.go   # build
go test ./...                          # unit tests (preferred)
.\go-dev.ps1 -- test ./...             # containerized fallback
make swagger                           # regenerate OpenAPI docs
make coverage                          # coverage report
```

## Conventions

### Frontend
- `"type": "module"` set — ESM throughout
- Zod for schema validation at API boundaries
- `@tanstack/react-query` for server state
- MSW (`msw`) for API mocking in tests
- Testing Library query priority: `getByRole` > `getByLabelText` > `getByText` > `getByTestId`
- Every mutation that creates/deletes an artifact **must** update `mutationEffects` to invalidate affected query caches

### Backend
- REST Level 3 with HATEOAS — all routes under `/api/v1/`
- Swagger annotations use relative paths (no `/api/v1/` prefix)
- Validation in domain model only — API translates exceptions to HTTP status codes
- No foreign keys — use domain model and events for relationships
- **Never modify a committed migration file** — create a new migration to fix issues
- Sequential migrations: 001, 002, 003…

### Architecture
- Strategic DDD: bounded contexts with business meaning
- No direct coupling between bounded contexts — use events
- Aggregates link by ID only, never by reference

## Activated Agent Templates

| Template | Purpose |
|----------|---------|
| `ts-enforcer` | TypeScript strict mode, no `any`, schema validation at boundaries |
| `esm-enforcer` | ESM-only, no `require()` or `module.exports` |
| `react-testing` | Testing Library best practices, behavior-first assertions |
| `front-end-testing` | Query priority, MSW usage, behavior-driven test names |
| `go-quality` | Error handling discipline, interface segregation, idiomatic Go |
| `twelve-factor-audit` | 12-factor compliance for containerized services |

Templates live in `.opencode/agents/`.
