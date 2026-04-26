---
name: easi-frontend-e2e-testing
description: MUST load when UI/frontend changes are completed, to verify functionality by running the app in a browser. 
compatibility: opencode
---

# EASI Frontend E2E Testing

## The local dev backend runs via Docker/Podman Compose. 

EASI's full stack (backend + Postgres + Dex OIDC) runs locally via Docker Compose at the repo root. Use `podman compose up --build -d` if it's not running already.

| Service | URL | Purpose |
|---|---|---|
| Frontend (Vite) | `http://localhost:5173` | What the user sees. Started via `npm run dev` from `frontend/`. |
| Backend (Go API) | `http://localhost:8080` | REST API + OIDC callback. |
| Dex (OIDC) | `http://localhost:5556/dex` | Local OIDC provider with seeded users. |
| Postgres | `localhost:5432` | DB (user/pass `easi`/`easi`, db `easi`). |

## Dex test users (local OIDC)

Defined in `dex-config.yaml` at the repo root. **Password for all users is `password`.**

| Email | Role |
|---|---|
| `architect@acme.com` | architect |
| `admin@acme.com` | admin |
| `stakeholder@acme.com` | stakeholder (read-only) |
| `nono@acme.com` | persona-non-grata (deny test) |


## Testing with Playwright

1. **Run the app in a browser before claiming a UI change done** — build + tests are necessary but not sufficient. 
