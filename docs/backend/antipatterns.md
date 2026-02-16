# Backend Anti-Patterns

Common mistakes to avoid in backend code.

See also: [standard-patterns.md](standard-patterns.md) for the preferred implementations.

## Error Handling Anti-Patterns

### Never Return Bare Errors at Boundaries

```go
// WRONG - Loses operation context
if err := json.Unmarshal(eventData, &event); err != nil {
    return err
}

// CORRECT - Adds operation + entity context and preserves cause
if err := json.Unmarshal(eventData, &event); err != nil {
    return fmt.Errorf("unmarshal CapabilityDeleted event data: %w", err)
}
```

### Never Break Error Chains with `%v` or String Construction

```go
// WRONG - Cause cannot be matched via errors.Is/errors.As
return fmt.Errorf("load import session %s: %v", sessionID, err)
// WRONG
return errors.New("load import session: " + err.Error())

// CORRECT
return fmt.Errorf("load import session %s: %w", sessionID, err)
```

### Never Use Vague, Context-Free Messages

```go
// WRONG
return fmt.Errorf("operation failed: %w", err)

// CORRECT
return fmt.Errorf("project ApplicationFitScoreSet for fit score %s: %w", event.ID, err)
```

### Never Duplicate Error Context Across Layers

```go
// WRONG - Repeats equivalent context without adding signal
return fmt.Errorf("failed to handle event failed to process event: %w", err)

// CORRECT - Adds one new layer of context
return fmt.Errorf("project OriginLinkDeleted for component %s origin type %s: %w", event.ComponentID, event.OriginType, err)
```

### Never Log One Error and Return Another Unrelated Message

```go
// WRONG - Logs original but returns different contextless error
log.Printf("failed to save import session: %v", err)
return fmt.Errorf("save failed")

// CORRECT - Return the same wrapped error that is logged
wrappedErr := fmt.Errorf("persist completed import session %s: %w", sessionID, err)
log.Printf("failed to save import session %s: %v", sessionID, wrappedErr)
return wrappedErr
```

### Never Rely on String Comparisons for Control Flow

```go
// WRONG
if strings.Contains(err.Error(), "not found") {
    // ...
}

// CORRECT
if errors.Is(err, repositories.ErrImportSessionNotFound) {
    // ...
}
```

## HATEOAS Anti-Patterns

### Never Return Plain Strings for Links

```go
// WRONG - Returns strings without HTTP methods
response := map[string]interface{}{
    "_links": map[string]string{
        "self": "/api/v1/resources/123",
    },
}

// CORRECT - Returns proper Link objects
response := map[string]interface{}{
    "_links": types.Links{
        "self": types.Link{Href: "/api/v1/resources/123", Method: "GET"},
    },
}
```

### Never Hardcode API Paths

```go
// WRONG - Hardcoded path
links["self"] = NewLink("/api/v1/capabilities/" + id, "GET")

// CORRECT - Use builder
links["self"] = NewLink(sharedAPI.BuildResourceLink(
    sharedAPI.ResourcePath("/capabilities"),
    sharedAPI.ResourceID(id),
), "GET")
```

### Never Skip HATEOAS on Custom Responses

```go
// WRONG - Custom response without proper links
response := map[string]interface{}{
    "summary": summary,
    "data":    data,
}

// CORRECT - Include proper links
response := CustomResponse{
    Summary: summary,
    Data:    data,
    Links: types.Links{
        "self": types.Link{Href: "/api/v1/custom-endpoint", Method: "GET"},
    },
}
```
