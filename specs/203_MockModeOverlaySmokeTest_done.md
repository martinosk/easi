# 203 — Mock-Mode E2E Project and Overlay Smoke Test

> **Status:** ongoing
> **Depends on:** 202 (layer tokens)

---

## Problem Statement

Spec 201 shipped with "browser check pending — no compose stack in the implementation environment". The bug it shipped with (overflow menu hidden behind the explorer tree) is a *painting-order* bug: jsdom cannot see it, only a browser can. Yet the frontend already has a complete MSW mock mode (`npm run dev:mock`) that renders every page without a backend — it just is not wired into Playwright.

This spec adds a second Playwright project that runs against mock mode, and one smoke test that would have caught spec 201's bug on the first run: open every header overlay on every page at two window widths and assert it is actually the topmost element.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Frontend developer** | Run layout/overlay browser tests locally and in CI without Docker, Postgres or Dex |
| **Reviewer** | A green check that means "overlays render on top on every page", not "unit tests pass" |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Browser tests run without the backend

  Scenario: Mock project starts its own dev server
    Given no backend is running
    When "npx playwright test --project=mock" runs
    Then Vite starts in mock mode on its own port
    And the test logs in through the mocked session without a Dex round-trip

  Scenario: Header overlays are topmost on every page
    Given each routed page (canvas, business domains, value streams, enterprise architecture)
    And a wide viewport and a narrow viewport that forces the overflow menu
    When each header control that opens an overlay is clicked (More, user menu)
    Then the overlay is visible
    And the element at the overlay's centre point belongs to the overlay

  Scenario: Backend project is unaffected
    When "npx playwright test --project=chromium" runs
    Then only the existing backend-bound specs run, against the e2e dev server as before
```

---

## Business Rules & Invariants

1. **Two projects, two servers** — `chromium` keeps `dev:e2e` on 5173; `mock` runs `dev:mock` on 5174. Neither project matches the other's spec files.
2. **Topmost means topmost** — the assertion is `document.elementFromPoint` on the overlay's centre resolving inside the overlay. Visibility alone (`toBeVisible`) is not sufficient; it was true in the spec 201 bug.
3. **Every page, every width** — the page list and the viewport list are data; adding a routed page means adding one entry.
4. **No backend calls** — the mock project must not depend on `localhost:8080/8081`; unhandled requests bypass to the network and fail harmlessly.

---

## Acceptance Criteria

- [x] `playwright.config.ts` declares a `mock` project with its own `webServer` entry (port 5174, `dev:mock`), `testMatch` limited to `e2e/mock/**`, and the `chromium` project ignores that folder.
- [x] `e2e/mock/overlays.spec.ts` iterates pages × viewports, opens each header overlay, and asserts `elementFromPoint` on its centre is inside the overlay.
- [x] The test fails when the spec 201 fix is reverted (verified once during implementation: with both region isolations and the old media rule restored, the narrow canvas case failed with "overlay is painted behind page content") and passes on the current tree — 8/8.
- [x] `npm run test:e2e:mock` script exists and runs only the mock project.

---

## Architecture

### Ownership

Frontend test infrastructure only — `frontend/playwright.config.ts`, `frontend/e2e/mock/`, `frontend/package.json` scripts.

### Frontend

- Playwright's `webServer` accepts an array; each project sets its own `baseURL`. The mock server command is `npm run dev:mock -- --port 5174`.
- The mock session (`src/test/mocks/devHandlers.ts`) grants architect permissions, so canvas, business domains, value streams and enterprise architecture are all reachable; Users and One-Pager Quality are not (no `users:read`, no session link) and are outside this test's page list until the mock session grows.
- The narrow viewport must be narrow enough that the four visible entries overflow; the test derives it from the presence of `nav-more` rather than hard-coding a mode.

---

## Design Decisions

1. **Separate Playwright project rather than a mock flag on the existing one** — the backend project's specs mutate real data through the API and cannot run against MSW. Alternative: one project with per-test skips (rejected: confusing, slower).
2. **`elementFromPoint` assertion** — the only browser-level check that encodes painting order. Alternative: screenshot comparison (rejected: brittle, hides the cause).
3. **Header overlays only in this slice** — context menus, colour pickers and filter popovers are converted to Mantine in spec 204 and gain coverage there.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Second dev server in CI | ~5s extra startup | Reuses an existing server locally (`reuseExistingServer`) |
| Mock session is fixed | Pages gated on permissions the mock lacks are not covered | Page list is data; extend the mock session when those pages need coverage |

---

## Checklist

- [x] Specification ready — approved by user directive ("implement all 4", 2026-08-26)
- [x] Implementation done
- [x] Unit tests implemented and passing — none needed; the deliverable is the browser test
- [x] Integration tests implemented if relevant — `e2e/mock/overlays.spec.ts`
- [x] API documentation updated — no API change
- [x] User sign-off
