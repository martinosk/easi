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
- Create restful API's with maturity level 3
- Document the API endpoints using OpenApi specifications
- Use opaque tokens for paging
- Always use appropriate HTTP status codes:
  - 200 OK: Successful GET, PUT, PATCH requests
  - 201 Created: Successful POST requests that create resources
  - 204 No Content: Successful DELETE requests
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

# Spec Management
- **NEVER modify a spec file with "done" status**
- If a done spec needs changes, it must be renamed to "reopened" status
- When reopening a spec:
  - Rename file from `XXX_SpecName_done.md` to `XXX_SpecName_reopened.md`
  - Keep all completed checkmarks for work already done
  - Add new uncompleted checkmarks explaining what additional work is needed
  - Require new user sign-off after changes are complete
- Spec status workflow: `pending` → `ongoing` → `done` (or `done` → `reopened` → `done`)
