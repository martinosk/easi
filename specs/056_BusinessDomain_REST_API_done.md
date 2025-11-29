# Business Domain REST API

## Description
REST Level 3 API with HATEOAS for managing business domains and their capability associations.

## Purpose
Enable frontend applications to manage business domains through hypermedia-driven REST endpoints.

## Dependencies
- Spec 053: Business Domain Aggregate
- Spec 054: Business Domain Assignment Aggregate
- Spec 055: Business Domain Read Models

## Endpoints

### GET /api/v1/business-domains
List all business domains.

**Query Parameters:**
- `limit` (int, optional): Page size (default 50, max 100)
- `after` (string, optional): Opaque cursor for pagination

**Response:** 200 OK
```json
{
  "data": [
    {
      "id": "bd-{guid}",
      "name": "Customer Experience",
      "description": "Customer-facing capabilities",
      "capabilityCount": 5,
      "createdAt": "2025-11-28T10:00:00Z",
      "updatedAt": "2025-11-28T14:00:00Z",
      "_links": {
        "self": "/api/v1/business-domains/bd-{guid}",
        "capabilities": "/api/v1/business-domains/bd-{guid}/capabilities",
        "update": "/api/v1/business-domains/bd-{guid}",
        "delete": "/api/v1/business-domains/bd-{guid}"
      }
    }
  ],
  "_links": {
    "self": "/api/v1/business-domains?limit=50",
    "next": "/api/v1/business-domains?after={cursor}&limit=50"
  }
}
```

**Error Responses:**
- 500 Internal Server Error

### POST /api/v1/business-domains
Create a new business domain.

**Request Body:**
```json
{
  "name": "Digital Innovation",
  "description": "Digital transformation capabilities"
}
```

**Response:** 201 Created
```json
Location: /api/v1/business-domains/bd-{guid}
{
  "id": "bd-{guid}",
  "name": "Digital Innovation",
  "description": "Digital transformation capabilities",
  "capabilityCount": 0,
  "createdAt": "2025-11-28T15:00:00Z",
  "_links": {
    "self": "/api/v1/business-domains/bd-{guid}",
    "capabilities": "/api/v1/business-domains/bd-{guid}/capabilities",
    "update": "/api/v1/business-domains/bd-{guid}",
    "delete": "/api/v1/business-domains/bd-{guid}",
    "collection": "/api/v1/business-domains"
  }
}
```

**Error Responses:**
- 400 Bad Request: Invalid input (empty name, exceeds length)
- 409 Conflict: Duplicate domain name

### GET /api/v1/business-domains/{id}
Get a specific business domain.

**Response:** 200 OK
(Same structure as single item in collection)

**Error Responses:**
- 404 Not Found: Domain does not exist

### PUT /api/v1/business-domains/{id}
Update a business domain.

**Request Body:**
```json
{
  "name": "Digital Innovation & Transformation",
  "description": "Updated description"
}
```

**Response:** 200 OK
(Updated domain with same structure)

**Error Responses:**
- 400 Bad Request: Invalid input
- 404 Not Found: Domain does not exist
- 409 Conflict: Name conflicts with another domain

### DELETE /api/v1/business-domains/{id}
Delete a business domain.

**Response:** 204 No Content

**Error Responses:**
- 404 Not Found: Domain does not exist
- 409 Conflict: Domain has associated capabilities

### GET /api/v1/business-domains/{id}/capabilities
List capabilities in a business domain.

**Query Parameters:**
- `limit` (int, optional): Page size (default 50)
- `after` (string, optional): Opaque cursor

**Response:** 200 OK
```json
{
  "data": [
    {
      "id": "cap-{guid}",
      "code": "BIZ-001",
      "name": "Customer Service Management",
      "description": "Managing customer support",
      "level": "L1",
      "assignedAt": "2025-11-20T08:00:00Z",
      "_links": {
        "self": "/api/v1/capabilities/cap-{guid}",
        "children": "/api/v1/capabilities/cap-{guid}/children",
        "businessDomains": "/api/v1/capabilities/cap-{guid}/business-domains",
        "removeFromDomain": "/api/v1/business-domains/{domainId}/capabilities/cap-{guid}"
      }
    }
  ],
  "pagination": {
    "hasMore": true,
    "limit": 50,
    "cursor": "{opaque-cursor}"
  },
  "_links": {
    "self": "/api/v1/business-domains/{id}/capabilities?limit=50",
    "next": "/api/v1/business-domains/{id}/capabilities?after={cursor}&limit=50",
    "domain": "/api/v1/business-domains/{id}"
  }
}
```

**Error Responses:**
- 404 Not Found: Domain does not exist

### POST /api/v1/business-domains/{id}/capabilities
Associate a capability with a business domain.

**Request Body:**
```json
{
  "capabilityId": "cap-{guid}"
}
```

**Response:** 201 Created
```json
Location: /api/v1/business-domains/{domainId}/capabilities/{capabilityId}
{
  "businessDomainId": "bd-{guid}",
  "capabilityId": "cap-{guid}",
  "assignedAt": "2025-11-28T16:00:00Z",
  "_links": {
    "capability": "/api/v1/capabilities/cap-{guid}",
    "businessDomain": "/api/v1/business-domains/bd-{guid}",
    "remove": "/api/v1/business-domains/bd-{guid}/capabilities/cap-{guid}"
  }
}
```

**Error Responses:**
- 400 Bad Request: Invalid capability ID or not L1 capability
- 404 Not Found: Domain or capability does not exist
- 409 Conflict: Capability already in this domain

### DELETE /api/v1/business-domains/{id}/capabilities/{capabilityId}
Remove a capability from a business domain.

**Response:** 204 No Content

**Error Responses:**
- 404 Not Found: Domain, capability, or association does not exist

### GET /api/v1/capabilities/{id}/business-domains
List business domains containing a specific capability.

**Response:** 200 OK
```json
{
  "data": [
    {
      "id": "bd-{guid}",
      "name": "Customer Experience",
      "description": "Customer-facing capabilities",
      "assignedAt": "2025-11-20T08:00:00Z",
      "_links": {
        "self": "/api/v1/business-domains/bd-{guid}",
        "capabilities": "/api/v1/business-domains/bd-{guid}/capabilities",
        "removeCapability": "/api/v1/business-domains/bd-{guid}/capabilities/{capabilityId}"
      }
    }
  ],
  "_links": {
    "self": "/api/v1/capabilities/{id}/business-domains",
    "capability": "/api/v1/capabilities/{id}"
  }
}
```

**Error Responses:**
- 404 Not Found: Capability does not exist

### GET /api/v1/capabilities?filter=unassigned
List L1 capabilities not assigned to any business domain.

**Query Parameters:**
- `filter=unassigned` (required)
- `limit` (int, optional): Page size (default 50)
- `after` (string, optional): Opaque cursor

**Response:** 200 OK
```json
{
  "data": [
    {
      "id": "cap-{guid}",
      "code": "BIZ-002",
      "name": "Risk Management",
      "description": "Enterprise risk assessment",
      "level": "L1",
      "_links": {
        "self": "/api/v1/capabilities/cap-{guid}",
        "children": "/api/v1/capabilities/cap-{guid}/children",
        "businessDomains": "/api/v1/capabilities/cap-{guid}/business-domains",
        "assignToDomain": {
          "href": "/api/v1/business-domains/{domainId}/capabilities",
          "templated": true,
          "method": "POST"
        }
      }
    }
  ],
  "pagination": {
    "hasMore": false,
    "limit": 50
  },
  "_links": {
    "self": "/api/v1/capabilities?filter=unassigned&limit=50",
    "allCapabilities": "/api/v1/capabilities"
  }
}
```

## HATEOAS Link Patterns

### Business Domain Links
- `self`: Link to domain resource
- `capabilities`: Link to capabilities in domain
- `update`: PUT endpoint for updates
- `delete`: DELETE endpoint (only if no capabilities)
- `collection`: Link to all domains

### Capability in Domain Context Links
- `self`: Link to capability resource
- `children`: Link to child capabilities
- `businessDomains`: Link to domains containing capability
- `removeFromDomain`: DELETE link (contextual to current domain)
- `assignToDomain`: POST template for unassigned capabilities

## Error Response Format

All errors use consistent structure:
```json
{
  "error": "Bad Request",
  "message": "Validation failed",
  "details": {
    "name": "Name must not be empty"
  }
}
```

## Implementation Notes
- Use `sharedAPI.RespondCollection()` for non-paginated lists
- Use `sharedAPI.RespondPaginated()` for paginated lists
- Use `sharedAPI.RespondJSON()` for single resources
- Use `sharedAPI.RespondError()` for all errors
- Map domain validation exceptions to 400 Bad Request
- Map business rule violations to 409 Conflict
- Include Location header on 201 Created responses
- Generate HATEOAS links based on current resource state
