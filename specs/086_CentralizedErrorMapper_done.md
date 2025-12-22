# Spec 086: Centralized Error Mapper

## Overview
Create a centralized error mapping system to consistently translate domain errors to HTTP status codes across all API handlers.

## Problem Statement
Error mapping approaches vary significantly across handlers:

1. **String matching** (fragile):
```go
if err.Error() == "relation not found" {
    sharedAPI.RespondError(w, http.StatusNotFound, err, "Relation not found")
}
```

2. **Typed errors** (better but inconsistent):
```go
if errors.Is(err, repositories.ErrComponentNotFound) {
    sharedAPI.RespondError(w, http.StatusNotFound, err, "Component not found")
}
```

3. **Centralized handler** (best, but only in auth):
```go
func (h *UserHandlers) handleCommandError(w http.ResponseWriter, err error, defaultMessage string) {
    switch {
    case errors.Is(err, repositories.ErrUserAggregateNotFound):
        sharedAPI.RespondError(w, http.StatusNotFound, err, "User not found")
    // ...
    }
}
```

4. **Custom error checking functions** (scattered):
- `isValidationError()`, `isNotFoundError()`, `isParentChangeError()` in different files
- `assignmentErrorMappings` slice pattern in business_domain_handlers

**Additional issue**: Per CLAUDE.md, "API endpoints should NOT duplicate validation logic - they only translate domain exceptions to HTTP status codes"

## Requirements

### R1: Error Category System
Define error categories that map to HTTP status codes:
- Not Found errors → 404
- Validation errors → 400
- Conflict errors (business rule violations) → 409
- Authorization errors → 403
- Internal errors → 500

### R2: Error Registration
Allow bounded contexts to register their domain errors with categories:
- Register at application startup
- Support error hierarchies (e.g., all repository NotFound errors)
- Avoid hardcoding errors in shared layer

### R3: Centralized Mapping Function
Enhance existing `MapErrorToStatusCode()`:
- Check registered error mappings
- Support custom error messages per error type
- Fall back to default status code if not mapped

### R4: Handler Helper
Create helper for common dispatch-and-respond pattern:
```go
func HandleCommandResult(w http.ResponseWriter, err error, successHandler func())
```
- On success: call success handler
- On error: map error and respond appropriately

### R5: Migration of Existing Patterns
- Remove string matching error checks
- Replace custom `isValidationError()` type functions
- Standardize all handlers on centralized mapper

### R6: Validation Error Details
Preserve field-specific validation error details:
- `ValidationError` with field name and message
- Maps to 400 with details in response body

## Affected Files

### Files to Modify
- `backend/internal/shared/api/response.go` (enhance MapErrorToStatusCode)
- `backend/internal/shared/api/domain_error_mapper.go` (expand registration)

### Files to Create
- `backend/internal/shared/api/error_registry.go`
- `backend/internal/shared/api/handler_helpers.go`

### Files to Refactor
- `backend/internal/architecturemodeling/infrastructure/api/component_handlers.go`
- `backend/internal/architecturemodeling/infrastructure/api/relation_handlers.go`
- `backend/internal/architectureviews/infrastructure/api/view_handlers.go`
- `backend/internal/capabilitymapping/infrastructure/api/*.go`
- `backend/internal/auth/infrastructure/api/*.go`

## Checklist
- [x] Specification approved
- [x] Error category system defined
- [x] Error registration mechanism created
- [x] MapErrorToStatusCode enhanced
- [x] Handler helper functions created
- [x] architecturemodeling handlers migrated
- [x] architectureviews handlers migrated
- [x] capabilitymapping handlers migrated
- [x] auth handlers migrated
- [x] String-matching error checks removed
- [x] All existing tests pass
- [x] User sign-off
