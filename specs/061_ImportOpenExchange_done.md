# Import from Open Exchange Files

## User Need
As an enterprise architect, I need to import capability models and application components from ArchiMate Open Exchange files so I can migrate existing architecture documentation into this tool without manually re-entering hundreds of elements.

## Dependencies
- Spec 023: Capability Model (done)
- Spec 002: ApplicationComponent (done)
- Spec 026: CapabilitySystemRealization (done)
- Spec 053-056: BusinessDomain (done)

## Architecture Decision

This feature introduces a new **Import bounded context** with backend orchestration. This approach:
- Keeps parsing and orchestration logic server-side
- Enables future format support (TOGAF, Sparx EA, CSV) via adapter pattern
- Provides transaction safety with rollback on failure
- Handles large files without browser memory constraints
- Issues commands to existing CapabilityMapping and ArchitectureModeling contexts

## ArchiMate Open Exchange Format

### Supported Element Types
| ArchiMate Type | Maps To | Notes |
|----------------|---------|-------|
| `Capability` | Capability | Import name, documentation as description |
| `ApplicationComponent` | Component | Import name, documentation as description |
| `ApplicationService` | Component | Treat as component, import name/documentation |

### Supported Relationship Types
| ArchiMate Type | Condition | Maps To |
|----------------|-----------|---------|
| `Aggregation` | Between Capabilities | Parent-child hierarchy (child.parentId = source) |
| `Composition` | Between Capabilities | Parent-child hierarchy (child.parentId = source) |
| `Realization` | From ApplicationComponent/ApplicationService to Capability | CapabilityRealization with notes from relationship name/documentation |

### Unsupported Types (examples)
- BusinessProcess, BusinessActor, BusinessFunction, BusinessService
- DataObject, DataComponent
- Technology*, Node, Device, SystemSoftware
- All other ArchiMate element and relationship types

## API Contract

### File Upload Constraints
- Maximum file size: 50MB
- Accepted MIME types: `application/xml`, `text/xml`
- File extension: `.xml`
- Character encoding: UTF-8

### Session Lifecycle
- Import sessions in "pending" status expire after 1 hour
- Import sessions in "completed"/"failed" status are retained for 24 hours

### POST /api/v1/imports
Upload file and get preview of what will be imported.

**Request:** `multipart/form-data`
- `file`: XML file (required)
- `sourceFormat`: `archimate-openexchange` (required)
- `businessDomainId`: UUID (optional) - assign L1 capabilities to this domain

**Response 201 Created:**
```json
{
  "id": "import-session-uuid",
  "status": "pending",
  "sourceFormat": "archimate-openexchange",
  "businessDomainId": "domain-uuid-or-null",
  "preview": {
    "supported": {
      "capabilities": 45,
      "components": 12,
      "parentChildRelationships": 38,
      "realizations": 15
    },
    "unsupported": {
      "elements": {
        "BusinessProcess": 5,
        "DataObject": 3
      },
      "relationships": {
        "Flow": 12,
        "Serving": 4
      }
    }
  },
  "createdAt": "2025-01-15T10:30:00Z",
  "_links": {
    "self": "/api/v1/imports/import-session-uuid",
    "confirm": "/api/v1/imports/import-session-uuid/confirm",
    "delete": "/api/v1/imports/import-session-uuid"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields or invalid sourceFormat
- `413 Payload Too Large` - File exceeds 50MB limit
- `415 Unsupported Media Type` - Not multipart/form-data or not XML file
- `422 Unprocessable Entity` - Valid XML but not valid ArchiMate format

### POST /api/v1/imports/{id}/confirm
Confirm and execute the import. Returns immediately with 202 Accepted.

**Response 202 Accepted:**

Headers: `Retry-After: 2` (suggests polling interval)

```json
{
  "id": "import-session-uuid",
  "status": "importing",
  "progress": {
    "phase": "creating_components",
    "totalItems": 110,
    "completedItems": 0
  },
  "_links": {
    "self": "/api/v1/imports/import-session-uuid"
  }
}
```

**Error Responses:**
- `404 Not Found` - Import session not found
- `409 Conflict` - Import already started/completed, or session expired

### GET /api/v1/imports/{id}
Get current import session status and progress.

**Response 200 OK (during import):**
```json
{
  "id": "import-session-uuid",
  "status": "importing",
  "progress": {
    "phase": "creating_capabilities",
    "totalItems": 110,
    "completedItems": 57
  },
  "errors": [],
  "_links": {
    "self": "/api/v1/imports/import-session-uuid"
  }
}
```

**Response 200 OK (completed):**
```json
{
  "id": "import-session-uuid",
  "status": "completed",
  "result": {
    "capabilitiesCreated": 45,
    "componentsCreated": 12,
    "realizationsCreated": 15,
    "domainAssignments": 8,
    "errors": [
      {
        "sourceElement": "capability-xyz",
        "sourceName": "Some Capability",
        "error": "Parent capability not found",
        "action": "skipped"
      }
    ]
  },
  "completedAt": "2025-01-15T10:32:15Z",
  "_links": {
    "self": "/api/v1/imports/import-session-uuid"
  }
}
```

### DELETE /api/v1/imports/{id}
Cancel a pending import session (before confirmation).

**Response 204 No Content**

**Error Responses:**
- `404 Not Found` - Import session not found
- `409 Conflict` - Cannot cancel, import already started or completed

### Error Response Structure
All error responses follow the standard format:
```json
{
  "error": "Bad Request",
  "message": "Invalid source format",
  "details": {
    "sourceFormat": "must be 'archimate-openexchange'"
  }
}
```

### HATEOAS Link Rules
Links are state-dependent:
- `pending` status: includes `self`, `confirm`, `delete`
- `importing` status: includes `self` only
- `completed`/`failed` status: includes `self` only

## Import Workflow

### Step 1: File Upload
User clicks "Import" in Architecture Canvas toolbar, selects an XML file, optionally selects a business domain. Frontend uploads to `POST /api/v1/imports`.

### Step 2: Preview
Backend parses XML and returns preview showing:
- **Will Import:** Capabilities, Components, Parent-child relationships, Realizations (with counts)
- **Will NOT Import:** Unsupported element types and relationship types (with counts)

Frontend displays this preview. User can cancel or proceed.

### Step 3: Confirm
User confirms. Frontend calls `POST /api/v1/imports/{id}/confirm`. Import executes asynchronously.

### Step 4: Progress
Frontend polls `GET /api/v1/imports/{id}` to show progress bar and current phase.

### Step 5: Results
When status is `completed` or `failed`, display summary with counts and any errors.

## Backend Structure

```
/backend/internal/importing/
├── domain/
│   ├── aggregates/
│   │   └── import_session.go
│   ├── valueobjects/
│   │   ├── source_format.go
│   │   ├── import_status.go
│   │   ├── import_preview.go
│   │   ├── import_progress.go
│   │   └── import_error.go
│   └── events/
│       ├── import_session_created.go
│       ├── import_started.go
│       ├── import_progress_updated.go
│       ├── import_completed.go
│       └── import_failed.go
├── application/
│   ├── commands/
│   │   ├── create_import_session.go
│   │   ├── confirm_import.go
│   │   └── cancel_import.go
│   ├── handlers/
│   │   └── import_handlers.go
│   ├── parsers/
│   │   ├── parser.go              (interface)
│   │   └── archimate_parser.go
│   ├── orchestrator/
│   │   └── import_orchestrator.go
│   └── readmodels/
│       └── import_session_readmodel.go
└── infrastructure/
    ├── api/
    │   └── import_api_handlers.go
    └── persistence/
        └── import_session_repository.go
```

## Frontend Structure

```
frontend/src/features/importing/
├── components/
│   ├── ImportDialog.tsx          # Main orchestrator (4 steps)
│   ├── ImportUploadStep.tsx      # File selection + domain dropdown
│   ├── ImportPreviewStep.tsx     # Will/won't import preview
│   ├── ImportProgressStep.tsx    # Progress bar + phase
│   ├── ImportResultsStep.tsx     # Summary + error list
│   └── ImportButton.tsx          # Toolbar button
├── hooks/
│   └── useImportSession.ts       # Session lifecycle management
├── types.ts                      # Import-specific TypeScript types
└── index.ts                      # Public exports
```

## Import Orchestration Logic

The ImportOrchestrator executes imports in this order:

1. **Create Components** (no dependencies)
   - Map source identifier → created component ID

2. **Create Capabilities** (level order: L1 → L2 → L3 → L4)
   - Use Aggregation/Composition relationships to determine hierarchy
   - Elements with no parent relationship = L1
   - Map source identifier → created capability ID

3. **Create Realizations** (after capabilities and components exist)
   - Use mapped IDs to link components to capabilities
   - Import relationship name/documentation as notes

4. **Assign to Business Domain** (if selected)
   - Assign L1 capabilities to the selected domain

Individual failures are recorded but do not abort the import.

## Success Criteria
- User can upload an ArchiMate Open Exchange XML file
- User sees clear preview of what will and will not be imported before confirming
- Capabilities are created with correct parent-child hierarchy
- Components are created from ApplicationComponent and ApplicationService elements
- Realization relationships create proper capability-system links with notes
- User can optionally assign imported L1 capabilities to a business domain
- Import failures for individual items do not abort the entire import
- User receives summary of import results with any errors listed

## Vertical Slices

### Slice 1: Import Context Foundation
- [x] Create importing bounded context folder structure
- [x] Define ImportSession aggregate with status (pending, importing, completed, failed)
- [x] Define value objects: SourceFormat, ImportStatus, ImportPreview, ImportProgress, ImportError
- [x] Define events: ImportSessionCreated, ImportStarted, ImportCompleted, ImportFailed
- [x] Create import_sessions table migration
- [x] Implement ImportSessionRepository

### Slice 2: File Upload and Preview API
- [x] Implement ArchiMate Open Exchange parser
- [x] Parse elements by type (Capability, ApplicationComponent, ApplicationService)
- [x] Parse relationships by type (Aggregation, Composition, Realization)
- [x] Count supported vs unsupported types for preview
- [x] Implement POST /api/v1/imports endpoint
- [x] Return preview with counts and HATEOAS links
- [x] Store parsed data in ImportSession for later execution

### Slice 3: Import Confirmation and Execution
- [x] Implement POST /api/v1/imports/{id}/confirm endpoint
- [x] Create ImportOrchestrator service
- [x] Execute component creation via ArchitectureModeling commands
- [x] Execute capability creation in hierarchy order via CapabilityMapping commands
- [x] Execute realization creation via CapabilityMapping commands
- [x] Track progress and errors in ImportSession
- [x] Update ImportSession status on completion

### Slice 4: Progress Polling and Results
- [x] Implement GET /api/v1/imports/{id} endpoint
- [x] Return current status, progress, and errors
- [x] Return final result summary when completed

### Slice 5: Business Domain Assignment
- [x] Add businessDomainId to ImportSession
- [x] After capability creation, assign L1 capabilities to domain
- [x] Include domain assignments in result summary

### Slice 6: Frontend Import UI
- [x] Add "Import" button to Architecture Canvas toolbar
- [x] Create ImportDialog component with step-based workflow
- [x] Implement ImportUploadStep (file input + domain dropdown)
- [x] Implement ImportPreviewStep (supported/unsupported counts)
- [x] Implement ImportProgressStep (progress bar + phase)
- [x] Implement ImportResultsStep (summary + error list)
- [x] Add polling mechanism for progress updates (2-second interval)
- [x] Use local state only (no global store needed)

### Slice 7: Cancel Import
- [x] Implement DELETE /api/v1/imports/{id} endpoint
- [x] Only allow cancel when status is "pending"
- [x] Add cancel button to frontend preview dialog

## Checklist
- [x] Specification approved
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [x] User sign-off
