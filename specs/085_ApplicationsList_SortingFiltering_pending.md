# Applications List Sorting and Filtering

## Description
Enhance the applications list with alphabetical sorting (backend) and freetext search filtering (frontend client-side).

## Requirements

### Alphabetical Sorting (Backend)
- Applications list sorted alphabetically by name (A-Z)
- Case-insensitive sorting
- Replace `created_at DESC` ordering with `LOWER(name) ASC, id ASC`

### Freetext Search (Frontend)
- Client-side filtering of loaded applications
- Matches against name and description (case-insensitive, partial match)
- Instant filtering without API calls

## API Changes

### GET /api/v1/components
- Sort order changes from `created_at DESC` to `LOWER(name) ASC`
- Cursor encoding changes from timestamp-based to name-based

## Implementation

### Backend
1. Update `ApplicationComponentReadModel.GetAllPaginated`:
   - Order by `LOWER(name) ASC, id ASC`
   - Cursor encodes `{name, id}` instead of `{timestamp, id}`
   - Cursor comparison: `(LOWER(name) > cursor_name) OR (LOWER(name) = cursor_name AND id > cursor_id)`

### Frontend
1. Add search input to applications list
2. Filter displayed results in-memory using `Array.filter()`

## Acceptance Criteria
- [x] Applications list sorted alphabetically by name
- [x] Frontend search input filters displayed applications
- [x] Search matches name and description (case-insensitive)
- [x] Pagination works with new sort order

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [ ] User sign-off
