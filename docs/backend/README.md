# Backend Conventions

Go backend development patterns and standards.

## Documentation

| File | Purpose |
|------|---------|
| [api-standards.md](api-standards.md) | REST Level 3, HATEOAS, response formats |
| [standard-patterns.md](standard-patterns.md) | Event deserialization, field extraction, upcasters |
| [antipatterns.md](antipatterns.md) | What to avoid |
| [testing.md](testing.md) | Test structure and commands |
| [database.md](database.md) | Migration rules and database conventions |

## Quick Reference

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests
./test-integration.sh
```

### API Response Helpers

```go
sharedAPI.RespondJSON(w, statusCode, resource)           // Single resource
sharedAPI.RespondCollection(w, statusCode, data, links)  // Collection
sharedAPI.RespondPaginated(w, statusCode, data, ...)     // Paginated
sharedAPI.RespondError(w, statusCode, err, message)      // Errors
```

### Core Principles

- REST Level 3 with HATEOAS
- CQRS with Event Sourcing for core domains
- Value objects for all aggregate properties
- No foreign keys - event-driven consistency
