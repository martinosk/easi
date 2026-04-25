---
name: easi-api-standards
description: MUST load when writing or reviewing any backend HTTP handler in EASI. Load when adding a new endpoint, writing response logic, generating HATEOAS links, checking DTO structure, or writing Swagger annotations.
compatibility: opencode
---

# EASI API Standards

## Overview

REST Level 3 (HATEOAS) throughout. Every response carries `_links` — clients check link presence, never hardcode permission logic. API spec is auto-generated from handler annotations via `swaggo/swag`; `backend/docs/` is generated output, never hand-edited.

## HTTP Status Codes

| Code | Use Case |
|------|----------|
| 200 | Successful GET, PUT, PATCH |
| 201 | POST that creates a resource |
| 204 | DELETE or PATCH with no response body |
| 400 | Validation errors, invalid input |
| 401 | Authentication required |
| 403 | Authenticated but lacks permission |
| 404 | Resource does not exist |
| 409 | Business rule violation, duplicate |
| 500 | Unhandled server error |

## Swagger Annotations

Every exported handler **must** have a `godoc` block — no annotations means no spec entry.

```go
// CreateCapability godoc
// @Summary Create a new business capability
// @Description Creates a new business capability in the capability map
// @Tags capabilities
// @Accept json
// @Produce json
// @Param capability body CreateCapabilityRequest true "Capability data"
// @Success 201 {object} readmodels.CapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /capabilities [post]
```

| Tag | When |
|-----|------|
| `@Summary`, `@Description`, `@Tags`, `@Produce json`, `@Router` | Always |
| `@Accept json` | POST / PUT / PATCH with a request body |
| `@Failure 401`, `@Failure 403`, `@Failure 500` | Every protected endpoint |
| `@Security CookieAuth` | Every endpoint behind auth middleware |

`@Router` paths are **relative to `@BasePath` (`/api/v1`)** — write `/capabilities`, not `/api/v1/capabilities`.

`@Param` format: `// @Param {name} {in: path|query|body|formData} {type} {required} "{description}"`

### Regenerating the spec

```bash
cd backend && make swagger
```

Generates `backend/docs/docs.go` + `backend/docs/swagger.json`, then copies to `frontend/openapi.json`. **Commit all three.** Never edit them by hand — fix annotations in the handler source instead.

## Response Helpers

Use `sharedAPI` — never raw `json.Marshal` + `w.Write`:

| Situation | Call |
|-----------|------|
| Single resource | `sharedAPI.RespondJSON(w, code, resource)` |
| Collection (non-paginated) | `sharedAPI.RespondCollection(w, code, data, links)` |
| Collection (paginated) | `sharedAPI.RespondPaginated(w, code, data, hasMore, cursor, limit, self, base)` |
| 201 Created | Set `Location` header, then `sharedAPI.RespondJSON(w, 201, resource)` |
| 204 No Content | `w.WriteHeader(http.StatusNoContent)` |
| Error | `sharedAPI.RespondError(w, code, err, message)` |
| Error + recovery links | `sharedAPI.RespondErrorWithLinks(w, ErrorWithLinksParams{...})` |

## HATEOAS Links

Use the link builder — never construct URL strings by hand:

```go
links := sharedAPI.NewResourceLinks().
    SelfWithID(sharedAPI.ResourcePath("/capabilities"), sharedAPI.ResourceID(id)).
    Edit(sharedAPI.ResourcePath("/capabilities"), sharedAPI.ResourceID(id)).
    Collection(sharedAPI.ResourcePath("/capabilities")).
    Build()
```

Gate every `edit` and `delete` link on the actor's permissions — never include them unconditionally.

### Standard relations

| Relation | When |
|----------|------|
| `self` | Always |
| `collection` | Always for items in a list |
| `edit` | Actor can modify |
| `delete` | Actor can delete |
| `up` | Resource has a parent |

Custom relations use the `x-` prefix (`x-children`, `x-remove`, `x-create-link`, etc.).

### DTO structure

Every DTO needs `Links types.Links `json:"_links,omitempty"```. Every handler test must assert the `self` link's `Href` and `Method`, and verify conditional links (`edit`, `delete`) are present or absent based on the test actor's permissions.

## Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| Exported handler without `godoc` block | Add full annotation block |
| `@Router` path starts with `/api/v1` | Remove the prefix — it duplicates `@BasePath` |
| Missing `@Failure 401/403` on protected endpoint | Always document auth failures |
| Editing `backend/docs/` by hand | Fix annotations in source, run `make swagger` |
| Committing handler changes without `make swagger` | Run it; commit the generated files |
| Plain string for `_links` value | Use `types.Link{Href, Method}` |
| Hardcoded URL path in link | Use `sharedAPI.ResourcePath` / `sharedAPI.ResourceID` |
| `_links` absent on custom response | Add `types.Links` field to every response struct |
| Links included regardless of permission | Gate on actor permission check |

## Guidelines

1. Every exported handler has a `godoc` annotation block — minimum: `@Summary`, `@Description`, `@Tags`, `@Produce`, `@Router`
2. `@Security CookieAuth` on every protected endpoint; `@Failure 401/403/500` always documented
3. Run `make swagger` after any annotation change; commit `docs.go` and `frontend/openapi.json`
4. Never edit `backend/docs/` — fix source annotations and regenerate
5. Use `sharedAPI.Respond*` helpers — no raw JSON writes
6. Every DTO has `_links`; every handler test asserts link presence, href, and method
7. Gate `edit`/`delete` links on actor permissions — never include unconditionally
8. Custom link relations use the `x-` prefix
