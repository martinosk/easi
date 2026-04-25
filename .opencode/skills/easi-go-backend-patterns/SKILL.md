---
name: easi-go-backend-patterns
description: MUST load when writing or reviewing any Go backend code in EASI. Load when adding error handling, event deserializers, field extraction, upcasters, or repository code. 
compatibility: opencode
---

# EASI Go Backend Patterns

---

## Error Handling

- Wrap with `fmt.Errorf("verb noun [id]: %w", err)` at every external boundary (DB, JSON, API). Always `%w`, never `%v`.
- Include relevant entity IDs in the message.
- Use `errors.Is` / `errors.As` for control flow. Declare sentinels with `errors.New` at package level.
- Log **once** at the outermost handler (`slog.ErrorContext`). Intermediate layers wrap and return silently.
- Use `errors.Join` when a secondary error must be preserved alongside a primary one.

```go
// message format
return fmt.Errorf("load capability %s: %w", id, err)

// sentinel
var ErrNotFound = errors.New("not found")
if errors.Is(err, ErrNotFound) { … }

// logging — outermost handler only
slog.ErrorContext(ctx, "handle CreateCapability", "error", err, "tenantID", tenantID)
```

**Never:** `return err` bare at a boundary · `%v` in error wrap · log then return in an intermediate layer · `strings.Contains` on error messages.

---

## context.Context

- First parameter of every exported function that does I/O, DB, or service calls.
- Propagate the same `ctx` — never create `context.Background()` mid-call.
- Use `db.QueryContext` / `db.ExecContext` (not the non-context variants).
- Never store `Context` in a struct field.

---

## Goroutines

- Every goroutine needs a bounded lifetime and a `ctx.Done()` cancellation path.
- Buffer error channels to at least 1 (`make(chan error, 1)`) so the sender never blocks.
- Use `sync.WaitGroup` when waiting for multiple goroutines.

---

## Event Deserialization

Infrastructure lives in `backend/internal/shared/infrastructure/repository/`.

### Prefer `JSONDeserializer[T]`

Use the generic helper when stored JSON maps directly to a domain struct:

```go
var deserializers = repository.NewEventDeserializers(
    map[string]repository.EventDeserializerFunc{
        "CapabilityCreated": repository.JSONDeserializer[events.CapabilityCreated],
        "CapabilityDeleted": repository.JSONDeserializer[events.CapabilityDeleted],
    },
)
```

### Custom deserializers — use typed helpers

Only write a custom deserializer when `JSONDeserializer[T]` cannot be used. Use `GetRequired*` / `GetOptional*` — never bare map type assertions.

```go
// WRONG
id := data["id"].(string)

// CORRECT
id, err := repository.GetRequiredString(data, "id")
if err != nil {
    return nil, err
}
```

Available helpers: `GetRequired/OptionalString`, `Int`, `Bool`, `Float64`, `Time`, `Map`, `StringSlice`.

### Forward compatibility

Unknown event types are **skipped with a warning**, not an error. Never add a catch-all handler or panic on unknown types.

### Schema evolution (upcasters)

Add a new required field to an existing event type by writing an upcaster that injects a safe default. Upcasters run before field extraction.

```go
type addFieldUpcaster struct{}

func (u addFieldUpcaster) EventType() string { return "CapabilityCreated" }
func (u addFieldUpcaster) Upcast(data map[string]interface{}) map[string]interface{} {
    if _, ok := data["newField"]; !ok {
        data["newField"] = "default"
    }
    return data
}

var deserializers = repository.NewEventDeserializers(
    map[string]repository.EventDeserializerFunc{ /* … */ },
    addFieldUpcaster{},
)
```

