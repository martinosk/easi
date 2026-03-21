---
description: Run quality gates and create a pull request for the EASI project
---

# PR — Pull Request Workflow

This command runs all quality gates and creates a pull request.

## Quality Gates

Run these checks before creating the PR. All must pass.

### Frontend (`frontend/`)

```bash
npm run build          # TypeScript compile + Vite build
npm run lint           # ESLint
npm test -- --run      # Vitest unit tests (single run)
npm run test:e2e       # Playwright end-to-end tests
```

### Backend (`backend/`)

```bash
go build -o bin/api cmd/api/main.go   # compile check
go test ./...                          # unit tests
make swagger                           # ensure OpenAPI docs are current
```

### Code Review

```bash
/code-review --changed                 # run review agents on changed files
```

## PR Creation

After all gates pass:

1. Ensure the branch is pushed: `git push -u origin HEAD`
2. Create the PR:

```bash
gh pr create \
  --title "<title>" \
  --body "$(cat <<'EOF'
## Summary
- <bullet 1>
- <bullet 2>

## Testing
- [ ] Frontend unit tests pass (`npm test -- --run`)
- [ ] Backend unit tests pass (`go test ./...`)
- [ ] E2E tests pass (`npm run test:e2e`)
- [ ] No linting errors (`npm run lint`)
- [ ] Build succeeds (`npm run build`)
EOF
)"
```

## Merge Strategy

- Prefer **squash merge** for feature branches to keep history clean
- Use **merge commit** for long-lived branches that need preserved history

## Notes

- Never force-push to `main`
- `gofmt` is not installed in the local dev environment — Go formatting relies on CI
- Frontend prettier + eslint run automatically via `.opencode/plugins/formatter.js`
