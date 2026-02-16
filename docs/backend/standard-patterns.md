# Backend Standard Patterns

See also: [antipatterns.md](antipatterns.md) for examples of what not to do.

## Go Error Handling

Use consistent, inspectable errors so incidents are debuggable and callers can reliably use `errors.Is` and `errors.As`.

### Wrap at Boundary Crossings

Wrap errors when crossing technical boundaries such as:
- event deserialization (`json.Unmarshal`, `json.Marshal`)
- repository/read-model DB calls
- command/query dispatch
- external gateway/API calls

Use operation-first, lower-case messages and `%w`:

```go
if err := json.Unmarshal(eventData, &event); err != nil {
	return fmt.Errorf("unmarshal CapabilityAssignedToDomain event data: %w", err)
}

if err := readModel.Insert(ctx, dto); err != nil {
	return fmt.Errorf("project CapabilityAssignedToDomain assignment insert for assignment %s: %w", event.ID, err)
}
```

### Include Identifying Context

Include IDs and key runtime context when available:
- aggregate ID / event type
- command ID / session ID
- tenant ID (where relevant)
- domain entity IDs (capability ID, component ID, etc.)

Good pattern:

```go
return fmt.Errorf("load import session %s: %w", sessionID, err)
```

### Preserve Error Chains

Always preserve the cause with `%w` so callers can branch safely:

```go
if err := handler.Handle(ctx, cmd); err != nil {
	return fmt.Errorf("dispatch confirm import command for session %s: %w", sessionID, err)
}
```

Caller-side checks remain valid:

```go
if errors.Is(err, repositories.ErrImportSessionNotFound) {
	// handle not found
}
```

### Keep Messages Additive, Not Repetitive

Add one new layer of context per boundary; avoid duplicating equivalent context at the same layer.

Preferred:

```go
return fmt.Errorf("resolve component name for fit score %s component %s: %w", event.ID, event.ComponentID, err)
```

### Logging and Returning

If you both log and return, log the wrapped error that you return, and avoid multiple logs for the same failure path.

```go
wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
log.Printf("failed to marshal event data: %v", wrappedErr)
return wrappedErr
```

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