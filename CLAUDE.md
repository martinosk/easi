# Code style
Never add comments to code unless specifically instructed to do so by the user.

# Architecture style
- Using the principles of strategic DDD, structure the code by bounded contexts. 
- Bounded contexts must have meaning to the business domain.
- There must never be direct coupling between bounded contexts. Use loosely coupled events if needed.
- Use the principles of tactical DDD when writing code. 
- Keep Domain Model separate of infrastructure concerns
- Use aggregates as transactional boundaries
- If aggregates must link to other aggregates, they do so only by their globally unique ID. Never by reference.
- Use immutable value objects for entities that does not have a lifecycle. This includes the aggregate id.
- **Aggregates must never expose primitive types directly. All properties must be value objects that encapsulate business invariants and domain concepts.**
- Value objects should be immutable records with validation in their constructors.
- Use API first principles. Any functionality is always done via API calls to the backend.

## CQRS with Event Sourcing
Core domains must use CQRS with event sourcing.
### Element Types
| Type | Purpose | Naming Convention | Examples |
|------|---------|------------------|----------|
| **Command** | User actions that change state | Action verbs | Add Item, Submit Order, Cancel Booking |
| **Event** | Past-tense facts about what happened | Past tense | Item Added, Order Submitted, Booking Cancelled |
| **Read Model** | Data views for presentation | Descriptive nouns | Cart Items, Customer Profile, Order History |
| **Screen** | UI representations | UI-focused nouns | Add Item Form, Cart Display, Order Summary |
| **Processor** | Background automation tasks | Process descriptions | Payment Processor, Notification Sender |

### Valid Dependency Patterns
```
Event → ReadModel: Event(OUTBOUND) → ReadModel(INBOUND)
Command → Event: Command(OUTBOUND) → Event(INBOUND)  
Screen → Command: Screen(OUTBOUND) → Command(INBOUND)
ReadModel → Screen: ReadModel(OUTBOUND) → Screen(INBOUND)
```

# API principles

## API Versioning
- **ALL API routes MUST be under `/api/v1/` prefix** (except `/health`)
- Never use hardcoded hosts in swagger - use relative URLs with schemes
- Examples:
  - ✅ `/api/v1/capabilities`
  - ✅ `/api/v1/platform/tenants`
  - ✅ `/api/v1/auth/sessions`
  - ❌ `/capabilities` (missing version)
  - ❌ `/api/platform/v1/tenants` (wrong structure)
  - ❌ `/auth/sessions` (missing version)

## General API Standards
- Create restful API's with maturity level 3
- Document the API endpoints using OpenApi specifications
- Use opaque tokens for paging
- Always use appropriate HTTP status codes:
  - 200 OK: Successful GET, PUT, PATCH requests
  - 201 Created: Successful POST requests that create resources
  - 204 No Content: Successful DELETE requests, PATCH requests that modify without returning data
  - 400 Bad Request: Client-side validation errors, invalid input
  - 401 Unauthorized: Authentication required
  - 403 Forbidden: Authenticated but lacks permission
  - 404 Not Found: Resource does not exist
  - 409 Conflict: Business rule violations, duplicate resources
  - 500 Internal Server Error: Unhandled server errors (should be minimized)
- **Business invariants and validation must ONLY be defined in the domain model (value objects, aggregates)**
- API endpoints should NOT duplicate validation logic - they only translate domain exceptions to HTTP status codes
- Catch domain exceptions (ArgumentException, etc.) and map them to appropriate HTTP status codes (typically 400 Bad Request)
- Never let unhandled exceptions return as 500 errors when they represent client errors

## Response Wrapping Standards
REST Level 3 APIs must follow consistent response structures:

### Single Resource Responses (GET by ID, POST, PUT)
- Return the resource directly at the root level
- Embed `_links` object within the resource for HATEOAS navigation
- Use `sharedAPI.RespondJSON(w, statusCode, resource)`
- Example:
```json
{
  "id": "123",
  "name": "Component A",
  "_links": {
    "self": "/api/v1/components/123",
    "update": "/api/v1/components/123",
    "delete": "/api/v1/components/123"
  }
}
```

### Non-Paginated Collection Responses (GET all)
- Use structured envelope with `data` and `_links`
- Use `sharedAPI.RespondCollection(w, statusCode, data, links)`
- Example:
```json
{
  "data": [
    {"id": "123", "name": "Item 1", "_links": {...}},
    {"id": "456", "name": "Item 2", "_links": {...}}
  ],
  "_links": {
    "self": "/api/v1/items"
  }
}
```

### Paginated Collection Responses (GET with pagination)
- Use structured envelope with `data`, `pagination`, and `_links`
- Use `sharedAPI.RespondPaginated(w, statusCode, data, hasMore, nextCursor, limit, selfLink, baseLink)`
- Example:
```json
{
  "data": [
    {"id": "123", "name": "Item 1", "_links": {...}},
    {"id": "456", "name": "Item 2", "_links": {...}}
  ],
  "pagination": {
    "hasMore": true,
    "limit": 50,
    "cursor": "eyJpZCI6IjQ1NiIsInRzIjoxNjQwMDAwMDAwfQ=="
  },
  "_links": {
    "self": "/api/v1/items?after=xyz&limit=50",
    "next": "/api/v1/items?after=abc&limit=50"
  }
}
```

### Success Responses (201 Created, 204 No Content)
- **201 Created with body**: Return created resource directly with Location header
  - Use: `w.Header().Set("Location", location)` then `sharedAPI.RespondJSON(w, http.StatusCreated, resource)`
- **201 Created without body**: Return 201 with Location header only
  - Use: `w.Header().Set("Location", location)` then `w.WriteHeader(http.StatusCreated)`
- **204 No Content**: No response body (for DELETE, PATCH updates)
  - Use: `w.WriteHeader(http.StatusNoContent)`
- NEVER wrap simple success messages in "data" envelopes

### Error Responses
- Always use consistent error structure via `sharedAPI.RespondError(w, statusCode, err, message)`
- Structure:
```json
{
  "error": "Bad Request",
  "message": "Validation failed",
  "details": {
    "fieldName": "Field-specific error"
  }
}
```

### Implementation Rules
1. NEVER use `sharedAPI.RespondSuccess()` or `sharedAPI.RespondCreated()` - these are deprecated wrappers
2. Single resources: `sharedAPI.RespondJSON(w, statusCode, resource)`
3. Non-paginated collections: `sharedAPI.RespondCollection(w, statusCode, data, links)`
4. Paginated collections: `sharedAPI.RespondPaginated(w, statusCode, data, hasMore, nextCursor, limit, selfLink, baseLink)`
5. Created: `w.Header().Set("Location", ...")` + `sharedAPI.RespondJSON(w, http.StatusCreated, resource)`
6. No content: `w.WriteHeader(http.StatusNoContent)`
7. Errors: `sharedAPI.RespondError(w, statusCode, err, message)`

# Spec Management
- **NEVER modify a spec file with "done" status**
- Never add "future" requirements in a spec. A spec always contain what is to be implemented NOW. Nothing less, nothing more.
- If a done spec needs changes, it must be renamed to "reopened" status
- Keep specs short, precise and descriptive. Avoid prescriptive code examples.
- When reopening a spec:
  - Rename file from `XXX_SpecName_done.md` to `XXX_SpecName_reopened.md`
  - Keep all completed checkmarks for work already done
  - Add new uncompleted checkmarks explaining what additional work is needed
  - Require new user sign-off after changes are complete
- Spec status workflow: `pending` → `ongoing` → `done` (or `done` → `reopened` → `done`)

# Database Migration Management
- **NEVER modify a migration file that has been committed to version control**
- Migrations are immutable once committed - treat them as historical records
- Database structure can be 100% inferred from sequential migration scripts
- Each migration must be a single atomic transaction - no partial application
- If a migration fails midway, the database must roll back to its previous state
- Migration files must be numbered sequentially (001, 002, 003, etc.)
- To fix issues in committed migrations, create a new migration that makes the correction
- Do not add conditional logic checking current schema state - migrations run in order
- Foreign key constraints should not be used in event-sourced read models:
  - Referential integrity is maintained by the domain model and event handlers

# Running Tests

## Frontend Tests
Frontend tests use Vitest with process isolation (takes ~45 seconds). Use these commands:
- `npm test -- --run` - Run all tests once (use 3 minute timeout)
- `npm test -- --run --reporter=dot` - Compact output (dots instead of verbose names)
- `npm test -- --run src/path/to/file.test.ts` - Run specific test file

Note: Always use `--run` flag to avoid watch mode which waits indefinitely.

## Backend Tests
- `go test ./...` - Run all Go tests
- `go test ./internal/path/to/package` - Run specific package tests
