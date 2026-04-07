# AGENTS.md

## Project: EASI (frontend + backend)

This file describes the tech stack, tooling, commands, and conventions for AI coding agents working in this repository.

---

## Tech Stack

| Layer    | Root          | Language   | Frameworks / Libraries                                                                          |
|----------|---------------|------------|-------------------------------------------------------------------------------------------------|
| Frontend | `frontend/`   | TypeScript | React 19, Vite 7, Vitest 4, Playwright, React Query, Zustand, Mantine, React Router, Zod, MSW  |
| Backend  | `backend/`    | Go         | chi v5, go-oidc, swaggo/swag, PostgreSQL (lib/pq), SCS sessions                                |
| Infra    | repo root     | Docker     | docker-compose (local), Kubernetes manifests in `k8s/`                                         |

---

## Package Manager

| Layer    | Manager |
|----------|---------|
| Frontend | npm (lockfile: `frontend/package-lock.json`) |
| Backend  | Go modules (`backend/go.mod`) |

---

## Formatter & Linter

| Layer    | Formatter                        | Linter                                                  | Detection    |
|----------|----------------------------------|---------------------------------------------------------|--------------|
| Frontend | ESLint (no separate formatter)   | ESLint — `frontend/eslint.config.js`                    | **Detected** |
| Backend  | gofmt (via golangci-lint)        | golangci-lint — `backend/.golangci.yml`                 | **Detected** |

> No Prettier, Biome, or standalone formatter config was found in `frontend/`. ESLint is the sole configured tool.

---

## Build / Test / Lint Commands

### Frontend (`frontend/`)

```bash
# Install dependencies
npm install

# Development server
npm run dev

# Production build
npm run build

# Unit tests (Vitest)
npm run test

# Unit tests with coverage
npm run test:ci

# E2E tests (Playwright)
npm run test:e2e

# Lint
npm run lint
```

### Backend (`backend/`)

```bash
# Build binary
make build          # go build -o bin/api cmd/api/main.go

# Run
make run            # go run cmd/api/main.go

# Unit tests
make test           # go test -v ./...

# Tests with coverage
make coverage

# Lint (golangci-lint)
golangci-lint run --fix ./...

# Generate Swagger docs
make swagger
```

---

## JS/TS Module Notes

- **ES Modules**: `"type": "module"` is set in `frontend/package.json`. The project is fully ESM.
- **require() usage**: 0 occurrences found in `frontend/src/` — no CommonJS migration candidates.
- **TypeScript**: `frontend/tsconfig.json` exists. No scaffolding needed.

---

## Conventions

- Frontend source lives in `frontend/src/`.
- Backend source lives in `backend/internal/` and `backend/cmd/`.
- All API docs are generated via swaggo; do not edit `backend/docs/` by hand.
- Environment variables follow `.env.example` patterns in each sub-directory.
- Docker Compose (`docker-compose.yml`) starts local PostgreSQL and Dex (OIDC).
