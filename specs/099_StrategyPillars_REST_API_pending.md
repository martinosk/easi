# Strategy Pillars REST API

## Description
REST Level 3 API with HATEOAS for managing strategy pillar configuration (MetaModel context) and domain-scoped capability-to-pillar alignments (CapabilityMapping context).

## Purpose
Enable frontend applications and administrators to manage tenant-specific strategy pillars and align capabilities to multiple pillars with strategic importance ratings scoped per business domain.

## Dependencies
- Spec 098: Strategy Pillars Meta Model

## Part 1: Strategy Pillar Configuration API (MetaModel Context)

### Base Path
`/api/v1/meta-model/strategy-pillars`

### GET /api/v1/meta-model/strategy-pillars

Get all strategy pillars for the authenticated tenant.

**Response**: 200 OK
```json
{
  "data": [
    {
      "id": "pillar-uuid-1",
      "name": "Always On",
      "description": "Core capabilities that must always be operational",
      "displayOrder": 1,
      "isActive": true
    },
    {
      "id": "pillar-uuid-2",
      "name": "Grow",
      "description": "Capabilities driving business growth",
      "displayOrder": 2,
      "isActive": true
    },
    {
      "id": "pillar-uuid-3",
      "name": "Transform",
      "description": "Capabilities enabling digital transformation",
      "displayOrder": 3,
      "isActive": true
    }
  ],
  "_links": {
    "self": "/api/v1/meta-model/strategy-pillars",
    "create": "/api/v1/meta-model/strategy-pillars"
  }
}
```

**Query Parameters**:
- `includeInactive=true`: Include soft-deleted pillars (default: false)

**Error Responses**:
- 401 Unauthorized: Not authenticated
- 404 Not Found: Configuration not found (tenant not initialized)

### GET /api/v1/meta-model/strategy-pillars/{pillarId}

Get a specific strategy pillar.

**Response**: 200 OK
```json
{
  "id": "pillar-uuid-1",
  "name": "Always On",
  "description": "Core capabilities that must always be operational",
  "displayOrder": 1,
  "isActive": true,
  "_links": {
    "self": "/api/v1/meta-model/strategy-pillars/pillar-uuid-1",
    "update": "/api/v1/meta-model/strategy-pillars/pillar-uuid-1",
    "delete": "/api/v1/meta-model/strategy-pillars/pillar-uuid-1",
    "collection": "/api/v1/meta-model/strategy-pillars"
  }
}
```

**Error Responses**:
- 404 Not Found: Pillar not found

### POST /api/v1/meta-model/strategy-pillars

Create a new strategy pillar.

**Request Body**:
```json
{
  "name": "Innovation",
  "description": "Capabilities driving innovation initiatives",
  "displayOrder": 4
}
```

**Response**: 201 Created
```json
{
  "id": "pillar-uuid-4",
  "name": "Innovation",
  "description": "Capabilities driving innovation initiatives",
  "displayOrder": 4,
  "isActive": true,
  "_links": {
    "self": "/api/v1/meta-model/strategy-pillars/pillar-uuid-4",
    "update": "/api/v1/meta-model/strategy-pillars/pillar-uuid-4",
    "delete": "/api/v1/meta-model/strategy-pillars/pillar-uuid-4",
    "collection": "/api/v1/meta-model/strategy-pillars"
  }
}
```

**Headers**:
- `Location: /api/v1/meta-model/strategy-pillars/pillar-uuid-4`

**Error Responses**:
- 400 Bad Request: Validation errors
  - Name empty or exceeds 100 characters
  - Description exceeds 500 characters
  - Display order not positive
  - Maximum pillars reached (20)
- 409 Conflict: Pillar name already exists

### PUT /api/v1/meta-model/strategy-pillars/{pillarId}

Update a strategy pillar.

**Request Body**:
```json
{
  "name": "Innovation & R&D",
  "description": "Updated description",
  "displayOrder": 4,
  "version": 1
}
```

**Response**: 200 OK (same structure as GET)

**Error Responses**:
- 400 Bad Request: Validation errors
- 404 Not Found: Pillar not found
- 409 Conflict: Version mismatch or name already exists

### DELETE /api/v1/meta-model/strategy-pillars/{pillarId}

Remove a strategy pillar (soft delete).

**Response**: 204 No Content

**Behavior**:
- Marks pillar as inactive (isActive = false)
- Existing alignments are preserved for historical data
- New alignments to this pillar are blocked

**Error Responses**:
- 400 Bad Request: Cannot delete last active pillar
- 404 Not Found: Pillar not found

---

## Part 2: Domain-Scoped Strategy Alignment API (CapabilityMapping Context)

### Base Path
`/api/v1/business-domains/{domainId}/capabilities/{capabilityId}/strategy-alignments`

This URL structure makes the domain scope explicit in the resource hierarchy.

### GET /api/v1/business-domains/{domainId}/capabilities/{capabilityId}/strategy-alignments

Get all strategy pillar alignments for a capability within a specific domain.

**Response**: 200 OK
```json
{
  "data": [
    {
      "id": "dom-strat-uuid-1",
      "pillarId": "pillar-uuid-1",
      "pillarName": "Always On",
      "strategicImportance": 5,
      "alignedAt": "2025-12-26T10:00:00Z"
    },
    {
      "id": "dom-strat-uuid-2",
      "pillarId": "pillar-uuid-2",
      "pillarName": "Grow",
      "strategicImportance": 3,
      "alignedAt": "2025-12-26T11:00:00Z"
    }
  ],
  "_links": {
    "self": "/api/v1/business-domains/domain-uuid/capabilities/cap-uuid/strategy-alignments",
    "businessDomain": "/api/v1/business-domains/domain-uuid",
    "capability": "/api/v1/capabilities/cap-uuid",
    "create": "/api/v1/business-domains/domain-uuid/capabilities/cap-uuid/strategy-alignments",
    "availablePillars": "/api/v1/meta-model/strategy-pillars"
  }
}
```

**Error Responses**:
- 404 Not Found: Business domain or capability not found

### POST /api/v1/business-domains/{domainId}/capabilities/{capabilityId}/strategy-alignments

Align a capability to a strategy pillar within a domain context.

**Request Body**:
```json
{
  "pillarId": "pillar-uuid-3",
  "strategicImportance": 4
}
```

**Response**: 201 Created
```json
{
  "id": "dom-strat-uuid-3",
  "pillarId": "pillar-uuid-3",
  "pillarName": "Transform",
  "strategicImportance": 4,
  "alignedAt": "2025-12-26T12:00:00Z",
  "_links": {
    "self": "/api/v1/business-domains/domain-uuid/capabilities/cap-uuid/strategy-alignments/dom-strat-uuid-3",
    "update": "/api/v1/business-domains/domain-uuid/capabilities/cap-uuid/strategy-alignments/dom-strat-uuid-3",
    "delete": "/api/v1/business-domains/domain-uuid/capabilities/cap-uuid/strategy-alignments/dom-strat-uuid-3",
    "businessDomain": "/api/v1/business-domains/domain-uuid",
    "capability": "/api/v1/capabilities/cap-uuid",
    "pillar": "/api/v1/meta-model/strategy-pillars/pillar-uuid-3"
  }
}
```

**Headers**:
- `Location: /api/v1/business-domains/domain-uuid/capabilities/cap-uuid/strategy-alignments/dom-strat-uuid-3`

**Error Responses**:
- 400 Bad Request: Validation errors
  - Strategic importance not in range 1-5
  - Pillar is inactive
- 404 Not Found: Business domain, capability, or pillar not found
- 409 Conflict: Capability already aligned to this pillar in this domain

### PUT /api/v1/business-domains/{domainId}/capabilities/{capabilityId}/strategy-alignments/{alignmentId}

Update strategic importance for an alignment.

**Request Body**:
```json
{
  "strategicImportance": 5
}
```

**Response**: 200 OK (same structure as POST response)

**Error Responses**:
- 400 Bad Request: Invalid strategic importance
- 404 Not Found: Alignment not found

### DELETE /api/v1/business-domains/{domainId}/capabilities/{capabilityId}/strategy-alignments/{alignmentId}

Remove an alignment.

**Response**: 204 No Content

**Error Responses**:
- 404 Not Found: Alignment not found

---

## Part 3: Domain Portfolio Query API

### GET /api/v1/business-domains/{domainId}/strategy-alignments

Get all strategy alignments for a business domain (domain portfolio view).

**Query Parameters**:
- `pillarId`: Filter by specific pillar
- `minImportance`: Filter by minimum strategic importance (1-5)
- `limit`: Page size (default: 50, max: 100)
- `cursor`: Pagination cursor

**Response**: 200 OK
```json
{
  "data": [
    {
      "id": "dom-strat-uuid-1",
      "capabilityId": "cap-uuid-1",
      "capabilityName": "Customer Onboarding",
      "capabilityLevel": "L1",
      "pillarId": "pillar-uuid-1",
      "pillarName": "Always On",
      "strategicImportance": 5,
      "alignedAt": "2025-12-26T10:00:00Z"
    },
    {
      "id": "dom-strat-uuid-2",
      "capabilityId": "cap-uuid-2",
      "capabilityName": "Payment Processing",
      "capabilityLevel": "L1",
      "pillarId": "pillar-uuid-2",
      "pillarName": "Transform",
      "strategicImportance": 4,
      "alignedAt": "2025-12-26T11:00:00Z"
    }
  ],
  "pagination": {
    "hasMore": true,
    "nextCursor": "eyJpZCI6ImRvbS1zdHJhdC11dWlkLTUwIn0",
    "limit": 50
  },
  "_links": {
    "self": "/api/v1/business-domains/domain-uuid/strategy-alignments",
    "next": "/api/v1/business-domains/domain-uuid/strategy-alignments?cursor=eyJpZCI6ImRvbS1zdHJhdC11dWlkLTUwIn0",
    "businessDomain": "/api/v1/business-domains/domain-uuid"
  }
}
```

---

## Part 4: Cross-Domain Comparison API

### GET /api/v1/capabilities/{capabilityId}/strategy-alignments

Get all strategy alignments for a capability across all business domains.

**Response**: 200 OK
```json
{
  "data": [
    {
      "id": "dom-strat-uuid-1",
      "businessDomainId": "domain-uuid-1",
      "businessDomainName": "Digital Banking",
      "pillarId": "pillar-uuid-1",
      "pillarName": "Transform",
      "strategicImportance": 5,
      "alignedAt": "2025-12-26T10:00:00Z"
    },
    {
      "id": "dom-strat-uuid-2",
      "businessDomainId": "domain-uuid-2",
      "businessDomainName": "Traditional Lending",
      "pillarId": "pillar-uuid-1",
      "pillarName": "Transform",
      "strategicImportance": 3,
      "alignedAt": "2025-12-26T11:00:00Z"
    }
  ],
  "_links": {
    "self": "/api/v1/capabilities/cap-uuid/strategy-alignments",
    "capability": "/api/v1/capabilities/cap-uuid"
  }
}
```

This enables capability owners to see how their capability is valued across different domains.

### GET /api/v1/strategy-pillars/{pillarId}/alignments

Get all alignments for a specific strategy pillar across all domains.

**Query Parameters**:
- `domainId`: Filter by specific domain
- `minImportance`: Filter by minimum strategic importance (1-5)
- `limit`: Page size (default: 50, max: 100)
- `cursor`: Pagination cursor

**Response**: 200 OK
```json
{
  "data": [
    {
      "id": "dom-strat-uuid-1",
      "businessDomainId": "domain-uuid-1",
      "businessDomainName": "Digital Banking",
      "capabilityId": "cap-uuid-1",
      "capabilityName": "Customer Onboarding",
      "capabilityLevel": "L1",
      "strategicImportance": 5,
      "alignedAt": "2025-12-26T10:00:00Z"
    }
  ],
  "pagination": {
    "hasMore": false,
    "limit": 50
  },
  "_links": {
    "self": "/api/v1/strategy-pillars/pillar-uuid-1/alignments",
    "pillar": "/api/v1/meta-model/strategy-pillars/pillar-uuid-1"
  }
}
```

---

## Part 5: Strategic Importance Reference

### GET /api/v1/capabilities/metadata/strategic-importance-levels

Reference data for strategic importance scale.

**Response**: 200 OK
```json
{
  "data": [
    { "value": 1, "label": "Low", "description": "Low strategic importance" },
    { "value": 2, "label": "Below Average", "description": "Below average strategic importance" },
    { "value": 3, "label": "Average", "description": "Average strategic importance" },
    { "value": 4, "label": "Above Average", "description": "Above average strategic importance" },
    { "value": 5, "label": "Critical", "description": "Critical strategic importance" }
  ],
  "_links": {
    "self": "/api/v1/capabilities/metadata/strategic-importance-levels"
  }
}
```

---

## Backward Compatibility

### GET /api/v1/capabilities/metadata/strategy-pillars (Deprecated)

The existing hardcoded endpoint will be updated to read from configuration.

**Current Response** (hardcoded):
```json
{
  "data": [
    { "value": "AlwaysOn" },
    { "value": "Grow" },
    { "value": "Transform" }
  ]
}
```

**New Behavior**:
- Returns active pillars from configuration
- Adds deprecation header: `Deprecation: true`
- Links to new endpoint

**Transition Period**:
- Existing StrategyPillar and PillarWeight fields in Capability DTO remain available (read-only)
- UpdateMetadata endpoint ignores these fields (they are now managed via alignments API)

---

## Caching

### Strategy Pillar Configuration
- `Cache-Control: private, max-age=300` (5 minutes, tenant-specific)
- `ETag` based on configuration version

### Alignments
- `Cache-Control: private, no-cache` (always fresh due to frequent updates)

### Reference Data
- `Cache-Control: public, max-age=86400` (24 hours, static data)

---

## HATEOAS Links Summary

### Pillar Resource Links
- `self`: Link to current pillar
- `update`: PUT endpoint (if user has PermMetaModelWrite)
- `delete`: DELETE endpoint (if user has PermMetaModelWrite)
- `collection`: Link to pillars collection
- `alignments`: Link to alignments for this pillar

### Alignment Resource Links
- `self`: Link to current alignment
- `update`: PUT endpoint (if user has PermCapabilityWrite)
- `delete`: DELETE endpoint (if user has PermCapabilityWrite)
- `businessDomain`: Link to parent business domain
- `capability`: Link to parent capability
- `pillar`: Link to strategy pillar

---

## Permissions

### MetaModel Endpoints
- GET endpoints: PermMetaModelRead
- POST/PUT/DELETE: PermMetaModelWrite

### Capability Alignment Endpoints
- GET endpoints: PermCapabilityRead
- POST/PUT/DELETE: PermCapabilityWrite

---

## Checklist
- [ ] Specification ready
- [ ] Part 1: Strategy Pillar Configuration API
  - [ ] Route registration
  - [ ] GET collection handler
  - [ ] GET single handler
  - [ ] POST handler
  - [ ] PUT handler
  - [ ] DELETE handler
  - [ ] HATEOAS links
  - [ ] OpenAPI documentation
- [ ] Part 2: Domain-Scoped Alignment API
  - [ ] Route registration
  - [ ] GET collection handler
  - [ ] POST handler
  - [ ] PUT handler
  - [ ] DELETE handler
  - [ ] HATEOAS links
  - [ ] OpenAPI documentation
- [ ] Part 3: Domain Portfolio Query API
  - [ ] Route registration
  - [ ] Handler with pagination
  - [ ] OpenAPI documentation
- [ ] Part 4: Cross-Domain Comparison API
  - [ ] Capability alignments endpoint
  - [ ] Pillar alignments endpoint
  - [ ] Pagination implementation
- [ ] Part 5: Reference Data
  - [ ] Strategic importance levels endpoint
  - [ ] Deprecated strategy-pillars endpoint update
- [ ] Caching headers implemented
- [ ] Unit tests for handlers
- [ ] Integration tests for API
- [ ] User sign-off
