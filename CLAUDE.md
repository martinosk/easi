# Code style
Never add comments to code unless specifically instructed to do so by the user.
Always verify build and tests after modifying files.

# Architecture style
- Using the principles of strategic DDD, structure the code by bounded contexts. 
- Bounded contexts must have meaning to the business domain.
- There must never be direct coupling between bounded contexts. Use loosely coupled events if needed.
- Use the principles of tactical DDD when writing backend code. 
- Keep Domain Model separate of infrastructure concerns.
- Use aggregates as transactional boundaries.
- If aggregates must link to other aggregates, they do so only by their globally unique ID. Never by reference.
- Use immutable value objects with validation in their constructors for entities that does not have a lifecycle. This includes the aggregate id.
- **Aggregates must never expose primitive types directly. All properties must be value objects that encapsulate business invariants and domain concepts.**
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
- **ALL API routes MUST resolve to `/api/v1/` prefix** (except `/health` and `/swagger`)
- Swagger `@Router` annotations must use **relative paths** without `/api/v1/` prefix
  - ✅ `@Router /capabilities [get]` (basePath will be prepended)
  - ❌ `@Router /api/v1/capabilities [get]` (creates double prefix)

### Route Registration
All API routes are registered inside a single `/api/v1` parent route in `router.go`. **Always use relative paths**
**Rule**: All route setup functions receive a router already scoped to `/api/v1`. Always use relative paths like `/auth`, `/users`, `/platform`.

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

### Non-Paginated Collection Responses (GET all)
- Use structured envelope with `data` and `_links`
- Use `sharedAPI.RespondCollection(w, statusCode, data, links)`

### Paginated Collection Responses (GET with pagination)
- Use structured envelope with `data`, `pagination`, and `_links`
- Use `sharedAPI.RespondPaginated(w, statusCode, data, hasMore, nextCursor, limit, selfLink, baseLink)`

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

### Implementation Rules
1. Single resources: `sharedAPI.RespondJSON(w, statusCode, resource)`
2. Non-paginated collections: `sharedAPI.RespondCollection(w, statusCode, data, links)`
3. Paginated collections: `sharedAPI.RespondPaginated(w, statusCode, data, hasMore, nextCursor, limit, selfLink, baseLink)`
4. Created: `w.Header().Set("Location", ...")` + `sharedAPI.RespondJSON(w, http.StatusCreated, resource)`
5. No content: `w.WriteHeader(http.StatusNoContent)`
6. Errors: `sharedAPI.RespondError(w, statusCode, err, message)`

# Spec Management
- **NEVER modify a spec file with "done" status**
- Never add "future" requirements in a spec. A spec always contain what is to be implemented NOW. Nothing less, nothing more.
- Keep specs short, precise and descriptive. Avoid prescriptive code examples.
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
- Foreign key constraints should not be used. Referential integrity is maintained by the domain model and event handlers.

# Running Tests

## Frontend Tests
Frontend tests use Vitest with process isolation (takes ~45 seconds). Use these commands:
- `npm test -- --run` - Run all tests once
Note: Always use `--run` flag to avoid watch mode which waits indefinitely.

## Backend Tests
- `go test ./...` - Run all Go tests
- `go test ./internal/path/to/package` - Run specific package tests
