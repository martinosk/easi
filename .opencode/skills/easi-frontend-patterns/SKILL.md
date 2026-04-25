---
name: easi-frontend-patterns
description: MUST load when writing or reviewing any frontend TypeScript/React code in EASI. Load when building UI components that show/hide actions, writing data-fetching hooks, setting up cache invalidation, or adding new mutation hooks.
compatibility: opencode
---

# EASI Frontend Patterns

## Overview

The EASI frontend is HATEOAS-driven: the backend controls what actions are available through `_links` in API responses. Client code never hardcodes business rules about permissions or role-based action availability. Data fetching uses TanStack Query with a centralized cache invalidation strategy.

## HATEOAS-Driven UI

### Link Type Definition

```typescript
interface HATEOASLinks {
  self?: { href: string; method: string };
  edit?: { href: string; method: string };
  delete?: { href: string; method: string };
  [key: string]: { href: string; method: string } | undefined;
}
```

All resource types must include `_links` as a required field:

```typescript
interface Capability {
  id: CapabilityId;
  name: string;
  _links: HATEOASLinks;  // Required — not optional
}
```

### HATEOAS Utilities

Use helpers from `src/utils/hateoas.ts`:

```typescript
import { hasLink, canEdit, canDelete } from '../utils/hateoas';

if (canEdit(resource)) { /* show edit button */ }
if (canDelete(resource)) { /* show delete button */ }
if (hasLink(resource, 'x-children')) { /* show expand control */ }
```

### Standard Link Relations

| Relation | Purpose |
|----------|---------|
| `self` | Current resource URL |
| `edit` | Update resource |
| `delete` | Delete resource |
| `collection` | Parent collection |
| `x-children` | Child resources |
| `x-remove` | Remove from relationship |
| `x-create-link` | Create association |

### Conditional Rendering — Gate on Link Presence

```tsx
// CORRECT — driven by backend permission model
{resource._links?.edit && (
  <button onClick={handleEdit}>Edit</button>
)}

// CORRECT — using utility helper
{canDelete(resource) && (
  <button onClick={handleDelete}>Delete</button>
)}

// WRONG — hardcoded business logic in the client
{userRole === 'admin' && <button>Edit</button>}
{!resource.isPrivate && <button>Edit</button>}
```

## Cache Invalidation & Mutations (TanStack Query)

### Key Files

| File | Purpose |
|------|---------|
| `src/lib/queryClient.ts` | QueryClient config (5min stale time, 30min gc) |
| `src/lib/queryKeys.ts` | Hierarchical query key definitions |
| `src/lib/mutationEffects.ts` | Cache invalidation rules per mutation type |
| `src/lib/invalidateFor.ts` | Helper to invalidate multiple keys atomically |

### Query Key Hierarchy

Query keys are hierarchical: `all` → `lists()` → `detail(id)`.

When adding a new domain:
1. Add its key factory to `src/lib/queryKeys.ts`
2. Follow the `all → lists() → detail(id)` shape

```typescript
// Pattern
const capabilityKeys = {
  all: ['capabilities'] as const,
  lists: () => [...capabilityKeys.all, 'list'] as const,
  detail: (id: string) => [...capabilityKeys.all, 'detail', id] as const,
};
```

### Mutation Hook Pattern

Standard mutation hooks follow: **call API → invalidate cache → show toast**.

```typescript
const mutation = useMutation({
  mutationFn: (data) => api.updateCapability(id, data),
  onSuccess: () => {
    invalidateFor(queryClient, mutationEffects.capabilities.update());
    toast.success('Capability updated');
  },
  onError: (err) => {
    toast.error('Failed to update capability');
  },
});
```

### Mutation Effects

Each mutation type has defined cache invalidation rules in `mutationEffects.ts`. When adding a new mutation:
1. Define the invalidation rules in `mutationEffects.ts`
2. Call `invalidateFor(queryClient, mutationEffects.x.y())` in `onSuccess`

### Conditional Queries

Use `enabled` to wait for required data:

```typescript
const { data: capability } = useQuery({
  queryKey: capabilityKeys.detail(id),
  queryFn: () => api.getCapability(id),
  enabled: !!id,  // wait for id to be available
});
```

### Static Metadata

Use `staleTime: Infinity` for data that never changes at runtime:

```typescript
const { data: metamodel } = useQuery({
  queryKey: metamodelKeys.configuration(),
  queryFn: api.getMetaModelConfiguration,
  staleTime: Infinity,
});
```

### Optimistic Updates

Avoid optimistic updates for domain state. Always wait for server confirmation before updating the UI.

```typescript
// CORRECT — wait for server
onSuccess: () => {
  invalidateFor(queryClient, mutationEffects.capabilities.delete());
}

// WRONG — optimistic removal before server confirms
onMutate: (id) => {
  queryClient.setQueryData(capabilityKeys.lists(), (old) =>
    old?.filter(c => c.id !== id)
  );
}
```

## Reference Implementation

See `src/features/components/hooks/useComponents.ts` for the canonical example of all these patterns together.

## Quick Reference

| Aspect | Pattern |
|--------|---------|
| Action availability | Check `_links` presence — never hardcode |
| Role-based UI | Never check `userRole` to show/hide actions |
| Query keys | Hierarchical: `all → lists() → detail(id)` |
| Cache invalidation | Centralized in `mutationEffects` |
| Mutation structure | `mutationFn` → `invalidateFor` → `toast` |
| Static data | `staleTime: Infinity` |
| Optimistic updates | Avoid for domain state |
| Conditional queries | `enabled: !!dependency` |
| Mantine components     | Use for all UI surfaces  |
| Dialog gates           | Grep for all call sites and update every surface before marking complete |

## UI Component Framework (Mantine v8)

All interactive UI surfaces in EASI use **Mantine v8** (`@mantine/core`). Never write plain HTML for dialogs, buttons, inputs, or form controls.

### Component Mapping

| UI Surface | Mantine Component |
|---|---|
| Dialogs / modals | `Modal` |
| Number inputs | `NumberInput` |
| Checkboxes | `Checkbox` |
| Buttons (primary / default) | `Button` (with `variant` prop) |
| Vertical spacing | `Stack` |
| Horizontal grouping | `Group` |
| Body text | `Text` |

### Test wrapper

Mantine component tests must be wrapped in `MantineTestWrapper`:

```typescript
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

render(
  <MantineTestWrapper>
    <MyComponent />
  </MantineTestWrapper>
);
```

Rendering without the wrapper causes test failures (missing MantineProvider context).

---

## Multi-Surface Feature Entry Points

Features in EASI can be triggered from multiple surfaces — canvas context menu, navigation tree context menu, toolbar buttons, or keyboard shortcuts. When you add a **dialog gate** (an interstitial dialog that intercepts an action to collect user input or confirmation), you must update **every** call site of the underlying action hook, not just the surface you are actively coding.

### Audit call sites before implementing a dialog gate

```bash
grep -r "yourHookFunction(" frontend/src/ --include="*.tsx" --include="*.ts"
```

### Checklist when adding a dialog gate

- [ ] `grep` for all current call sites of the intercepted hook function
- [ ] Update every call site to route through the dialog gate state
- [ ] Add a test per surface (unit or E2E) that verifies the dialog appears

## Guidelines

1. **Gate all action visibility on `_links` presence** — never on role, user ID, or flag
2. **Use `canEdit`, `canDelete`, `hasLink`** from `src/utils/hateoas.ts`
3. **All resource interfaces must include `_links: HATEOASLinks`** as a required field
4. **Register all query keys** in `src/lib/queryKeys.ts` with the hierarchical pattern
5. **Register all mutation invalidation rules** in `src/lib/mutationEffects.ts`
6. **Mutations always call `invalidateFor`** in `onSuccess`
7. **Never use optimistic updates** for domain state mutations
8. **Use `staleTime: Infinity`** for static metadata queries
9. **Use Mantine for all UI components** — no plain HTML for dialogs, inputs, buttons, or checkboxes
10. **Never invent CSS class names** for UI elements — they will not exist in the EASI stylesheet
11. **Wrap Mantine component tests** in `MantineTestWrapper` from `src/test/helpers/mantineTestWrapper.tsx`
12. **Audit all call sites** before adding a dialog gate — grep first, update every entry point, not just the one you are coding
