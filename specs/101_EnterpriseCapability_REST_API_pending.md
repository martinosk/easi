# Enterprise Capability REST API

## Description
REST Level 3 API with HATEOAS for managing enterprise capabilities and their links to domain capabilities.

## Purpose
Enable architects to discover capability overlap, create enterprise capability groupings, and analyze cross-domain maturity gaps.

## Bounded Context
**Enterprise Architecture** - See Spec 100 for context definition.

## Dependencies
- Spec 100: Enterprise Capability

## Part 1: Enterprise Capability Management API

### Base Path
`/api/v1/enterprise-architecture/enterprise-capabilities`

### GET /api/v1/enterprise-architecture/enterprise-capabilities

Get all enterprise capabilities for the authenticated tenant.

**Query Parameters**:
- `category`: Filter by category
- `includeInactive=true`: Include soft-deleted capabilities
- `limit`: Page size (default: 50)
- `cursor`: Pagination cursor

**Response**: 200 OK
```json
{
  "data": [
    {
      "id": "ent-cap-uuid-1",
      "name": "Payroll",
      "description": "Employee compensation processing",
      "category": "HR Operations",
      "linkedCapabilityCount": 4,
      "domainCount": 3,
      "isActive": true
    }
  ],
  "pagination": {
    "hasMore": false,
    "limit": 50
  },
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities",
    "create": "/api/v1/enterprise-architecture/enterprise-capabilities"
  }
}
```

### GET /api/v1/enterprise-architecture/enterprise-capabilities/{id}

Get a specific enterprise capability with linked capabilities.

**Response**: 200 OK
```json
{
  "id": "ent-cap-uuid-1",
  "name": "Payroll",
  "description": "Employee compensation processing",
  "category": "HR Operations",
  "isActive": true,
  "version": 1,
  "createdAt": "2025-12-26T10:00:00Z",
  "createdBy": "architect@company.com",
  "linkedCapabilities": [
    {
      "linkId": "ent-link-uuid-1",
      "capabilityId": "cap-uuid-1",
      "capabilityName": "Payroll Management",
      "businessDomainId": "domain-uuid-1",
      "businessDomainName": "IT Support",
      "maturityValue": 15,
      "maturitySection": "Genesis",
      "linkedAt": "2025-12-26T11:00:00Z"
    },
    {
      "linkId": "ent-link-uuid-2",
      "capabilityId": "cap-uuid-2",
      "capabilityName": "Salary Processing",
      "businessDomainId": "domain-uuid-2",
      "businessDomainName": "Customer Service",
      "maturityValue": 65,
      "maturitySection": "Product",
      "linkedAt": "2025-12-26T11:30:00Z"
    }
  ],
  "strategyAlignments": [
    {
      "alignmentId": "ent-strat-uuid-1",
      "pillarId": "pillar-uuid-1",
      "pillarName": "Standardization",
      "strategicImportance": 5
    }
  ],
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "update": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "delete": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "links": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links",
    "strategyAlignments": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/strategy-alignments",
    "collection": "/api/v1/enterprise-architecture/enterprise-capabilities"
  }
}
```

### POST /api/v1/enterprise-architecture/enterprise-capabilities

Create a new enterprise capability.

**Request Body**:
```json
{
  "name": "Payroll",
  "description": "Employee compensation processing",
  "category": "HR Operations"
}
```

**Response**: 201 Created
```json
{
  "id": "ent-cap-uuid-1",
  "name": "Payroll",
  "description": "Employee compensation processing",
  "category": "HR Operations",
  "isActive": true,
  "version": 1,
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "links": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links"
  }
}
```

**Headers**:
- `Location: /api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1`

**Error Responses**:
- 400 Bad Request: Name empty or exceeds 200 characters
- 409 Conflict: Name already exists

### PUT /api/v1/enterprise-architecture/enterprise-capabilities/{id}

Update an enterprise capability.

**Request Body**:
```json
{
  "name": "Payroll & Compensation",
  "description": "Updated description",
  "category": "HR Operations",
  "version": 1
}
```

**Response**: 200 OK (same structure as GET)

**Error Responses**:
- 400 Bad Request: Validation errors
- 404 Not Found: Enterprise capability not found
- 409 Conflict: Version mismatch or name already exists

### DELETE /api/v1/enterprise-architecture/enterprise-capabilities/{id}

Delete an enterprise capability (soft delete).

**Response**: 204 No Content

**Error Responses**:
- 404 Not Found: Enterprise capability not found

---

## Part 2: Capability Linking API

### Base Path
`/api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/links`

### GET /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/links

Get all linked domain capabilities.

**Response**: 200 OK
```json
{
  "data": [
    {
      "linkId": "ent-link-uuid-1",
      "capabilityId": "cap-uuid-1",
      "capabilityName": "Payroll Management",
      "businessDomainId": "domain-uuid-1",
      "businessDomainName": "IT Support",
      "maturityValue": 15,
      "maturitySection": "Genesis",
      "linkedAt": "2025-12-26T11:00:00Z",
      "linkedBy": "architect@company.com"
    }
  ],
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links",
    "enterpriseCapability": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "create": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links"
  }
}
```

### POST /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/links

Link a domain capability to this enterprise capability.

**Request Body**:
```json
{
  "capabilityId": "cap-uuid-3"
}
```

**Response**: 201 Created
```json
{
  "linkId": "ent-link-uuid-3",
  "capabilityId": "cap-uuid-3",
  "capabilityName": "Compensation Admin",
  "businessDomainId": "domain-uuid-3",
  "businessDomainName": "Finance",
  "maturityValue": 45,
  "maturitySection": "Custom Built",
  "linkedAt": "2025-12-26T12:00:00Z",
  "linkedBy": "architect@company.com",
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links/ent-link-uuid-3",
    "delete": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links/ent-link-uuid-3",
    "capability": "/api/v1/capabilities/cap-uuid-3",
    "enterpriseCapability": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1"
  }
}
```

**Error Responses**:
- 400 Bad Request: Capability already linked to another enterprise capability
- 404 Not Found: Enterprise capability or domain capability not found
- 409 Conflict: Capability already linked to this enterprise capability

### DELETE /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/links/{linkId}

Unlink a domain capability.

**Response**: 204 No Content

**Error Responses**:
- 404 Not Found: Link not found

---

## Part 3: Enterprise Strategy Alignment API

### Base Path
`/api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/strategy-alignments`

### GET /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/strategy-alignments

Get strategy alignments for an enterprise capability.

**Response**: 200 OK
```json
{
  "data": [
    {
      "alignmentId": "ent-strat-uuid-1",
      "pillarId": "pillar-uuid-1",
      "pillarName": "Standardization",
      "strategicImportance": 5,
      "alignedAt": "2025-12-26T10:00:00Z"
    }
  ],
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/strategy-alignments",
    "enterpriseCapability": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "create": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/strategy-alignments",
    "availablePillars": "/api/v1/meta-model/strategy-pillars"
  }
}
```

### POST /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/strategy-alignments

Align enterprise capability to a strategy pillar.

**Request Body**:
```json
{
  "pillarId": "pillar-uuid-1",
  "strategicImportance": 5
}
```

**Response**: 201 Created (same structure as GET item)

**Error Responses**:
- 400 Bad Request: Invalid strategic importance (must be 1-5)
- 404 Not Found: Enterprise capability or pillar not found
- 409 Conflict: Already aligned to this pillar

### PUT /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/strategy-alignments/{alignmentId}

Update strategic importance.

**Request Body**:
```json
{
  "strategicImportance": 4
}
```

**Response**: 200 OK

### DELETE /api/v1/enterprise-architecture/enterprise-capabilities/{enterpriseCapabilityId}/strategy-alignments/{alignmentId}

Remove alignment.

**Response**: 204 No Content

---

## Part 4: Analysis APIs

### GET /api/v1/enterprise-architecture/enterprise-capabilities/standardization-candidates

Get enterprise capabilities marked for standardization with multiple implementations.

**Response**: 200 OK
```json
{
  "data": [
    {
      "enterpriseCapabilityId": "ent-cap-uuid-1",
      "enterpriseCapabilityName": "Payroll",
      "standardizationImportance": 5,
      "implementationCount": 4,
      "domainCount": 3,
      "minMaturity": 15,
      "maxMaturity": 65,
      "maturitySpread": 50,
      "domains": [
        {
          "domainId": "domain-uuid-1",
          "domainName": "IT Support",
          "capabilityName": "Payroll Management",
          "maturityValue": 15,
          "maturitySection": "Genesis"
        },
        {
          "domainId": "domain-uuid-2",
          "domainName": "Customer Service",
          "capabilityName": "Salary Processing",
          "maturityValue": 65,
          "maturitySection": "Product"
        }
      ]
    }
  ],
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/standardization-candidates"
  }
}
```

### GET /api/v1/enterprise-architecture/enterprise-capabilities/{id}/maturity-gaps

Get maturity gap analysis for an enterprise capability.

**Response**: 200 OK
```json
{
  "enterpriseCapabilityId": "ent-cap-uuid-1",
  "enterpriseCapabilityName": "Payroll",
  "targetMaturity": 65,
  "gaps": [
    {
      "businessDomainId": "domain-uuid-1",
      "businessDomainName": "IT Support",
      "capabilityId": "cap-uuid-1",
      "capabilityName": "Payroll Management",
      "currentMaturity": 15,
      "currentSection": "Genesis",
      "targetMaturity": 65,
      "targetSection": "Product",
      "gapValue": 50,
      "investmentPriority": "High"
    },
    {
      "businessDomainId": "domain-uuid-2",
      "businessDomainName": "Customer Service",
      "capabilityId": "cap-uuid-2",
      "capabilityName": "Salary Processing",
      "currentMaturity": 65,
      "currentSection": "Product",
      "targetMaturity": 65,
      "targetSection": "Product",
      "gapValue": 0,
      "investmentPriority": "None"
    }
  ],
  "_links": {
    "self": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/maturity-gaps",
    "enterpriseCapability": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1"
  }
}
```

---

## Part 5: Discovery API

### GET /api/v1/capabilities/{capabilityId}/enterprise-link

Check if a domain capability is linked to an enterprise capability.

**Response**: 200 OK (if linked)
```json
{
  "linkId": "ent-link-uuid-1",
  "enterpriseCapabilityId": "ent-cap-uuid-1",
  "enterpriseCapabilityName": "Payroll",
  "linkedAt": "2025-12-26T11:00:00Z",
  "_links": {
    "self": "/api/v1/capabilities/cap-uuid-1/enterprise-link",
    "enterpriseCapability": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1",
    "unlink": "/api/v1/enterprise-architecture/enterprise-capabilities/ent-cap-uuid-1/links/ent-link-uuid-1"
  }
}
```

**Response**: 404 Not Found (if not linked)
```json
{
  "error": "Not Found",
  "message": "Capability is not linked to any enterprise capability",
  "_links": {
    "enterpriseCapabilities": "/api/v1/enterprise-architecture/enterprise-capabilities"
  }
}
```

---

## Permissions

### Enterprise Architecture Endpoints
- GET endpoints: PermEnterpriseArchitectureRead
- POST/PUT/DELETE: PermEnterpriseArchitectureWrite

Note: New permissions will need to be added to the platform context permission system.

---

## Checklist
- [ ] Specification ready
- [ ] Part 1: Enterprise Capability CRUD
  - [ ] Route registration
  - [ ] GET collection handler
  - [ ] GET single handler
  - [ ] POST handler
  - [ ] PUT handler
  - [ ] DELETE handler
  - [ ] OpenAPI documentation
- [ ] Part 2: Linking API
  - [ ] Route registration
  - [ ] GET links handler
  - [ ] POST link handler
  - [ ] DELETE link handler
  - [ ] OpenAPI documentation
- [ ] Part 3: Strategy Alignment API
  - [ ] Route registration
  - [ ] CRUD handlers
  - [ ] OpenAPI documentation
- [ ] Part 4: Analysis APIs
  - [ ] Standardization candidates endpoint
  - [ ] Maturity gaps endpoint
- [ ] Part 5: Discovery API
  - [ ] Enterprise link check endpoint
- [ ] Unit tests
- [ ] Integration tests
- [ ] User sign-off
