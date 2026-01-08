# Event Deserializer Error Handling

## Description

Refactor the event deserializer infrastructure to provide proper error handling while maintaining backward and forward compatibility with existing event streams. The current implementation silently swallows errors, making debugging impossible. The proposed changes must respect event sourcing invariants.

## Problem Statement

The current event deserializers have critical issues:

1. **Silent failures**: Type assertion failures return zero values without any indication of error
2. **No distinction between optional and required fields**: All fields use the same extraction pattern
3. **No context in errors**: When failures occur, there's no way to identify which aggregate or event caused them
4. **Forward compatibility broken**: Unknown event types would cause hard failures

## Requirements

### Core Requirements

Given an event store contains valid events
When events are deserialized
Then all events are successfully converted to domain events
And the aggregate is correctly reconstituted

Given an event store contains an event with a missing required field
When events are deserialized
Then an error is returned with the aggregate ID, event type, and field name
And the error is actionable (indicates exactly what's wrong)

Given an event store contains an event with a wrong field type
When events are deserialized
Then an error is returned identifying the expected vs actual type
And the aggregate ID and event type are included in the error

Given an event store contains an unknown event type (forward compatibility)
When events are deserialized
Then the unknown event is skipped with a warning logged
And remaining events are processed normally
And the aggregate is reconstituted from known events

Given an event has optional fields that are missing
When events are deserialized
Then default values are used for optional fields
And no error is returned

Given an event has optional fields with explicit null values
When events are deserialized
Then default values are used (null treated same as missing for optional fields)

### Schema Evolution Requirements

Given an old event version exists in the store
When the event is deserialized
Then upcasters transform it to the current schema BEFORE field extraction
And required field validation runs against the upcasted data

Given a new required field is added to an event type
Then an upcaster MUST be provided that adds the field with a default value
And existing events continue to deserialize successfully

### Error Context Requirements

All deserialization errors MUST include:
- Aggregate ID
- Event type name
- Event sequence number (position in stream)
- Specific field that failed (if applicable)
- Expected type vs actual type (for type errors)

### Logging Requirements

Given an unknown event type is encountered
When it is skipped for forward compatibility
Then a warning is logged with aggregate ID and event type
And metrics are emitted for monitoring

## Design Decisions

### 1. Forward Compatibility via Skip-with-Warning

Unknown event types are skipped rather than causing failures. This enables:
- Rolling deployments where new code writes events old code doesn't understand
- Safe rollbacks without data corruption
- Gradual feature rollouts

**Trade-off**: Aggregates loaded by old code may have incomplete state. This is acceptable because:
- The old code doesn't know about the new behavior anyway
- Once all instances are upgraded, full state is available
- Critical invariants should be enforced at write time, not read time

### 2. Explicit Optional vs Required Field Helpers

Two sets of helper functions:

**Required fields** - return error if missing or wrong type:
- `GetRequiredString(data, key) (string, error)`
- `GetRequiredInt(data, key) (int, error)`
- `GetRequiredFloat64(data, key) (float64, error)`
- `GetRequiredBool(data, key) (bool, error)`
- `GetRequiredTime(data, key) (time.Time, error)`

**Optional fields** - return default value if missing:
- `GetOptionalString(data, key, defaultVal) string`
- `GetOptionalInt(data, key, defaultVal) int`
- `GetOptionalFloat64(data, key, defaultVal) float64`
- `GetOptionalBool(data, key, defaultVal) bool`
- `GetOptionalTime(data, key, defaultVal) time.Time`

### 3. Rich Error Context via Wrapper

Deserializers return field-level errors. The framework wraps them with context:

```go
type DeserializationError struct {
    AggregateID    string
    EventType      string
    SequenceNumber int
    FieldName      string
    Cause          error
}
```

### 4. Upcaster Ordering

Upcasters run BEFORE required field validation. This allows:
- Adding required fields via upcaster defaults
- Renaming fields
- Transforming field types

Order of operations:
1. Load raw event from store
2. Apply upcaster chain (transforms event data)
3. Call deserializer function (validates and extracts fields)
4. Return domain event or error

### 5. Backward Compatibility with Existing Code

Existing helper functions are deprecated but not removed:
- `GetString` → deprecated, use `GetRequiredString` or `GetOptionalString`
- `GetInt` → deprecated, use `GetRequiredInt` or `GetOptionalInt`
- `GetTime` → deprecated, use `GetRequiredTime` or `GetOptionalTime`

The old functions continue to work (return zero values on failure) to allow incremental migration.

## Implementation Details

### Phase 1: Infrastructure (Non-Breaking)

Add new helper functions alongside existing ones:

```
/backend/internal/shared/infrastructure/repository/
├── event_deserializer.go      # Add new helpers, deprecate old ones
├── deserialization_error.go   # New: rich error type
└── field_extractors.go        # New: GetRequired*, GetOptional* functions
```

Update `EventDeserializers.Deserialize()`:
- Change return type to `([]domain.DomainEvent, error)`
- Skip unknown events with warning log (don't fail)
- Wrap deserializer errors with context

### Phase 2: Repository Migration (Per-Repository)

Migrate repositories one at a time:
1. Update deserializer functions to return `(domain.DomainEvent, error)`
2. Replace `GetString` → `GetRequiredString` or `GetOptionalString` based on domain rules
3. Add tests for error cases
4. Verify with production event data sample

### Phase 3: Cleanup

- Remove deprecated helper functions
- Update documentation
- Add CI check that new code doesn't use deprecated functions

### Files to Create/Modify

**New Files:**
- `/backend/internal/shared/infrastructure/repository/field_extractors.go`
- `/backend/internal/shared/infrastructure/repository/field_extractors_test.go`
- `/backend/internal/shared/infrastructure/repository/deserialization_error.go`

**Modified Files:**
- `/backend/internal/shared/infrastructure/repository/event_deserializer.go`
- `/backend/internal/shared/infrastructure/repository/event_sourced_repository.go`
- All repository files in bounded contexts (one PR per context)

### Field Classification per Event Type

Each event type must document which fields are required vs optional:

**Example: ViewCreated**
| Field | Required | Default | Notes |
|-------|----------|---------|-------|
| id | Yes | - | Aggregate identifier |
| name | Yes | - | View name |
| description | Yes | - | Can be empty string |
| createdAt | Yes | - | Event timestamp |
| isPrivate | No | false | Added in v1.5 |
| ownerUserId | No | "" | Required if isPrivate=true |
| ownerEmail | No | "" | Required if isPrivate=true |

### Testing Requirements

**Unit Tests for Field Extractors:**
- Missing required field → error with field name
- Wrong type for required field → error with expected/actual types
- Missing optional field → default value returned
- Null value for optional field → default value returned
- Valid values → correct extraction

**Unit Tests for Deserializers:**
- Valid event → successful deserialization
- Missing required field → error with full context
- Unknown event type → skipped, warning logged
- Upcasted event with new required field → success (upcaster provides default)

**Integration Tests:**
- Load aggregate with mixed event versions
- Load aggregate with unknown future event type
- Verify error messages are actionable

### Migration Safety

Before migrating each repository:
1. Export sample of production events for that aggregate type
2. Run new deserializers against sample data
3. Verify no errors for valid historical data
4. Verify errors are caught for intentionally corrupted test data

### Rollback Plan

If issues are discovered:
1. The old deprecated functions still work
2. Revert the specific repository changes
3. Framework changes (skip unknown events, error wrapping) are backward compatible

## Out of Scope

- Typed event serialization (protobuf, etc.) - separate initiative
- Event store schema validation at write time
- Automatic schema migration tooling
- Event replay/projection error handling

## Checklist

- [x] Specification ready
- [x] Phase 1: Infrastructure implementation
- [x] Phase 1: Unit tests for field extractors
- [ ] Phase 2: Repository migration (architecturemodeling)
- [ ] Phase 2: Repository migration (architectureviews)
- [ ] Phase 2: Repository migration (auth)
- [ ] Phase 2: Repository migration (capabilitymapping)
- [ ] Phase 2: Repository migration (enterprisearchitecture)
- [ ] Phase 2: Repository migration (importing)
- [ ] Phase 2: Repository migration (metamodel)
- [ ] Phase 3: Remove deprecated functions
- [ ] Documentation updated
- [ ] User sign-off
