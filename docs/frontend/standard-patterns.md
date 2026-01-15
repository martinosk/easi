# Frontend Standard Patterns

## HATEOAS-Driven UI

**Never hardcode action availability.** The backend controls what actions are available through `_links` in API responses.

### Link Structure

```typescript
interface HATEOASLinks {
  self?: { href: string; method: string };
  edit?: { href: string; method: string };
  delete?: { href: string; method: string };
  [key: string]: { href: string; method: string } | undefined;
}
```

### Utilities

Use helpers from `src/utils/hateoas.ts`:

```typescript
import { hasLink, canEdit, canDelete } from '../utils/hateoas';

if (canEdit(resource)) { /* show edit button */ }
```

### Standard Relations

| Relation | Purpose | Custom Relations |
|----------|---------|------------------|
| `self` | Current resource | `x-children` |
| `edit` | Update resource | `x-remove` |
| `delete` | Delete resource | `x-create-link` |
| `collection` | Parent collection | |

### Conditional Rendering

```tsx
// CORRECT - gate on link presence
{resource._links?.edit && <button onClick={handleEdit}>Edit</button>}

// WRONG - hardcoded logic
{userRole === 'admin' && <button>Edit</button>}
```

### Type Definitions

```typescript
interface Capability {
  id: CapabilityId;
  name: string;
  _links: HATEOASLinks;  // Required
}
```

---

## Cache Invalidation & Mutations

TanStack Query with centralized cache invalidation.

### Key Files

| File | Purpose |
|------|---------|
| `src/lib/queryClient.ts` | QueryClient config (5min stale, 30min gc) |
| `src/lib/queryKeys.ts` | Hierarchical query key definitions |
| `src/lib/mutationEffects.ts` | Cache invalidation rules per mutation |
| `src/lib/invalidateFor.ts` | Helper to invalidate multiple keys |

### Patterns

1. **Query keys** are hierarchical: `all` → `lists()` → `detail(id)`. Add new domains to `queryKeys.ts`.

2. **Mutation effects** define which queries to invalidate per mutation type. Add new effects to `mutationEffects.ts`.

3. **Mutation hooks** follow the pattern: call API → `invalidateFor(queryClient, mutationEffects.x.y())` → show toast.

4. **Conditional queries** use `enabled: !!dependency` to wait for required data.

5. **Static metadata** uses `staleTime: Infinity`.

6. **Optimistic updates** are avoided for domain state. Always wait for server confirmation.

### Reference Implementation

See `src/features/components/hooks/useComponents.ts` for the standard patterns.

---

## Quick Reference

| Aspect | Pattern |
|--------|---------|
| Action availability | Check `_links` presence, never hardcode |
| Query keys | Hierarchical: `all` → `lists()` → `detail(id)` |
| Cache invalidation | Centralized in `mutationEffects` |
| Mutation structure | `mutationFn` → `invalidateFor` → `toast` |
| Static data | `staleTime: Infinity` |
| Optimistic updates | Avoid for domain state |
