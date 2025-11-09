# ArchiMate Documentation Links Feature

## Description
Implements HATEOAS (Hypermedia as the Engine of Application State) links in API responses to provide direct navigation simple ArchiMate documentation for components and relations. This enhances the REST API to Level 3 maturity and helps users understand the semantic meaning of ArchiMate elements by linking to a short description of their meaning.


## Requirements

### Application Component Documentation Links
When retrieving Application Components, the response should include:
- A `_links` section following HAL (Hypertext Application Language) conventions
- A `self` link to the component resource
- An `archimate-spec` link pointing to a short text with the Archimate Application Component definition

### Component Relation Documentation Links
When retrieving Component Relations, the response should include:
- A `_links` section following HAL conventions
- A `self` link to the relation resource
- An `archimate-spec` link pointing to a short text with the ArchiMate relationship type definition
- Links for `source` and `target` components

## API Response Format

### GET /api/application-component/{id}

**Response:** 200 OK
```json
{
  "id": "guid",
  "name": "string",
  "description": "string",
  "createdAt": "datetime",
  "_links": {
    "self": {
      "href": "/api/application-component/{id}"
    },
    "archimate-spec": {
      "href": "/api/documentation/application-component/",
      "title": "ArchiMate Application Component"
    },
    "relations-from": {
      "href": "/api/component-relation/from/{id}",
      "title": "Outgoing relations"
    },
    "relations-to": {
      "href": "/api/component-relation/to/{id}",
      "title": "Incoming relations"
    }
  }
}
```

### GET /api/component-relation
**Response:** 200 OK
```json
[
  {
    "id": "guid",
    "sourceComponentId": "guid",
    "targetComponentId": "guid",
    "relationType": "Triggers",
    "name": "string",
    "description": "string",
    "createdAt": "datetime",
    "_links": {
      "self": {
        "href": "/api/component-relation/{id}"
      },
      "archimate-spec": {
        "href": "/api/documentation/component-relation-triggers",
        "title": "ArchiMate Triggering Relationship"
      },
      "source": {
        "href": "/api/application-component/{sourceComponentId}",
        "title": "Source component"
      },
      "target": {
        "href": "/api/application-component/{targetComponentId}",
        "title": "Target component"
      }
    }
  }
]
```

## Implementation Considerations

### Read Model Extensions
- Add `_links` property to read models (ApplicationComponentReadModel, ComponentRelationReadModel)
- `_links` should be a dictionary of link objects
- Each link object contains `href` (required) and optional `title` properties

### Link Generation
- Create a service or helper class to generate HATEOAS links
- Generate self links based on current request context
- Generate related resource links dynamically

### REST Maturity Level 3
This implementation achieves REST Level 3 (Hypermedia Controls) by:
- Including hypermedia links in responses
- Enabling client navigation through API resources
- Providing contextual information about ArchiMate semantics
- Supporting discoverability of related resources

## Checklist
- [ ] Specification ready
- [ ] Read models extended with _links property
- [ ] Projections updated to include links
- [ ] API endpoints return responses with HATEOAS links
- [ ] OpenAPI specification updated
- [ ] Unit tests for link generation
- [ ] Integration tests verify links in responses
- [ ] User sign-off
