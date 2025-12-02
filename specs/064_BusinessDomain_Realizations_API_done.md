# Business Domain Realizations API

## Status
**Done**

## User Need
Enterprise architects viewing a business domain grid need to see which applications realise capabilities within that domain. The current implementation passes all visible capability IDs as query parameters, which creates performance issues as domains scale (50-200+ capabilities). The API should accept semantic business intent (domain ID and display preferences) rather than implementation details (lists of capability IDs).

## Dependencies
- Spec 058: Business Domain Visualization UI (grid visualization)
- Spec 063: Show Applications in Capability Grid (frontend display)

---

## User Stories

1. As an architect, I can view applications realising capabilities within a business domain without performance degradation
2. As an architect, I can paginate through large result sets
3. As an architect, I can filter displayed realizations by depth and inheritance (existing frontend controls)

## Success Criteria

- New endpoint accepts domain ID as path parameter
- Backend computes which capabilities belong to the domain
- Returns all realizations for those capabilities (frontend filters for display)
- Cursor-based pagination supports large result sets
- Response includes all fields needed for frontend display
- Frontend hook updated to use new endpoint

---

## Vertical Slices

### Slice 1: New API Endpoint

Create endpoint that fetches realizations for all capabilities assigned to a business domain.

**Backend:**
- [x] Create endpoint `GET /api/v1/business-domains/{domainId}/capability-realizations`
- [x] Accept query parameters: `limit` (integer), `after` (cursor)
- [x] Query capabilities assigned to domain from `domain_capability_assignments`
- [x] Fetch all realizations for those capabilities
- [x] Return paginated response with cursor

### Slice 2: Frontend Integration

Update frontend to use the new semantic endpoint.

**Frontend:**
- [x] Update `useCapabilityRealizations` hook to accept `domainId` instead of `capabilityIds`
- [x] Filter results client-side based on visible capabilities (depth) and inherited toggle
- [x] Handle pagination if needed (lazy load on scroll or explicit "load more")

### Slice 3: Remove Legacy Endpoint

Clean up the previous implementation.

- [x] Remove `GET /api/v1/capability-realizations?capabilityIds=...` endpoint
- [x] Remove unused read model methods

---

## Technical Requirements

### Backend

**New Endpoint:**
```
GET /api/v1/business-domains/{domainId}/capability-realizations
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | int | 100 | Page size (max 500) |
| after | string | - | Pagination cursor |

**Design Rationale:**
- No `maxDepth` parameter: The frontend already knows which capabilities are visible based on its depth slider. It filters the response client-side.
- No `includeInherited` parameter: The frontend owns the toggle state and filters by `origin` field client-side. This avoids extra API calls when toggling.

**Implementation Approach:**
1. Query `domain_capability_assignments` for all capabilities assigned to domain
2. Query `capability_realizations` for those capability IDs
3. Apply cursor-based pagination
4. Return all realizations - frontend filters for display

**Response:** Standard paginated collection following CLAUDE.md conventions.

### Frontend

**Hook Signature Change:**
```typescript
// Before
useCapabilityRealizations(capabilityIds: CapabilityId[], enabled: boolean)

// After
useCapabilityRealizations(domainId: string, enabled: boolean)
```

Frontend filters the returned realizations based on:
- `visibleCapabilityIds` - derived from depth slider state
- `showInherited` toggle - filters by `origin` field

---

## Acceptance Criteria Summary

- [x] `GET /api/v1/business-domains/{domainId}/capability-realizations` endpoint exists
- [x] Endpoint returns all realizations for capabilities in the domain
- [x] Cursor-based pagination works correctly
- [x] Frontend uses new endpoint with domain ID
- [x] Frontend filters by visible capabilities and inherited toggle client-side
- [x] Legacy capability IDs endpoint removed

## Checklist
- [x] Specification ready
- [x] User sign-off on spec
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [x] Final user sign-off
