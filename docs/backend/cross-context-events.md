# Cross-Context Event Integration

## Published Language

Each publishing bounded context exposes a `publishedlanguage/events.go` package with typed string constants for its event types:

```
backend/internal/metamodel/publishedlanguage/events.go
backend/internal/architecturemodeling/publishedlanguage/events.go
backend/internal/capabilitymapping/publishedlanguage/events.go
backend/internal/architectureviews/publishedlanguage/events.go
```

These packages contain **only constants**. No structs, no constructors, no logic.

```go
package publishedlanguage

const (
    ApplicationComponentCreated = "ApplicationComponentCreated"
    ApplicationComponentUpdated = "ApplicationComponentUpdated"
    ApplicationComponentDeleted = "ApplicationComponentDeleted"
)
```

### When to add a constant

When a bounded context publishes an event that another context subscribes to. Intra-context subscriptions do not need constants.

### When to create a new package

When a bounded context becomes a publisher for the first time (i.e., another context needs to subscribe to its events).

## Anti-Corruption Layer (ACL)

Consuming contexts **never import domain event structs** from the publishing context. Instead:

1. Import the **published language constants** for subscription and handler dispatch
2. Define **local deserialization structs** with only the fields the consumer needs

```go
import archPL "easi/backend/internal/architecturemodeling/publishedlanguage"

// Local struct - only the fields this projector needs
type componentDeletedEvent struct {
    ID string `json:"id"`
}

func (p *Projector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
    switch eventType {
    case archPL.ApplicationComponentDeleted:
        var event componentDeletedEvent
        json.Unmarshal(eventData, &event)
        // ...
    }
}
```

### Import alias convention

| Alias | Package |
|-------|---------|
| `mmPL` | `metamodel/publishedlanguage` |
| `archPL` | `architecturemodeling/publishedlanguage` |
| `cmPL` | `capabilitymapping/publishedlanguage` |
| `avPL` | `architectureviews/publishedlanguage` |
