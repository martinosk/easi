# Backend Standard Patterns

## Event Deserialization

Event deserializers convert stored event data into domain events. Located in `backend/internal/shared/infrastructure/repository/`.

### Field Extraction

Use explicit required/optional helpers to extract fields from event data:

**Required fields** - Return error if missing or wrong type:
- `GetRequired*(data, key) (type, error)`

**Optional fields** - Return default value if missing, error only on type mismatch:
- `GetOptional*(data, key, defaultVal) (type, error)`

### Error Types

**FieldError** - Field-level extraction failures

**DeserializationError** - Wraps field errors with context:

### Forward Compatibility

Unknown event types are skipped with a warning logged.

### Schema Evolution

When adding new required fields to existing event types:
1. Create an upcaster that adds the field with a default value
2. Upcasters run BEFORE field extraction
3. Existing events continue to deserialize successfully

Order of operations:
1. Load raw event from store
2. Apply upcaster chain (transforms event data)
3. Call deserializer function (validates and extracts fields)
4. Return domain event or error