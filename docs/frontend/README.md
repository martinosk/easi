# Frontend Conventions

React/TypeScript frontend development patterns.

## Documentation

| File | Purpose |
|------|---------|
| [standard-patterns.md](standard-patterns.md) | HATEOAS-driven UI, cache invalidation, mutations |

## Quick Reference

### Running Tests

```bash
npm test -- --run
```

### Core Principles

- HATEOAS-driven UI - never hardcode action availability
- Check `_links` presence for conditional rendering
- TanStack Query for data fetching
- Centralized cache invalidation via `mutationEffects`
- No optimistic updates for domain state

### Testing

- MSW handlers in `src/test/mocks/handlers.ts` mock all API endpoints
- When adding a new API endpoint consumed by the frontend, add a handler there — otherwise existing tests fail with "Cannot bypass a request" errors
- Keep filtering/transformation logic in pure `utils/` functions for easy unit testing without rendering components
