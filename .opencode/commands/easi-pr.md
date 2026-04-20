# /pr — Create a Pull Request

Run the full quality pipeline, generate release notes if this is a release branch, then open a pull request.

## Steps

1. **Lint** — `npm run lint` (from `frontend/`)
2. **Type-check** — `tsc -b` (from `frontend/`)
3. **Unit tests** — `npm run test` (from `frontend/`)
4. **Backend tests** — `make test` (from `backend/`)
5. **Build** — `npm run build` (from `frontend/`) + `make build` (from `backend/`)
6. **Release notes** — run `/release-notes` to draft, review, commit, and tag the release. Skip this step if the PR is not a release (e.g. a feature branch mid-development).
7. If all steps pass, create a pull request with a clear title and description summarising the changes.
8. If any step fails, fix the issue and re-run that step before proceeding.

## Notes

- E2E tests (`npm run test:e2e`) are optional for PRs targeting non-release branches.
- Swagger docs should be regenerated (`make swagger` from `backend/`) if any API handlers changed.
- Do not merge if lint, type-check, unit tests, or build fail.
- The `/release-notes` step creates its own commit and tag — do not squash it away when merging.
