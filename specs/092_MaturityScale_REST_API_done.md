# Maturity Scale REST API

## Description
REST Level 3 API with HATEOAS for managing maturity scale configuration within the MetaModel context.

## Purpose
Enable frontend applications and administrators to view and manage tenant-specific maturity scale configuration through a consistent, discoverable API.

## Dependencies
- Spec 090: MetaModel Bounded Context
- Spec 091: Maturity Scale Configuration Aggregate

## Base Path
`/api/v1/meta-model`

## Endpoints

### GET /api/v1/meta-model/maturity-scale

Get the current maturity scale configuration for the authenticated tenant.

**Response**: 200 OK
```json
{
  "id": "config-uuid-123",
  "tenantId": "tenant-123",
  "sections": [
    {
      "order": 1,
      "name": "Genesis",
      "minValue": 0,
      "maxValue": 24
    },
    {
      "order": 2,
      "name": "Custom Built",
      "minValue": 25,
      "maxValue": 49
    },
    {
      "order": 3,
      "name": "Product",
      "minValue": 50,
      "maxValue": 74
    },
    {
      "order": 4,
      "name": "Commodity",
      "minValue": 75,
      "maxValue": 99
    }
  ],
  "version": 1,
  "isDefault": true,
  "modifiedAt": "2025-12-26T10:00:00Z",
  "modifiedBy": "system",
  "_links": {
    "self": "/api/v1/meta-model/maturity-scale",
    "update": "/api/v1/meta-model/maturity-scale",
    "reset": "/api/v1/meta-model/maturity-scale/reset",
    "maturityLevels": "/api/v1/capabilities/metadata/maturity-levels"
  }
}
```

**Error Responses**:
- 404 Not Found: Configuration not found (tenant not initialized)

### PUT /api/v1/meta-model/maturity-scale

Update the maturity scale configuration.

**Request Body**:
```json
{
  "sections": [
    {
      "order": 1,
      "name": "Genesis",
      "minValue": 0,
      "maxValue": 30
    },
    {
      "order": 2,
      "name": "Custom Built",
      "minValue": 31,
      "maxValue": 50
    },
    {
      "order": 3,
      "name": "Product",
      "minValue": 51,
      "maxValue": 75
    },
    {
      "order": 4,
      "name": "Commodity",
      "minValue": 76,
      "maxValue": 99
    }
  ],
  "version": 1
}
```

**Response**: 200 OK (same structure as GET)

**Error Responses**:
- 400 Bad Request: Validation errors
  - Section names empty or too long
  - Values not in 0-99 range
  - Sections not contiguous
  - Invalid number of sections
- 409 Conflict: Version mismatch (optimistic locking)
- 404 Not Found: Configuration not found

**400 Error Response Example**:
```json
{
  "error": "Bad Request",
  "message": "Maturity scale validation failed",
  "details": [
    {
      "field": "sections[1].minValue",
      "message": "Section must start at 31 (previous section ended at 30)"
    }
  ]
}
```

**409 Error Response Example**:
```json
{
  "error": "Conflict",
  "message": "Configuration was modified by another user. Please refresh and try again.",
  "currentVersion": 3
}
```

### POST /api/v1/meta-model/maturity-scale/reset

Reset the maturity scale to default configuration.

**Request Body**: Empty or `{}`

**Response**: 200 OK (returns the reset configuration, same structure as GET)

**Behavior**:
- Restores sections to: Genesis (0-24), Custom Built (25-49), Product (50-74), Commodity (75-99)
- Increments version
- Sets isDefault = true
- Updates modifiedAt and modifiedBy

**Error Responses**:
- 404 Not Found: Configuration not found

## Backward Compatibility: Maturity Levels Endpoint

The existing `/api/v1/capabilities/metadata/maturity-levels` endpoint will be updated to read from the MetaModel configuration.

### GET /api/v1/capabilities/metadata/maturity-levels

**Current Response** (hardcoded):
```json
{
  "data": [
    { "value": "Genesis", "numericValue": 1 },
    { "value": "Custom Build", "numericValue": 2 },
    { "value": "Product", "numericValue": 3 },
    { "value": "Commodity", "numericValue": 4 }
  ],
  "_links": { "self": "/api/v1/capabilities/metadata/maturity-levels" }
}
```

**New Response** (from configuration):
```json
{
  "data": [
    {
      "value": "Genesis",
      "numericValue": 1,
      "range": { "min": 0, "max": 24 }
    },
    {
      "value": "Custom Built",
      "numericValue": 2,
      "range": { "min": 25, "max": 49 }
    },
    {
      "value": "Product",
      "numericValue": 3,
      "range": { "min": 50, "max": 74 }
    },
    {
      "value": "Commodity",
      "numericValue": 4,
      "range": { "min": 75, "max": 99 }
    }
  ],
  "_links": {
    "self": "/api/v1/capabilities/metadata/maturity-levels",
    "configuration": "/api/v1/meta-model/maturity-scale"
  }
}
```

**Backward Compatibility**:
- `value` field remains (now from configuration)
- `numericValue` field remains (section order 1-4)
- `range` field is additive (new)
- `_links.configuration` is additive (new)

## HATEOAS Links

### Maturity Scale Configuration Links
- `self`: Link to current resource
- `update`: PUT endpoint (included if user has permission)
- `reset`: Reset action (included if configuration is not default)
- `maturityLevels`: Link to metadata endpoint that uses this configuration

## Caching

The configuration endpoint should have appropriate cache headers:
- `Cache-Control: private, max-age=300` (5 minutes, tenant-specific)
- `ETag` based on version number

The maturity-levels endpoint remains cacheable:
- `Cache-Control: public, max-age=3600` (1 hour - but must vary by tenant)
- Consider `Vary: X-Tenant-ID` or tenant-aware cache keys

## Implementation Notes

### Route Registration
Register routes under `/meta-model` subrouter with standard REST methods.

### Response Handling
- Use shared API response helpers for consistency
- Map domain validation errors to 400 Bad Request
- Map version conflicts to 409 Conflict

### Cross-Context Integration
The maturity-levels endpoint in CapabilityMapping context must query MetaModel context.

**Required Approach:** REST API call via Anti-Corruption Layer (see Spec 094).

Direct database queries between bounded contexts are NOT allowed as they violate context isolation. The maturity-levels handler must use the `MaturityScaleGateway` interface to call the MetaModel REST API with caching.

## Checklist
- [x] Specification ready
- [x] Route registration implemented
- [x] GET handler implemented
- [x] PUT handler implemented with validation
- [x] POST /reset handler implemented
- [ ] Maturity-levels endpoint updated to read from configuration (see Spec 094)
- [x] HATEOAS links implemented
- [x] Error responses match specification
- [x] Cache headers implemented
- [x] OpenAPI/Swagger documentation updated
- [x] Unit tests for handlers
- [x] Integration tests for API
- [x] User sign-off
