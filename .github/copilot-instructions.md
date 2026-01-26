# EASI AI Coding Agent Instructions

## Project Overview
EASI is a multi-tenant enterprise architecture modeling tool using Strategic DDD with CQRS and Event Sourcing.

**Tech Stack:** Go backend (Chi, PostgreSQL event store + RLS), React frontend (Vite, TanStack Query, React Flow).

**Bounded Contexts:** `capabilitymapping`, `architecturemodeling`, `architectureviews`, `viewlayouts`, `metamodel`, `releases`, `importing` at `backend/internal/<context>/`.

## Quick Start

**Build & Test Backend (containerized, no local Go):**
```powershell
cd backend
.\go-dev.ps1 -- build -o bin/api cmd/api/main.go  # Use -- before dash args
.\go-dev.ps1 test ./...
bash ./test_integration.sh  # Requires bash/WSL/Git Bash
```

**Frontend:**
```bash
cd frontend
npm run dev      # Dev server
npm test -- --run
```

See [backend/README.container-dev.md](../backend/README.container-dev.md) for containerized workflow details.

## Critical Rules

1. **Never modify committed migrations or "done" specs** - forward-only
2. **No foreign keys** - use domain events for cross-aggregate integrity
3. **Check `_links` in UI** - never hardcode action availability (`{resource._links?.edit && ...}`)
4. **Multi-tenancy everywhere** - all read models filter by `tenant_id` from context
5. **No comments** unless explicitly requested

## Documentation Index

- **Architecture & DDD patterns:** [docs/architecture/README.md](../docs/architecture/README.md)
- **Backend conventions (HATEOAS, event sourcing, projectors):** [docs/backend/README.md](../docs/backend/README.md)
- **Frontend patterns (TanStack Query, cache invalidation):** [docs/frontend/README.md](../docs/frontend/README.md)
- **Complete index:** [docs/INDEX.md](../docs/INDEX.md)

When in doubt, consult the docs before implementing.
