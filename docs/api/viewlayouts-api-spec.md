# ViewLayouts API Specification (REST Level 3)

## Overview

RESTful API for managing visual layout containers and element positions across different UI contexts with concurrent editing support.

## Base URL

```
/api/v1/layouts
```

## Design Decisions

### URL Structure

Uses path parameters for context identification (consistent with existing codebase patterns):
- `/layouts/{contextType}/{contextRef}` - Layout container for a specific context
- `/layouts/{contextType}/{contextRef}/elements/{elementId}` - Individual element position

### Context Types

| Context Type | Context Ref | Use Case |
|--------------|-------------|----------|
| `architecture-canvas` | View ID | Component positions on architecture canvas |
| `business-domain-grid` | Domain ID | Capability positions in business domain grid |

### Concurrency Strategy

- **Optimistic locking** using ETags for container-level operations (preferences)
- **Last-write-wins** for individual element position operations (acceptable for position data)
- **Atomic batch** operations for multiple element updates

---

## Endpoints

### 1. Get Layout Container

```http
GET /api/v1/layouts/{contextType}/{contextRef}
```

**Purpose**: Retrieve layout container with all element positions for a specific context.

**Path Parameters**:
- `contextType`: `architecture-canvas` | `business-domain-grid`
- `contextRef`: Reference ID (view_id or domain_id)

**Success Response (200 OK)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "contextType": "business-domain-grid",
  "contextRef": "domain-finance",
  "preferences": {
    "colorScheme": "pastel",
    "layoutDirection": "TB"
  },
  "elements": [
    {
      "elementId": "cap-123",
      "x": 100,
      "y": 200,
      "customColor": "#3b82f6",
      "_links": {
        "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-123" },
        "update": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-123", "method": "PUT" },
        "delete": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-123", "method": "DELETE" }
      }
    },
    {
      "elementId": "cap-456",
      "x": 300,
      "y": 150,
      "sortOrder": 2,
      "_links": {
        "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-456" },
        "update": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-456", "method": "PUT" },
        "delete": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-456", "method": "DELETE" }
      }
    }
  ],
  "version": 5,
  "createdAt": "2025-11-30T10:00:00Z",
  "updatedAt": "2025-11-30T14:23:15Z",
  "_links": {
    "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" },
    "updatePreferences": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/preferences", "method": "PATCH" },
    "batchUpdate": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements", "method": "PATCH" },
    "delete": { "href": "/api/v1/layouts/business-domain-grid/domain-finance", "method": "DELETE" }
  }
}
```

**Response Headers**:
```
ETag: "5"
```

**Not Found (404 Not Found)**:
```json
{
  "error": "Not Found",
  "message": "No layout found for business-domain-grid/domain-finance",
  "_links": {
    "create": { "href": "/api/v1/layouts/business-domain-grid/domain-finance", "method": "PUT" }
  }
}
```

---

### 2. Create or Update Layout Container

```http
PUT /api/v1/layouts/{contextType}/{contextRef}
```

**Purpose**: Create a new layout container or update preferences if it exists (upsert).

**Request Body**:
```json
{
  "preferences": {
    "colorScheme": "pastel",
    "layoutDirection": "TB"
  }
}
```

**Success Response (201 Created)** - New container:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "contextType": "business-domain-grid",
  "contextRef": "domain-finance",
  "preferences": {
    "colorScheme": "pastel",
    "layoutDirection": "TB"
  },
  "elements": [],
  "version": 1,
  "createdAt": "2025-11-30T10:00:00Z",
  "updatedAt": "2025-11-30T10:00:00Z",
  "_links": {
    "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" }
  }
}
```

**Response Headers**:
```
Location: /api/v1/layouts/business-domain-grid/domain-finance
ETag: "1"
```

**Success Response (200 OK)** - Updated existing:
Same structure with updated version.

---

### 3. Delete Layout Container

```http
DELETE /api/v1/layouts/{contextType}/{contextRef}
```

**Purpose**: Delete layout container and all its element positions.

**Success Response (204 No Content)**: Empty body

**Not Found (404 Not Found)**:
```json
{
  "error": "Not Found",
  "message": "Layout not found"
}
```

---

### 4. Update Container Preferences

```http
PATCH /api/v1/layouts/{contextType}/{contextRef}/preferences
```

**Purpose**: Update only the preferences object, using optimistic locking.

**Request Headers**:
```
If-Match: "5"
```

**Request Body**:
```json
{
  "colorScheme": "dark",
  "layoutDirection": "LR"
}
```

**Success Response (200 OK)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "contextType": "business-domain-grid",
  "contextRef": "domain-finance",
  "preferences": {
    "colorScheme": "dark",
    "layoutDirection": "LR"
  },
  "version": 6,
  "_links": {
    "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" }
  }
}
```

**Response Headers**:
```
ETag: "6"
```

**Precondition Failed (412 Precondition Failed)**:
```json
{
  "error": "Precondition Failed",
  "message": "Layout was modified by another user. Please refresh and try again.",
  "currentVersion": 7,
  "_links": {
    "current": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" }
  }
}
```

---

### 5. Upsert Element Position

```http
PUT /api/v1/layouts/{contextType}/{contextRef}/elements/{elementId}
```

**Purpose**: Create or replace an element's position. Idempotent, last-write-wins.

**Request Body**:
```json
{
  "x": 120.5,
  "y": 200.0,
  "width": 180.0,
  "height": 100.0,
  "customColor": "#FF5733",
  "sortOrder": 1
}
```

**Success Response (200 OK)** - Updated existing:
```json
{
  "elementId": "cap-123",
  "x": 120.5,
  "y": 200.0,
  "width": 180.0,
  "height": 100.0,
  "customColor": "#FF5733",
  "sortOrder": 1,
  "_links": {
    "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-123" },
    "layout": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" },
    "delete": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-123", "method": "DELETE" }
  }
}
```

**Success Response (201 Created)** - New element:
Same body with Location header.

**Layout Not Found (404 Not Found)**:
Layout container must exist before adding elements. Frontend should call PUT on container first.

---

### 6. Delete Element Position

```http
DELETE /api/v1/layouts/{contextType}/{contextRef}/elements/{elementId}
```

**Success Response (204 No Content)**: Empty body

---

### 7. Batch Update Element Positions

```http
PATCH /api/v1/layouts/{contextType}/{contextRef}/elements
```

**Purpose**: Update multiple element positions atomically.

**Request Body**:
```json
{
  "updates": [
    { "elementId": "cap-123", "x": 130.5, "y": 210.0 },
    { "elementId": "cap-456", "x": 360.0, "y": 160.0, "sortOrder": 3 }
  ]
}
```

**Success Response (200 OK)**:
```json
{
  "updated": 2,
  "elements": [
    {
      "elementId": "cap-123",
      "x": 130.5,
      "y": 210.0,
      "_links": { "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-123" } }
    },
    {
      "elementId": "cap-456",
      "x": 360.0,
      "y": 160.0,
      "sortOrder": 3,
      "_links": { "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements/cap-456" } }
    }
  ],
  "_links": {
    "self": { "href": "/api/v1/layouts/business-domain-grid/domain-finance/elements" },
    "layout": { "href": "/api/v1/layouts/business-domain-grid/domain-finance" }
  }
}
```

---

## Status Code Summary

| Code | Usage |
|------|-------|
| 200 OK | Successful GET, successful PUT (update), successful PATCH |
| 201 Created | Successful PUT (create) |
| 204 No Content | Successful DELETE |
| 400 Bad Request | Validation errors, malformed requests |
| 404 Not Found | Layout or element not found |
| 412 Precondition Failed | ETag mismatch in optimistic locking |

---

## Concurrency Handling

### Container Preferences (PATCH preferences)
- **Mechanism**: Optimistic locking via ETag / If-Match
- **ETag value**: String representation of version number
- **Client workflow**:
  1. GET layout, receive ETag: "5"
  2. User modifies preferences
  3. PATCH with If-Match: "5"
  4. If successful, receive new ETag: "6"
  5. If 412, refresh and retry with new ETag

### Element Positions (PUT, DELETE single element)
- **Mechanism**: Last-write-wins (no version checking)
- **Rationale**: Position conflicts are visually obvious and self-correcting
- **Trade-off**: Simpler implementation, acceptable UX for visual editing

### Batch Element Operations (PATCH elements)
- **Mechanism**: Atomic batch within database transaction
- **All-or-nothing**: If any update fails, entire batch is rolled back

---

## Event Integration

When domain entities are deleted, event handlers clean up orphaned positions:

| Domain Event | Handler Action |
|--------------|---------------|
| ComponentDeleted | Remove from all `architecture-canvas` layouts |
| CapabilityDeleted | Remove from all `business-domain-grid` layouts |
| ViewDeleted | Delete `architecture-canvas` layout for that view |
| BusinessDomainDeleted | Delete `business-domain-grid` layout for that domain |

These are handled by event processors, not exposed via REST API.
