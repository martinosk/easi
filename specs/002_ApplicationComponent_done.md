# Application Component Feature

## Description
Implements the ability to model Application Components. An Application Component represents a modular, deployable, and replaceable part of a software system that encapsulates its contents and exposes its functionality through a set of interfaces.

## Command

### CreateApplicationComponent
Creates a new Application Component in the system.

**Properties:**
- `Name` (string, required): The name of the application component
- `Description` (string, optional): A description of the application component's purpose and functionality

**Validation Rules:**
- Name must not be empty or whitespace only

## Event

### ApplicationComponentCreated
Indicates that a new Application Component has been successfully created.

**Properties:**
- `Id` (Guid, required): Unique identifier for the application component
- `Name` (string, required): The name of the application component
- `Description` (string, optional): The description of the application component
- `CreatedAt` (DateTime, required): Timestamp when the component was created

## API Endpoints

### POST /api/application-component
Creates a new application component.

**Request Body:**
```json
{
  "name": "string",
  "description": "string"
}
```

**Response:** 201 Created
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
    "reference": {
      "href": "/api/v1/reference/components",
      "title": "Application Component Reference"
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

### GET /api/application-component
Gets all application components.

**Response:** 200 OK
```json
[
  {
    "id": "guid",
    "name": "string",
    "description": "string",
    "createdAt": "datetime",
    "_links": {
      "self": {
        "href": "/api/application-component/{id}"
      },
      "reference": {
        "href": "/api/v1/reference/components",
        "title": "Application Component Reference"
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
]
```

### GET /api/application-component/{id}
Gets a specific application component by ID.

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
    "reference": {
      "href": "/api/v1/reference/components",
      "title": "Application Component Reference"
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

**Error Responses:**
- 404 Not Found: Component does not exist

## Checklist
- [x] Specification ready
- [x] Command handler implemented
- [x] Event projection implemented
- [x] API endpoint implemented
- [x] OpenAPI specification documented
- [x] Unit tests implemented and passing
- [x] Integration tests implemented (API endpoints covered)
- [x] HATEOAS links implementation
- [x] Read model updated with _links property
- [x] API responses include navigation links
- [ ] User sign-off
