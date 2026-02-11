# Tree-View Filters

## Description

Add two multi-select filters to the navigation tree explorer: "Created by" and "Assigned to domain". Filters only affect the tree-view, never the canvas. The Views section is never filtered. When both filters are active they combine with AND logic (an artifact must match both).

## Requirements

### Backend

#### "Created by" endpoint

No `createdBy` field exists on any artifact. The event store has `actor_id` and `actor_email` on every event (migration 062). A new read model queries the creator (actor on the first event, i.e. `version = 1`) for each tree-relevant aggregate.

- New read model: `ArtifactCreatorReadModel` in `shared/audit`
- Query filters by specific creation event types to avoid leaking internal aggregates:
  ```sql
  SELECT DISTINCT aggregate_id, actor_id
  FROM events
  WHERE tenant_id = $1
    AND version = 1
    AND event_type IN (
      'ApplicationComponentCreated',
      'CapabilityCreated',
      'VendorCreated',
      'InternalTeamCreated',
      'AcquiredEntityCreated'
    )
  ```
- New endpoint: `GET /api/v1/artifact-creators` returns `{ aggregateId, creatorId }`
- Response excludes `creatorEmail` -- the frontend resolves display names from existing user data
- Authorization: `RequirePermission(PermAuditRead)`

#### "Assigned to domain" -- no new backend work

Existing endpoints and frontend data are sufficient. The frontend already loads all capabilities, components, origin entities, and can fetch domain-capability assignments via `useDomainCapabilities(domainId)` per domain. The data volumes (tens to low-hundreds of items) make client-side traversal practical.

### Frontend

#### Filter UI

- Two multi-select dropdowns placed between the `TreeHeader` ("Explorer") and the first section (`ApplicationsSection`) inside `NavigationTreeContent`
- "Created by" dropdown: populated from `GET /api/v1/artifact-creators`, deduplicated by `creatorId`, displayed using existing user display name resolution
- "Assigned to domain" dropdown: populated from `useBusinessDomains()`, plus a synthetic "Unassigned" option
- When no selections are made in a filter, it is inactive (everything passes)
- A "Clear filters" affordance when any filter is active

#### Filter state

- Local component state in `NavigationTree.tsx` (not persisted, not in URL)
- Two pieces of state: `selectedCreatorIds: string[]` and `selectedDomainIds: string[]` (where `"__unassigned__"` is the sentinel for "Unassigned")

#### "Created by" filtering logic

1. Fetch `artifact-creators` once via React Query hook `useArtifactCreators()`
2. Build a `Map<aggregateId, creatorId>` from the response
3. When filter is active: an artifact is visible only if its `creatorId` is in `selectedCreatorIds`
4. Capability hierarchy preservation: if a parent capability is filtered out but a descendant matches, the parent is still shown as a structural node so the tree remains navigable

#### "Assigned to domain" filtering logic

All data needed is already loaded or available via existing hooks. The traversal:

1. For each selected domain, fetch its directly-assigned capability IDs via `useDomainCapabilities(domainId)` (one cached React Query call per domain)
2. Expand to include all descendant capabilities (walk `parentId` tree downward using the already-loaded `capabilities` array)
3. Collect component IDs that realize any included capability (use existing `useCapabilityRealizations(domainId, depth)` per domain)
4. Collect origin entity IDs that are origin of any included component (use `useOriginRelationshipsQuery()` which already returns all relationships)
5. The union of all these IDs across all selected domains forms the visible set

For "Unassigned": an artifact is "unassigned" if it does NOT appear in the visible set computed from ALL domains. This includes:
- Capabilities not assigned (directly or via ancestry) to any domain
- Components that don't realize any domain-assigned capability
- Origin entities not linked to any domain-reachable component

**Data sources already available in `NavigationTree`:**
- `capabilities` (all, with `parentId`)
- `components` (all)
- `acquiredEntities`, `vendors`, `internalTeams` (all)

**Data to add via existing hooks:**
- `useBusinessDomains()` for domain list
- `useDomainCapabilities(domainId)` per domain for capability assignments (cached, small payloads)
- `useCapabilityRealizations(domainId, depth)` per domain for component-capability links
- `useOriginRelationshipsQuery()` for component-to-origin-entity mapping

#### Filtering the sections

Each section receives pre-filtered arrays. Filtering happens in `NavigationTree.tsx` via a `useMemo` that applies both active filters to produce filtered versions of `components`, `capabilities`, `acquiredEntities`, `vendors`, `internalTeams`. `views` is never filtered.

For the capability hierarchy: if a parent capability is filtered out but a descendant matches, the parent must still be shown (as a structural node) so the tree remains navigable. This applies to both filters and must be implemented from Slice 1.

## API Changes

### New Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/artifact-creators` | Returns creator ID for tree-relevant aggregates |

### `GET /api/v1/artifact-creators`

Authorization: `audit:read` (Admin, Architect, Stakeholder)

Response:
```json
{
  "data": [
    { "aggregateId": "uuid", "creatorId": "user-uuid" }
  ],
  "_links": {
    "self": { "href": "/api/v1/artifact-creators", "method": "GET" }
  }
}
```

Only returns creators for components, capabilities, acquired entities, vendors, and internal teams. Does not return email addresses -- the frontend resolves display names from existing user data.

## Implementation

### Slice 1: "Created by" filter (backend + frontend end-to-end)

**Backend:**
- Add `ArtifactCreatorReadModel` in `shared/audit` with event-type-filtered query
- Add handler and route for `GET /api/v1/artifact-creators` with `RequirePermission(PermAuditRead)`
- Add Swagger annotation
- Unit test for read model
- Integration test for endpoint

**Frontend:**
- Add `useArtifactCreators()` React Query hook
- Add `CreatedByFilter` multi-select component
- Add filter state to `NavigationTree`
- Wire filtering logic into section data via `useMemo`
- Implement capability hierarchy preservation (structural parent nodes shown when only descendants match)
- Unit tests for filtering logic (pure function)
- Unit tests for hierarchy preservation logic
- Unit tests for `CreatedByFilter` component

### Slice 2: "Assigned to domain" filter (frontend-only, uses existing backend)

**Frontend:**
- Add `DomainFilter` multi-select component with "Unassigned" option
- Add `useBusinessDomains()`, `useDomainCapabilities()` per domain, `useCapabilityRealizations()` per domain, and `useOriginRelationshipsQuery()` to `NavigationTree`
- Implement domain traversal logic as a pure function: `computeVisibleArtifactsByDomain(selectedDomainIds, domainCapabilities, allCapabilities, realizations, originRelationships)`
- Wire into existing filter pipeline in `NavigationTree`
- AND logic when both filters are active
- "Clear all filters" button
- Empty state messaging when filters produce no results
- Section counts update to reflect filtered totals
- Unit tests for traversal logic (including "Unassigned" computation)
- Unit tests for `DomainFilter` component

## Acceptance Criteria

### "Created by" filter
- [x] Dropdown lists all users who have created at least one artifact, deduplicated
- [x] Selecting one or more users shows only artifacts created by those users in the tree
- [x] Deselecting all users (clearing filter) shows all artifacts
- [x] Canvas continues to show all objects regardless of filter state
- [x] Views section is never filtered
- [x] Capability hierarchy remains navigable (structural parents shown when only descendants match)

**Example:** Given Capability X created by user A, Capability Y created by user B, App Z created by user C, Vendor U created by user B -- when filtering by "Created by: user A, user C" then only Capability X and App Z appear in tree. Canvas still shows everything.

### "Assigned to domain" filter
- [x] Dropdown lists all business domains plus "Unassigned"
- [x] Selecting a domain shows: its directly-assigned capabilities, their descendant capabilities, components realizing those capabilities, and origin entities of those components
- [x] "Unassigned" shows artifacts not reachable from any domain (including orphan components with no realizations and orphan origin entities with no component links)
- [x] Multiple domains can be selected (union of their artifacts)

**Example:** Given capability X (parent of Y), app Z realizes Y, vendor U is origin of Z, X is assigned to domain 1, capability P is not assigned to domain 1 -- when filtering by "Assigned to domain 1" then X, Y, Z, U are shown but P is not.

### Combined filters
- [x] When both filters are active, only artifacts matching BOTH filters are shown (AND logic)
- [x] Clearing one filter still applies the other
- [x] Clearing all filters restores the full tree

### General
- [x] Filters are placed between the Explorer header and the first tree section
- [x] Filter state does not persist across page navigation (local state only)
- [x] Section item counts reflect filtered results
- [x] Capability tree hierarchy remains navigable (structural parents shown when needed)

## Out of Scope

- Filter persistence across sessions or in URL
- Search/type-ahead within filter dropdowns (can be added later if dropdown lists grow large)
- Backend-computed domain visibility (client-side traversal is sufficient for current data volumes)
- Bulk capability-realizations endpoint (existing per-domain hooks are sufficient)
- Loading/error state design beyond React Query defaults

## Checklist

- [x] Specification ready
- [x] Slice 1: "Created by" backend + frontend implemented (includes hierarchy preservation)
- [x] Slice 2: "Assigned to domain" frontend + combined filter logic
- [x] Unit tests implemented and passing (Slice 1 + Slice 2)
- [ ] Integration tests for new endpoint
- [ ] User sign-off
