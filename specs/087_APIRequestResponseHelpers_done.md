# Spec 087: API Request/Response Helpers

## Overview
Create shared helpers for common request parsing and response generation patterns to eliminate boilerplate and improve consistency across API handlers.

## Problem Statement

### Request Parsing Duplication
Identical JSON decode pattern repeated 40+ times:
```go
var req CreateSomethingRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
    return
}
```

### Response Pattern Duplication
Each handler creates its own helper for common patterns:
```go
func (h *SomeHandlers) respondWithEntity(w http.ResponseWriter, r *http.Request, id string, statusCode int) {
    entity, err := h.readModel.GetByID(r.Context(), id)
    if err != nil {
        sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve entity")
        return
    }
    entity.Links = h.hateoas.EntityLinks(id)
    sharedAPI.RespondJSON(w, statusCode, entity)
}
```

### HATEOAS Inconsistencies
- Links built inline vs helper methods vs sometimes forgotten
- **BUG**: `capability_handlers.go:112` uses `/api/capabilities/` instead of `/api/v1/capabilities/`

### Pagination Cursor Duplication
- Custom `decodePaginationCursor()` methods duplicate `PaginationHelper`
- Different cursor strategies without clear reasoning

## Requirements

### R1: Generic Request Decoder
Create type-safe request decoder:
- Generic function: `DecodeRequest[T](r *http.Request) (T, error)`
- Returns decoded request or appropriate error
- Handles JSON decode errors with user-friendly messages

### R2: Request Decoder with Validation
Optional validation after decode:
- `DecodeAndValidate[T](r *http.Request, validator func(T) error) (T, error)`
- Separates structural validation (JSON) from domain validation
- Note: Per CLAUDE.md, domain validation should be in domain layer, not API

### R3: Path Parameter Helpers
Simplify path parameter extraction:
- `GetPathParam(r, "id") string`
- `GetPathParamAsUUID(r, "id") (string, error)` - with UUID validation

### R4: Standard Response Helpers
Enhance existing response helpers for common patterns:
- `RespondWithEntity(w, statusCode, entity, linkGenerator)` - single resource with links
- `RespondCreated(w, location, entity)` - 201 with Location header and body
- `RespondDeleted(w)` - 204 No Content

### R5: HATEOAS Link Builder
Create consistent link builder with API version prefix:
- Automatically includes `/api/v1/` prefix
- Methods for common patterns: `Self()`, `Update()`, `Delete()`, `Collection()`
- Type-safe for each resource type

### R6: Pagination Response Helper
Simplify paginated responses:
- Integrate with existing `PaginationHelper`
- Standard cursor encoding/decoding
- Consistent response structure

### R7: Fix API Version Bug
Fix the missing `/api/v1/` prefix in capability_handlers.go:112

## Affected Files

### Files to Create
- `backend/internal/shared/api/request_helpers.go`
- `backend/internal/shared/api/link_builder.go`

### Files to Modify
- `backend/internal/shared/api/response.go` (add response helpers)
- `backend/internal/shared/api/hateoas.go` (enhance with version prefix)
- `backend/internal/capabilitymapping/infrastructure/api/capability_handlers.go` (fix bug)

### Files to Refactor
All API handlers across bounded contexts to use new helpers

## Bug Fix
**File**: `backend/internal/capabilitymapping/infrastructure/api/capability_handlers.go`
**Line**: 112
**Current**: `location := fmt.Sprintf("/api/capabilities/%s", capabilityID)`
**Fixed**: `location := fmt.Sprintf("/api/v1/capabilities/%s", capabilityID)`

## Checklist
- [x] Specification approved
- [x] Generic request decoder created with tests
- [x] Path parameter helpers created
- [x] Response helpers enhanced
- [x] HATEOAS link builder created with version prefix
- [x] API version bug fixed in capability_handlers.go
- [x] architecturemodeling handlers refactored
- [x] architectureviews handlers refactored
- [x] capabilitymapping handlers refactored
- [x] auth handlers refactored
- [x] All existing tests pass
- [ ] User sign-off
