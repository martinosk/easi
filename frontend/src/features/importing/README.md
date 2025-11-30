# Import Feature

This feature provides a complete system for importing ArchiMate Open Exchange files into the application.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                       FRONTEND                                          │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│  ┌──────────────┐     ┌─────────────────────────────────────────────────────────────┐  │
│  │ ImportButton │────▶│                    ImportDialog                              │  │
│  └──────────────┘     │  ┌─────────────┬──────────────┬────────────┬─────────────┐  │  │
│                       │  │ UploadStep  │ PreviewStep  │ Progress   │ ResultsStep │  │  │
│                       │  │             │              │ Step       │             │  │  │
│                       │  └─────────────┴──────────────┴────────────┴─────────────┘  │  │
│                       └─────────────────────────────────────────────────────────────┘  │
│                                            │                                            │
│                              ┌─────────────▼──────────────┐                            │
│                              │    useImportSession Hook   │                            │
│                              │  - createSession()         │                            │
│                              │  - confirmSession()        │                            │
│                              │  - cancelSession()         │                            │
│                              │  - auto-polling (2s)       │                            │
│                              └─────────────┬──────────────┘                            │
└────────────────────────────────────────────│────────────────────────────────────────────┘
                                             │ REST API
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                          BACKEND - Importing Bounded Context                            │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│  ┌────────────────────────────────── API Layer ─────────────────────────────────────┐  │
│  │                                                                                   │  │
│  │  POST /imports ─────────▶ CreateImportSession    (upload + parse XML)            │  │
│  │  GET  /imports/{id} ────▶ GetImportSession       (poll status/progress)          │  │
│  │  POST /imports/{id}/confirm ▶ ConfirmImport      (start import)                  │  │
│  │  DELETE /imports/{id} ──▶ CancelImport           (cancel pending)                │  │
│  │                                                                                   │  │
│  └───────────────────────────────────┬───────────────────────────────────────────────┘  │
│                                      │                                                  │
│  ┌───────────────────────────────────▼──────────────────────────────────────────────┐  │
│  │                              Application Layer                                    │  │
│  │                                                                                   │  │
│  │  ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────────────────┐ │  │
│  │  │    Commands     │     │    Handlers     │     │      Orchestrator           │ │  │
│  │  │                 │     │                 │     │                             │ │  │
│  │  │ CreateImport    │────▶│ CreateImport    │     │  Execute():                 │ │  │
│  │  │ Session         │     │ SessionHandler  │     │  1. Create Components       │ │  │
│  │  │                 │     │                 │     │  2. Create Capabilities     │ │  │
│  │  │ ConfirmImport   │────▶│ ConfirmImport   │────▶│  3. Create Realizations     │ │  │
│  │  │                 │     │ Handler         │     │  4. Create Component Rels   │ │  │
│  │  │ CancelImport    │────▶│ CancelImport    │     │  5. Assign to Domain        │ │  │
│  │  │                 │     │ Handler         │     │                             │ │  │
│  │  └─────────────────┘     └─────────────────┘     └──────────────┬──────────────┘ │  │
│  │                                                                  │                │  │
│  │  ┌─────────────────┐     ┌─────────────────┐                    │                │  │
│  │  │ ArchiMate       │     │   Projector     │◀───Events──────────┘                │  │
│  │  │ Parser          │     │                 │                                     │  │
│  │  │                 │     │ Projects events │                                     │  │
│  │  │ Parses XML into │     │ to ReadModel    │───────────────────────┐             │  │
│  │  │ ParseResult     │     │                 │                       │             │  │
│  │  └─────────────────┘     └─────────────────┘                       ▼             │  │
│  │                                                          ┌─────────────────┐     │  │
│  │                                                          │   ReadModel     │     │  │
│  │                                                          │                 │     │  │
│  │                                                          │ ImportSessionDTO│     │  │
│  │                                                          │ (query side)    │     │  │
│  │                                                          └─────────────────┘     │  │
│  └───────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                         │
│  ┌──────────────────────────────── Domain Layer ────────────────────────────────────┐  │
│  │                                                                                   │  │
│  │  ┌─────────────────────────────────────┐    ┌──────────────────────────────────┐ │  │
│  │  │        ImportSession Aggregate      │    │          Domain Events           │ │  │
│  │  │                                     │    │                                  │ │  │
│  │  │  - ID: ImportSessionId              │    │  • ImportSessionCreated          │ │  │
│  │  │  - Status: ImportStatus             │    │  • ImportStarted                 │ │  │
│  │  │  - SourceFormat: SourceFormat       │    │  • ImportProgressUpdated         │ │  │
│  │  │  - BusinessDomainID: string         │    │  • ImportCompleted               │ │  │
│  │  │  - ParsedData: ParsedData           │    │  • ImportFailed                  │ │  │
│  │  │  - Progress: ImportProgress         │    │  • ImportSessionCancelled        │ │  │
│  │  │  - Result: ImportResult             │    │                                  │ │  │
│  │  │                                     │    └──────────────────────────────────┘ │  │
│  │  │  Methods:                           │                                         │  │
│  │  │  - Start()                          │    ┌──────────────────────────────────┐ │  │
│  │  │  - UpdateProgress()                 │    │         Value Objects            │ │  │
│  │  │  - Complete()                       │    │                                  │ │  │
│  │  │  - Fail()                           │    │  • ImportSessionId               │ │  │
│  │  │  - Cancel()                         │    │  • SourceFormat                  │ │  │
│  │  └─────────────────────────────────────┘    │  • ImportStatus                  │ │  │
│  │                                             │  • ImportPreview                 │ │  │
│  │                                             │  • ImportProgress                │ │  │
│  │                                             │  • ImportError                   │ │  │
│  │                                             └──────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                         │
│  ┌─────────────────────────────── Infrastructure ───────────────────────────────────┐  │
│  │                                                                                   │  │
│  │  ┌─────────────────────────────────────┐    ┌──────────────────────────────────┐ │  │
│  │  │   ImportSessionRepository           │    │         Event Store              │ │  │
│  │  │   (Event-Sourced)                   │───▶│                                  │ │  │
│  │  │                                     │    │   Persists domain events         │ │  │
│  │  │  - Save(session)                    │    │   Rebuilds aggregate state       │ │  │
│  │  │  - GetByID(id)                      │    │                                  │ │  │
│  │  └─────────────────────────────────────┘    └──────────────────────────────────┘ │  │
│  │                                                                                   │  │
│  └───────────────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                             │
              ┌──────────────────────────────┴──────────────────────────────┐
              │                      Command Bus                            │
              ▼                                                             ▼
┌─────────────────────────────────┐                   ┌─────────────────────────────────┐
│  capabilitymapping Context      │                   │  architecturemodeling Context   │
│                                 │                   │                                 │
│  • CreateCapability             │                   │  • CreateApplicationComponent   │
│  • LinkSystemToCapability       │                   │  • CreateComponentRelation      │
│  • AssignCapabilityToDomain     │                   │                                 │
└─────────────────────────────────┘                   └─────────────────────────────────┘
```

## Data Flow

### Import Session Lifecycle

```
┌────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  UPLOAD    │────▶│   PENDING   │────▶│  IMPORTING  │────▶│  COMPLETED  │
│            │     │             │     │             │     │             │
│ Parse XML  │     │ Show        │     │ Execute     │     │ Show        │
│ Validate   │     │ Preview     │     │ Phases      │     │ Results     │
│ Create     │     │ Confirm?    │     │ Track       │     │             │
│ Session    │     │             │     │ Progress    │     │             │
└────────────┘     └──────┬──────┘     └─────────────┘     └─────────────┘
                         │                                        │
                         │ Cancel                                 │
                         ▼                                        ▼
                  ┌─────────────┐                          ┌─────────────┐
                  │  CANCELLED  │                          │   FAILED    │
                  └─────────────┘                          └─────────────┘
```

### Import Execution Phases

```
1. Creating Components
   └── ArchiMate ApplicationComponent → CreateApplicationComponent command

2. Creating Capabilities
   └── ArchiMate Capability → CreateCapability command (with hierarchy L1-L4)

3. Creating Realizations
   └── ArchiMate Realization relationship → LinkSystemToCapability command

4. Creating Component Relations
   └── ArchiMate Triggering/Serving → CreateComponentRelation command

5. Assigning to Domain (optional)
   └── L1 Capabilities → AssignCapabilityToDomain command
```