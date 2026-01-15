# Backend Anti-Patterns

Common mistakes to avoid in backend code.

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
