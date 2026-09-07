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

Defined in `dex-config.yaml` at the repo root. **Password for all users is `password`.** Tenant is `acme` (domain `acme.com`).

| Email | Role |
|---|---|
| `architect@acme.com` | architect |
| `admin@acme.com` | admin |
| `stakeholder@acme.com` | stakeholder (read-only) |
| `nono@acme.com` | persona-non-grata (deny test) |

## Login requires an invitation, not just a dex user

Dex only authenticates. The OIDC callback then requires an invitation for the email — without one, the backend logs `no valid invitation for email <email>` and the app bounces back to the login screen with no visible error.

Seed invitations through the platform admin API (header `X-Platform-Admin-Key: localdev` locally). The invitation route is **`/auth/invitations`**, with the tenant in the body — not a sub-resource of the tenant:

```
curl -H "X-Platform-Admin-Key: localdev" http://localhost:8080/api/v1/platform/tenants
curl -X POST -H "X-Platform-Admin-Key: localdev" -H "Content-Type: application/json" \
  -d '{"tenantId":"acme","email":"architect@acme.com","role":"architect"}' \
  http://localhost:8080/api/v1/auth/invitations
```

**Never seed `auth.invitations` with SQL.** Invitations are event-sourced: the login path loads the aggregate from `infrastructure.events`, so a hand-inserted read-model row authenticates and then fails with `invitation not found`. Go through the API.

Check current invitations: `podman exec easi-postgres psql -U easi -d easi -c "SELECT email, role, status FROM auth.invitations;"`


## Testing with Playwright

1. **Run the app in a browser before claiming a UI change done** — build + tests are necessary but not sufficient.
2. **Restart the backend after Go changes.** `podman compose up -d --build backend`. Verify with curl before browser testing. The `migrate` container image goes stale the same way — `podman compose build migrate` when the DB is missing recent migrations.
3. **Test the full golden path.** Create → list → act → verify list updates. Not just "API returns 200".
4. **Test empty states, error states, and permission gating.** Navigate with no data. Submit invalid input. Log in as stakeholder.
5. **Headless typing into Mantine/React controlled inputs may not register.** If a typed value doesn't appear, set it via the native value setter and dispatch an `input` event in page context: `Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set.call(input, text); input.dispatchEvent(new Event('input', {bubbles: true}))`. Plain dex forms accept direct `.value` assignment + `form.submit()`.
