# Component Relation Feature

## Description
Implements the ability to model relationships between Application Components. Relations are unidirectional, representing how one component interacts with or depends on another component. This follows ArchiMate modelling principles where relationships have specific types and semantics.

## Relation Types
- **Triggers**: Indicates that the source component initiates or activates functionality in the target component
- **Serves**: Indicates that the source component provides services or functionality to the target component

## Command

### CreateComponentRelation
Creates a new relation between two Application Components.

**Properties:**
- `SourceComponentId` (Guid, required): The ID of the component that is the source of the relation
- `TargetComponentId` (Guid, required): The ID of the component that is the target of the relation
- `RelationType` (enum, required): The type of relation (Triggers, Serves)
- `Name` (string, optional): A descriptive name for the relation
- `Description` (string, optional): Additional details about the relation

**Validation Rules:**
- SourceComponentId must not be empty
- TargetComponentId must not be empty
- SourceComponentId and TargetComponentId must reference existing Application Components
- SourceComponentId must not equal TargetComponentId (no self-references)
- RelationType must be a valid enum value

## Event

### ComponentRelationCreated
Indicates that a new relation between two Application Components has been successfully created.

**Properties:**
- `Id` (Guid, required): Unique identifier for the relation
- `SourceComponentId` (Guid, required): The ID of the source component
- `TargetComponentId` (Guid, required): The ID of the target component
- `RelationType` (enum, required): The type of relation
- `Name` (string, optional): The name of the relation
- `Description` (string, optional): The description of the relation
- `CreatedAt` (DateTime, required): Timestamp when the relation was created

## API Endpoints

### POST /api/component-relation
Creates a new component relation.

**Request Body:**
```json
{
  "sourceComponentId": "guid",
  "targetComponentId": "guid",
  "relationType": "Triggers" | "Serves",
  "name": "string (optional)",
  "description": "string (optional)"
}
```

**Response:** 201 Created
```json
{
  "id": "guid",
  "sourceComponentId": "guid",
  "targetComponentId": "guid",
  "relationType": "Triggers" | "Serves",
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
```

**Error Responses:**
- 400 Bad Request: Invalid input (empty IDs, invalid relation type, self-reference)
- 404 Not Found: Source or target component does not exist

### GET /api/component-relation
Gets all component relations.

**Response:** 200 OK
```json
[
  {
    "id": "guid",
    "sourceComponentId": "guid",
    "targetComponentId": "guid",
    "relationType": "Triggers" | "Serves",
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

### GET /api/component-relation/from/{componentId}
Gets all relations where the specified component is the source.

**Response:** 200 OK
```json
[
  {
    "id": "guid",
    "sourceComponentId": "guid",
    "targetComponentId": "guid",
    "relationType": "Triggers" | "Serves",
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

### GET /api/component-relation/to/{componentId}
Gets all relations where the specified component is the target.

**Response:** 200 OK
```json
[
  {
    "id": "guid",
    "sourceComponentId": "guid",
    "targetComponentId": "guid",
    "relationType": "Triggers" | "Serves",
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

### GET /api/component-relation/{id}
Gets a specific component relation by ID.

**Response:** 200 OK
```json
{
  "id": "guid",
  "sourceComponentId": "guid",
  "targetComponentId": "guid",
  "relationType": "Triggers" | "Serves",
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
```

**Error Responses:**
- 404 Not Found: Relation does not exist

## Domain Model

### ComponentRelation Aggregate
- Properties: Id, SourceComponentId, TargetComponentId, RelationType, Name, Description, CreatedAt
- All properties should be value objects following tactical DDD principles

### Value Objects
- `ComponentRelationId`: Wraps Guid with validation
- `RelationType`: Enum-based value object (Triggers, Serves)
- `RelationName`: Wraps optional string
- `RelationDescription`: Wraps optional string
- Reuse: `ApplicationComponentId` (already exists), `CreatedAt` (already exists)

## Bounded Context Considerations
- ComponentRelation is part of the ArchitectureModelling bounded context
- Relations only reference components by ID (no direct aggregate references)
- If a component is deleted, consideration needed for orphaned relations (future requirement)

## Checklist
- [x] Specification ready
- [x] Value objects created
- [x] Aggregate implemented
- [x] Command handler implemented
- [x] Event projection implemented
- [x] Read models created
- [x] API endpoints implemented
- [x] OpenAPI specification documented
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented
- [ ] User sign-off
