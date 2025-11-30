# Reference Documentation Links Feature

## Description
Implements HATEOAS (Hypermedia as the Engine of Application State) links in API responses to provide direct navigation to reference documentation for components and relations. This enhances the REST API to Level 3 maturity and helps users understand the semantic meaning of architectural elements by linking to a short description of their meaning.


## Requirements

### Application Component Documentation Links
When retrieving Application Components, the response should include:
- A `_links` section following HAL (Hypertext Application Language) conventions
- A `self` link to the component resource
- A `reference` link pointing to the Application Component reference documentation

### Component Relation Documentation Links
When retrieving Component Relations, the response should include:
- A `_links` section following HAL conventions
- A `self` link to the relation resource
- A `reference` link pointing to the relationship type reference documentation
- Links for `source` and `target` components

## API Response Format

### GET /api/v1/components/{id}

**Response:** 200 OK
```json
{
  "id": "guid",
  "name": "string",
  "description": "string",
  "createdAt": "datetime",
  "_links": {
    "self": "/api/v1/components/{id}",
    "reference": "/api/v1/reference/components",
    "delete": "/api/v1/components/{id}",
    "collection": "/api/v1/components"
  }
}
```

### GET /api/v1/relations
**Response:** 200 OK
```json
{
  "data": [
    {
      "id": "guid",
      "sourceComponentId": "guid",
      "targetComponentId": "guid",
      "relationType": "Triggers",
      "name": "string",
      "description": "string",
      "createdAt": "datetime",
      "_links": {
        "self": "/api/v1/relations/{id}",
        "reference": "/api/v1/reference/relations/triggering",
        "delete": "/api/v1/relations/{id}",
        "collection": "/api/v1/relations"
      }
    }
  ],
  "_links": {
    "self": "/api/v1/relations"
  }
}
```

## Reference Documentation URLs

| Resource Type | Reference URL |
|---------------|---------------|
| Components | `/api/v1/reference/components` |
| Triggering Relations | `/api/v1/reference/relations/triggering` |
| Serving Relations | `/api/v1/reference/relations/serving` |
| Generic Relations | `/api/v1/reference/relations/generic` |

## Reference Documentation API

The reference documentation endpoints return simple JSON with title and description.

### GET /api/v1/reference/components
**Response:** 200 OK
```json
{
  "title": "Application Component",
  "description": "An Application Component represents a modular, deployable, and replaceable part of a software system that encapsulates its contents and exposes its functionality through a set of interfaces."
}
```

### GET /api/v1/reference/relations/triggering
**Response:** 200 OK
```json
{
  "title": "Triggering Relationship",
  "description": "A Triggering relationship represents a temporal or causal dependency between two elements. The source element initiates or triggers the behavior of the target element."
}
```

### GET /api/v1/reference/relations/serving
**Response:** 200 OK
```json
{
  "title": "Serving Relationship",
  "description": "A Serving relationship represents that an element provides its functionality to another element. The source element serves or provides services to the target element."
}
```

### GET /api/v1/reference/relations/generic
**Response:** 200 OK
```json
{
  "title": "Relationship",
  "description": "A relationship represents a connection or dependency between two architectural elements. Relationships define how elements interact with or depend on each other."
}
```

## Implementation Considerations

### Read Model Extensions
- Add `_links` property to read models (ApplicationComponentReadModel, ComponentRelationReadModel)
- `_links` should be a dictionary of link strings
- Each link provides navigation to related resources

### Link Generation
- Create a service or helper class to generate HATEOAS links
- Generate self links based on current request context
- Generate reference documentation links based on element type

### REST Maturity Level 3
This implementation achieves REST Level 3 (Hypermedia Controls) by:
- Including hypermedia links in responses
- Enabling client navigation through API resources
- Providing contextual reference documentation
- Supporting discoverability of related resources

## Checklist

### Phase 1: HATEOAS Links (Complete)
- [x] Specification ready
- [x] Read models extended with _links property
- [x] Projections updated to include links
- [x] API endpoints return responses with HATEOAS links
- [x] OpenAPI specification updated
- [x] Integration tests verify links in responses

### Phase 2: Reference Documentation API (Complete)
- [x] Create reference handlers in shared/api or new reference package
- [x] Implement GET /api/v1/reference/components endpoint
- [x] Implement GET /api/v1/reference/relations/triggering endpoint
- [x] Implement GET /api/v1/reference/relations/serving endpoint
- [x] Implement GET /api/v1/reference/relations/generic endpoint
- [x] Register routes in router
- [x] Add Swagger documentation for reference endpoints
- [x] Regenerate OpenAPI spec
- [ ] User sign-off
