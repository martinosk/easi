---
name: easi-contract-first-development
description: MUST load when building a feature contract-first or frontend-before-backend in EASI — designing an API contract before code, mocking that contract, building a runnable frontend against the mock, then implementing the backend to the same contract. Load whenever you hear "frontend first", "stub/mock the API", "contract first".
compatibility: opencode
---

# EASI Contract-First Development

## Overview

Contract-first means the **API contract is the agreement that lets the frontend and backend be built independently**, in either order, and meet in the middle without rework. In EASI the common case is *frontend-first*: build and demo the UI against a mock API before the Go backend exists.

The point is a **frontend you can actually run and click through against the mock** — not just tests that pass.

Treat the four activities below as a **cycle traversed per thin slice**, not four big sequential stages you complete once.

## Invariants — true on every iteration

These hold continuously, no matter which loop you're on:

1. **The contract doc is the single source of truth.** Any change to a shape, status code, `_links` relation, or business rule lands in the contract doc **first**, then propagates to the mock, the TypeScript types, and the FE/BE **in the same change**. Never let the mock or a component drift ahead of the doc.
2. **The mock conforms to the current contract and is wired into both transports** — the test runner (node server) **and** the running app (browser worker).
3. **The FE gates affordances on `_links`** (`easi-frontend-data`), never on hardcoded rules — so a contract change to who-can-do-what needs no component rewrite.

## The Loop

Four activities. Traverse them for one **thin vertical slice** (one screen + its endpoint(s) + states, end to end), learn, fold back, then take the next slice. 

### 1 — Contract (a slice of it)

Design the HTTP contract for the slice you're about to build. Use the `api-design-expert` agent and `easi-api-standards`.

- Per endpoint: method, path, params (types, required/optional), request body, success response JSON (every field, type, nullability), all status codes, and exact error bodies (`ErrorResponse`: `message`/`details`/`_links`).
- The HATEOAS `_links` matrix for this slice: which relations appear, under what state. These gate UI affordances — part of the contract, not an afterthought.
- The business rules that change responses (eligibility, immutability, cardinality, …), precise enough to reimplement in Go.

Artifact: a contract doc in `/specs` (e.g. `NNN_Feature_api-contract.md`) with a short **revision log** at the top. Each iteration appends a dated line describing what changed and why.

### 2 — Mock (the slice)

Implement MSW handlers + an in-memory store that conform to the contract **exactly** for this slice: same field names/types/nullability, status codes, error bodies, `_links`, and enough business-rule logic to drive **every** UI state the slice renders (empty, loaded, each role/status, each rejection).

### 3 — Frontend (the slice), against the mock

Build the UI with TDD (`easi-test-driven-development`) against the mock at the HTTP boundary — not by mocking the api-client module — and **run it**:

- Tests go through MSW (node server): validates api client + hooks + contract together.
- `npm run dev` (mock flag on) goes through MSW (browser worker): the demoable deliverable. Walk every state. This is where the contract's flaws surface.
- When a flaw surfaces, **go back to activity 1** for this slice: amend the doc, then the mock, types, and FE together. That round trip is the method working, not a failure.

### 4 — Backend, to the same contract

When the backend is built (now, or a later iteration), it implements to the contract doc; the FE swaps mock→real by turning the flag off, unchanged. Integration mismatches are contract bugs: fix them in the **doc first**, then mock + FE + BE — never patch one side silently. `flag defaults off`.

## Architecture — where the mock lives and how to wire it

### Single source of handlers, two transports

```
src/test/mocks/
  handlers.ts        # all request handlers (the contract) — the single source
  server.ts          # msw/node  setupServer(...handlers)  → used by vitest
  browser.ts         # msw/browser setupWorker(...handlers) → used by `npm run dev`
  <feature>/
    handlers.ts      # feature handlers, spread into the root handlers array
    store.ts         # in-memory state + DTO/HATEOAS builders (the contract shapes)
    resolver.ts      # pure business-rule logic (unit-tested directly)
    seed.ts          # realistic demo dataset for dev runtime
```

The handlers are written **once** and consumed by both `server.ts` (Node, for tests) and `browser.ts` (browser, for dev). Never fork them. A bug fixed in the mock is fixed for both.

### Tests — node server

`src/test/setup.ts` calls `server.listen({ onUnhandledRequest: 'error' })` and resets the store in `beforeEach`. Tests seed per-case state. `onUnhandledRequest: 'error'` is deliberate — every endpoint the UI touches must have a handler, which keeps the mock honest against the contract.

### Dev runtime — browser worker

1. Generate the service worker file once: `npx msw init public/ --save` (creates `public/mockServiceWorker.js`; commit it).
2. Create `src/test/mocks/browser.ts`:
   ```ts
   import { setupWorker } from 'msw/browser';
   import { handlers } from './handlers';
   export const worker = setupWorker(...handlers);
   ```
3. Start it in `main.tsx` **behind an env flag, via dynamic import**, before rendering, and seed the dev dataset:
   ```ts
   async function enableMocking() {
     if (import.meta.env.VITE_USE_MOCK_API !== 'true') return;
     const { worker } = await import('./test/mocks/browser');
     const { seedDevData } = await import('./test/mocks/<feature>/seed');
     seedDevData();
     await worker.start({ onUnhandledRequest: 'bypass' });
   }
   enableMocking().then(() => createRoot(...).render(...));
   ```
4. Add the flag to `.env.example` (`VITE_USE_MOCK_API=false`) and document `VITE_USE_MOCK_API=true npm run dev` as the contract-first run mode.

### Best-practice rules for the wiring

- **Flag-gated + dynamic import.** The mock must never reach the production bundle. Gate on an env var and `await import(...)` so the bundler tree-shakes it out when the flag is off. Default the flag off.
- **`onUnhandledRequest`: `error` in tests, `bypass` in dev.** Tests must fail on an unstubbed call (keeps the mock complete). Dev should pass through anything unmocked (e.g. real auth) rather than crash.
- **Same origin as the real client.** Handlers must match the URL the `httpClient` actually calls (`VITE_API_URL`, e.g. `http://localhost:8080`). If they don't match, the worker silently won't intercept.
- **Seed for dev, seed per-test for tests.** Dev needs one realistic dataset that shows the feature well; tests seed minimal per-case state and reset between cases.
- **Contract fidelity over convenience.** Emit real status codes and error bodies, and the exact `_links` per state — the UI gates on them. A mock that returns `200 {}` everywhere cannot validate the frontend.
- **Types derive from the contract.** The TypeScript response types and the mock both reflect the contract doc; when the contract changes, change the types, the mock, and the FE together.
- **Keep business rules in a pure, tested module.** Non-trivial logic (eligibility, carve-outs, derived views) goes in a `resolver.ts` unit-tested directly, not buried in handler closures — so the mock's behavior is itself trustworthy.

## Definition of Done

**Per slice (every loop):**
- [ ] Contract doc updated for the slice (with a revision-log line), open questions marked.
- [ ] Mock covers the slice's endpoints with real status codes, error bodies, and `_links`.
- [ ] Mock runs in both `vitest` (node) and `npm run dev` (browser worker, flag-gated).
- [ ] FE for the slice built TDD; tests pass through the mock; the slice walked in the running app.

**Per feature (before calling the FE work done):**
- [ ] Every screen/state reachable in `npm run dev` against the mock, no backend running — verified by actually running it.
- [ ] Contract doc is internally consistent and its open questions are either resolved or explicitly handed to backend.
- [ ] Flag defaults off; mock excluded from the production bundle.

## Anti-Patterns

| Anti-pattern | Why it's wrong | Do instead |
|---|---|---|
| **Big-bang contract** — specify the whole feature's contract before building any UI | You're guessing; the first screen built will prove parts wrong, and you've over-invested in spec | Contract one thin slice, build it, learn, expand |
| **Frozen contract** — refusing to change the contract when the FE/BE reveals a flaw | Forces workarounds in code; the doc rots into a lie | Amend the doc first, then mock + types + FE/BE; it's a living artifact |
| **Silent drift** — changing the mock or a component without updating the doc | Doc no longer matches reality; backend implements the wrong thing | Doc is source of truth; change it first, propagate atomically |
| **Mock only in `vitest` setup** | App isn't runnable/demoable; contract unvalidated by real use | Wire the browser worker into `main.tsx` behind a flag |
| **`vi.mock('../api/...')` per test** | Tests your assumptions, skips the api client + contract | Mock at the HTTP boundary with MSW |
| **Forking test vs dev handlers** | They drift; bugs fixed in one persist in the other | One handler source, two transports |
| **Mock returns `200 {}` / omits `_links`** | UI states and affordances can't be exercised or gated | Match the contract: status codes, error bodies, `_links` per state |
| **Hardcoding permissions in components** | Breaks the moment the backend gates differently | Read `_links`; let the contract gate affordances |
| **Mock bundled into production** | Ships a fake API to users | Env flag + dynamic import so it tree-shakes out |

## Related skills

- `easi-api-standards` — the contract's HTTP/HATEOAS conventions.
- `easi-frontend-data` — consuming `_links`, query keys, mutation hooks, invalidation.
- `easi-test-driven-development` — the RED/GREEN loop the FE is built with.
- `easi-spec-driven-development` — where the contract doc and feature spec live and how they evolve.
- `easi-frontend-e2e-testing` — running the app to verify the mock-backed UI.
