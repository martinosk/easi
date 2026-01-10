# Backend API Standards

This document covers HATEOAS-specific patterns. For general API standards (HTTP status codes, response wrapping, API versioning), see `CLAUDE.md`.

## HATEOAS Link Types

Always use the shared types for HATEOAS links:

```go
import "easi/backend/internal/shared/types"

// Link represents a HATEOAS link with href and HTTP method
type Link struct {
    Href   string `json:"href"`
    Method string `json:"method"`
}

// Links is a map of relation names to Link objects
type Links map[string]Link
```

## Creating Links

Use the `HATEOASLinks` helper for consistent link generation:

```go
import sharedAPI "easi/backend/internal/shared/api"

// Initialize once per handler struct
hateoas := sharedAPI.NewHATEOASLinks("")

// Generate links for a resource
links := hateoas.CapabilityLinks(capability.ID, capability.ParentID)
```

For custom endpoints, use the link builder:

```go
import sharedAPI "easi/backend/internal/shared/api"

// Build individual links
selfLink := sharedAPI.BuildResourceLink(
    sharedAPI.ResourcePath("/capabilities"),
    sharedAPI.ResourceID(id),
)

// Or use the fluent builder
links := sharedAPI.NewResourceLinks().
    SelfWithID(sharedAPI.ResourcePath("/capabilities"), sharedAPI.ResourceID(id)).
    Edit(sharedAPI.ResourcePath("/capabilities"), sharedAPI.ResourceID(id)).
    Collection(sharedAPI.ResourcePath("/capabilities")).
    Build()
```

## Standard Link Relations

| Relation | Meaning | When to Include |
|----------|---------|-----------------|
| `self` | Current resource URL | Always |
| `edit` | Update endpoint | User can modify |
| `delete` | Delete endpoint | User can delete |
| `collection` | Parent collection | Always for items |
| `up` | Parent in hierarchy | Has parent resource |
| `describedby` | API documentation | Optional |

## Custom Link Relations

Prefix custom relations with `x-`:

| Relation | Purpose |
|----------|---------|
| `x-children` | Child resources |
| `x-create-link` | Create association |
| `x-remove` | Remove from relationship |
| `x-capability` | Related capability |
| `x-component` | Related component |

## Conditional Links

Include links based on permissions and business rules:

```go
func (h *HATEOASLinks) ViewLinksWithPermissions(viewID string, perms ViewPermissions) Links {
    links := Links{
        "self":       NewLink(fmt.Sprintf("%s/views/%s", h.baseURL, viewID), "GET"),
        "collection": NewLink(fmt.Sprintf("%s/views", h.baseURL), "GET"),
    }

    isOwner := perms.OwnerUserID != nil && *perms.OwnerUserID == perms.CurrentUser
    canEdit := !perms.IsPrivate || isOwner

    if canEdit {
        links["edit"] = NewLink(fmt.Sprintf("%s/views/%s/name", h.baseURL, viewID), "PATCH")
    }

    if canEdit && !perms.IsDefault {
        links["delete"] = NewLink(fmt.Sprintf("%s/views/%s", h.baseURL, viewID), "DELETE")
    }

    return links
}
```

## DTO Structure

DTOs should include a Links field:

```go
type ResourceDTO struct {
    ID          string      `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description,omitempty"`
    CreatedAt   time.Time   `json:"createdAt"`
    Links       types.Links `json:"_links,omitempty"`
}
```

## Error Responses with Links

Errors can include links for recovery actions:

```go
sharedAPI.RespondErrorWithLinks(w, sharedAPI.ErrorWithLinksParams{
    StatusCode: http.StatusConflict,
    Err:        err,
    Message:    "Cannot delete: resource has dependencies",
    Links: map[string]sharedAPI.Link{
        "dependencies": {Href: "/api/v1/resources/123/dependencies", Method: "GET"},
    },
})
```

## Testing HATEOAS Responses

Verify link structure in handler tests:

```go
func TestGetCapability_ReturnsHATEOASLinks(t *testing.T) {
    // ... setup ...

    resp := httptest.NewRecorder()
    handler.GetByID(resp, req)

    var result CapabilityDTO
    json.Unmarshal(resp.Body.Bytes(), &result)

    assert.NotNil(t, result.Links["self"])
    assert.Equal(t, "GET", result.Links["self"].Method)
    assert.Contains(t, result.Links["self"].Href, "/api/v1/capabilities/")

    // Verify conditional links based on permissions
    if userCanEdit {
        assert.NotNil(t, result.Links["edit"])
    } else {
        assert.Nil(t, result.Links["edit"])
    }
}
```
