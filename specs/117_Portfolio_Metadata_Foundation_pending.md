# Portfolio Metadata Foundation

**Status**: pending

**Series**: Application Landscape & TIME (1 of 4)
- **Spec 117**: Portfolio Metadata Foundation (this spec)
- Spec 118: Pillar Fit Type & TIME Suggestions
- Spec 119: TIME Classification & Application Landscape
- Spec 120: Portfolio Rationalization Analysis

## User Value

> "As an enterprise architect, I want to link applications to the M&A entity they came from, the vendor who sold them, or the team who built them, so I can analyze our portfolio by origin and track acquisition integration."

> "As a portfolio manager, I want to see all applications that came from a specific acquisition, so I can assess integration progress and identify technical debt from M&A."

> "As a domain architect, I want to be identified as the owner of my business domain, so stakeholders know who to contact for domain-specific decisions."

> "As an architect, I want to express that an application was both acquired via M&A AND built by the acquired company's team, so I can accurately model complex origin scenarios."

## Dependencies

- Architecture Modeling context (existing)
- Business Domains (existing, Capability Mapping)

---

## Domain Concepts

### Origin Entities

Three new aggregate types in Architecture Modeling, each representing a source of applications:

**AcquiredEntity** - A company or business unit acquired through M&A
- Examples: "TechCorp", "AcmeCo", "DataInc Merger 2021"
- Properties: name, acquisition date, integration status

**Vendor** - An external company from whom software was purchased
- Examples: "SAP", "Microsoft", "Salesforce"
- Properties: name, implementation partner

**InternalTeam** - An internal team or department that builds software
- Examples: "Platform Engineering", "Finance IT", "Data Team"
- Properties: name, department, contact person

### Origin Relationships

Typed relationships connecting origin entities to application components:

| Source Entity | Relationship Type | Target |
|---------------|-------------------|--------|
| AcquiredEntity | **Acquired via** | ApplicationComponent |
| Vendor | **Purchased from** | ApplicationComponent |
| InternalTeam | **Built by** | ApplicationComponent |

**Key design decisions:**
- An application can have **multiple origin relationships** (e.g., acquired via TechCorp AND built by TechCorp Engineering)
- Each relationship type is a **separate aggregate** to allow type-specific metadata in future
- Relationships are created by drawing lines in the canvas (like ComponentRelation)
- The relationship type is **automatically determined** by the source entity type

### Domain Architect

Business domains have an assigned domain architect who owns the capabilities within that domain.

**Selection**: UI shows dropdown of users with role Architect or Admin.

**Storage**: Stores the user's ID (UUID string).

**Display**: Resolve user ID to name via user service. If user no longer exists, display "Unknown user".

---

## Data Model

### AcquiredEntity (Architecture Modeling - new aggregate)

```
id: AcquiredEntityId
name: string (max 100 chars)
acquisitionDate: date | null
integrationStatus: NOT_STARTED | IN_PROGRESS | COMPLETED | null
notes: string (max 500 chars) | null
```

### Vendor (Architecture Modeling - new aggregate)

```
id: VendorId
name: string (max 100 chars)
implementationPartner: string (max 100 chars) | null
notes: string (max 500 chars) | null
```

### InternalTeam (Architecture Modeling - new aggregate)

```
id: InternalTeamId
name: string (max 100 chars)
department: string (max 100 chars) | null
contactPerson: string (max 100 chars) | null
notes: string (max 500 chars) | null
```

### AcquiredViaRelationship (Architecture Modeling - new aggregate)

```
id: AcquiredViaRelationshipId
acquiredEntityId: AcquiredEntityId
componentId: ComponentId
notes: string (max 500 chars) | null
```

Uniqueness: One relationship per (acquiredEntityId, componentId) pair.

### PurchasedFromRelationship (Architecture Modeling - new aggregate)

```
id: PurchasedFromRelationshipId
vendorId: VendorId
componentId: ComponentId
notes: string (max 500 chars) | null
```

Uniqueness: One relationship per (vendorId, componentId) pair.

### BuiltByRelationship (Architecture Modeling - new aggregate)

```
id: BuiltByRelationshipId
internalTeamId: InternalTeamId
componentId: ComponentId
notes: string (max 500 chars) | null
```

Uniqueness: One relationship per (internalTeamId, componentId) pair.

### Domain Architect (Capability Mapping extension)

Add to BusinessDomain:
```
domainArchitectId: UUID | null
```

---

## User Experience

### Origin Entities in Tree/Canvas

Origin entities appear in the Architecture Modeling tree and can be placed on the canvas:

```
Architecture Modeling
├─ Applications
│   ├─ SAP HR
│   ├─ Legacy Payroll
│   └─ CRM System
├─ Acquired Entities
│   ├─ TechCorp (2021)
│   └─ AcmeCo (2019)
├─ Vendors
│   ├─ SAP
│   ├─ Microsoft
│   └─ Salesforce
└─ Internal Teams
    ├─ Platform Engineering
    ├─ Finance IT
    └─ Data Team
```

### Drawing Origin Relationships

On the canvas, user draws a line from an origin entity to an application component:

```
┌─────────────────┐                    ┌─────────────────┐
│  TechCorp       │───Acquired via────▶│  SAP HR         │
│  (Acquired)     │                    │  (Application)  │
└─────────────────┘                    └─────────────────┘

┌─────────────────┐
│  TechCorp Eng   │───Built by────────▶│  SAP HR         │
│  (Internal Team)│                    │  (Application)  │
└─────────────────┘                    └─────────────────┘
```

The relationship type is automatically determined:
- Line from AcquiredEntity → "Acquired via"
- Line from Vendor → "Purchased from"
- Line from InternalTeam → "Built by"

### Origin Entity Detail Panel

When viewing/editing an AcquiredEntity:

```
┌─────────────────────────────────────────────────────────────────┐
│  TechCorp (Acquired Entity)                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Name: [TechCorp                               ]                │
│  Acquisition Date: [2021-03-15]                                 │
│  Integration Status: [IN_PROGRESS ▼]                            │
│  Notes: [Cloud infrastructure company acquired for...]          │
│                                                                  │
│  Applications acquired via TechCorp: 12                          │
│  ├─ SAP HR                                                       │
│  ├─ Cloud Platform                                               │
│  └─ ... (link to filtered view)                                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Application Component - Origins Section

When viewing an application, show its origin relationships:

```
┌─────────────────────────────────────────────────────────────────┐
│  SAP HR (Application)                                            │
│  Type: Application · Status: Active                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Origins                                                         │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Acquired via: TechCorp (2021)                           │    │
│  │ Integration: In Progress                                │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ Built by: TechCorp Engineering                          │    │
│  │ Department: Cloud Platform                              │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Strategic Fit Scores                                            │
│  ...                                                             │
└─────────────────────────────────────────────────────────────────┘
```

### Domain Architect

In Business Domain detail/edit:

```
┌─────────────────────────────────────────────────────────────────┐
│  Finance (Business Domain)                                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Domain Architect: [Alice Smith ▼]                              │
│                    ├─ Alice Smith                               │
│                    ├─ Bob Johnson                               │
│                    ├─ Carol Williams                            │
│                    └─ (None)                                    │
│                                                                  │
│  L1 Capabilities in this domain: 12                              │
│  ...                                                             │
└─────────────────────────────────────────────────────────────────┘
```

- Dropdown populated from `GET /users?role=Architect,Admin`
- On selection: stores the user ID
- Display: resolved user name, or "Unknown user" if deleted

---

## API Requirements

All endpoints follow REST Level 3 with HATEOAS. Responses include `_links` for navigation.

### Acquired Entities
- `POST /acquired-entities` - Create
- `GET /acquired-entities` - List all
- `GET /acquired-entities/{id}` - Get details
- `PUT /acquired-entities/{id}` - Update
- `DELETE /acquired-entities/{id}` - Delete

### Vendors
- `POST /vendors` - Create
- `GET /vendors` - List all
- `GET /vendors/{id}` - Get details
- `PUT /vendors/{id}` - Update
- `DELETE /vendors/{id}` - Delete

### Internal Teams
- `POST /internal-teams` - Create
- `GET /internal-teams` - List all
- `GET /internal-teams/{id}` - Get details
- `PUT /internal-teams/{id}` - Update
- `DELETE /internal-teams/{id}` - Delete

### Acquisition Relationships
- `POST /acquisition-relationships` - Create (body: acquiredEntityId, componentId, notes)
- `GET /acquisition-relationships/by-component/{componentId}` - Get by component
- `GET /acquisition-relationships/by-entity/{acquiredEntityId}` - Get by acquired entity
- `DELETE /acquisition-relationships/{id}` - Delete

### Vendor Relationships
- `POST /vendor-relationships` - Create (body: vendorId, componentId, notes)
- `GET /vendor-relationships/by-component/{componentId}` - Get by component
- `GET /vendor-relationships/by-vendor/{vendorId}` - Get by vendor
- `DELETE /vendor-relationships/{id}` - Delete

### Team Relationships
- `POST /team-relationships` - Create (body: internalTeamId, componentId, notes)
- `GET /team-relationships/by-component/{componentId}` - Get by component
- `GET /team-relationships/by-team/{internalTeamId}` - Get by team
- `DELETE /team-relationships/{id}` - Delete

### Component Origins (convenience)
- `GET /components/{id}/origins` - Get all origin relationships for a component
  ```json
  {
    "acquisitions": [{ "id": "...", "acquiredEntity": { "id": "...", "name": "TechCorp" }, "notes": "..." }],
    "vendors": [{ "id": "...", "vendor": { "id": "...", "name": "SAP" }, "notes": "..." }],
    "teams": [{ "id": "...", "team": { "id": "...", "name": "Platform Eng" }, "notes": "..." }],
    "_links": { ... }
  }
  ```

### Domain Architect (Capability Mapping)
- `PUT /business-domains/{id}` - Update domain including domainArchitectId
- `GET /business-domains/{id}` - Returns domainArchitectId with resolved user name (or "Unknown user" if deleted)

### Users for Domain Architect Selection
- `GET /users?roles=Architect,Admin` - Users eligible for domain architect assignment
  - Returns: `[{ id, name, role }]`

---

## Events

### New Events (Architecture Modeling)

**Acquired Entities:**
- `AcquiredEntityCreated`
- `AcquiredEntityUpdated`
- `AcquiredEntityDeleted`

**Vendors:**
- `VendorCreated`
- `VendorUpdated`
- `VendorDeleted`

**Internal Teams:**
- `InternalTeamCreated`
- `InternalTeamUpdated`
- `InternalTeamDeleted`

**Origin Relationships:**
- `AcquiredViaRelationshipCreated`
- `AcquiredViaRelationshipDeleted`
- `PurchasedFromRelationshipCreated`
- `PurchasedFromRelationshipDeleted`
- `BuiltByRelationshipCreated`
- `BuiltByRelationshipDeleted`

### Extended Events (Capability Mapping)
- `BusinessDomainUpdated` - Now includes domainArchitectId changes

---

## Cascade Deletion

When an origin entity is deleted:
- All relationships to that entity are deleted
- Events emitted for each deleted relationship

When an ApplicationComponent is deleted:
- All origin relationships to that component are deleted (handled by existing cascade logic)

---

## Checklist

- [ ] Specification approved

**Acquired Entities:**
- [ ] AcquiredEntity aggregate (domain model, events)
- [ ] AcquiredEntity CRUD command handlers
- [ ] AcquiredEntity API endpoints
- [ ] AcquiredEntity UI (tree, detail panel, create/edit)

**Vendors:**
- [ ] Vendor aggregate (domain model, events)
- [ ] Vendor CRUD command handlers
- [ ] Vendor API endpoints
- [ ] Vendor UI (tree, detail panel, create/edit)

**Internal Teams:**
- [ ] InternalTeam aggregate (domain model, events)
- [ ] InternalTeam CRUD command handlers
- [ ] InternalTeam API endpoints
- [ ] InternalTeam UI (tree, detail panel, create/edit)

**Origin Relationships:**
- [ ] AcquiredViaRelationship aggregate (domain model, events)
- [ ] PurchasedFromRelationship aggregate (domain model, events)
- [ ] BuiltByRelationship aggregate (domain model, events)
- [ ] Origin relationship command handlers
- [ ] Origin relationship API endpoints (path-based filtering)
- [ ] Canvas: draw relationships from origin entities to components
- [ ] Auto-detect relationship type based on source entity
- [ ] Component detail panel: show origins section
- [ ] Component origins convenience endpoint

**Domain Architect:**
- [ ] Domain architect field (domain model, events)
- [ ] Domain architect API (update business domain endpoint)
- [ ] Users query API (filter by Architect/Admin role)
- [ ] Domain architect UI (dropdown from eligible users)

**General:**
- [ ] Cascade deletion for origin entities
- [ ] Test-data added to scripts\seed-test-data.ts
- [ ] Tests passing
- [ ] User sign-off
