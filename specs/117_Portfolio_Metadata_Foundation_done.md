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

## Frontend Architecture Refactoring

This spec requires refactoring the frontend architecture to enable reusable patterns for all canvas entity types. Currently, features like details view, context menus, selection, and panning are duplicated per entity type.

### Problem Statement

The following features exist for ApplicationComponents and Capabilities but are missing or duplicated for origin entities:
- **Details view** when selecting in tree or canvas
- **Context menu in treeview** with edit/delete from model
- **Context menu in canvas** for nodes
- **Details view and context menu for edges** (origin relationships)
- **Tree-view indicator** showing if component is in active view
- **Pan to component** when selecting in treeview

### Architecture Patterns

#### 1. Entity Type Detection Utilities

Centralized utilities for entity type handling based on node ID prefixes are in `frontend/src/features/canvas/utils/nodeFactory.ts`:

```typescript
// Implemented in frontend/src/features/canvas/utils/nodeFactory.ts
export type OriginEntityType = 'acquired' | 'vendor' | 'team';

export function getOriginEntityTypeFromNodeId(nodeId: string): OriginEntityType | null {
  if (nodeId.startsWith('acq-')) return 'acquired';
  if (nodeId.startsWith('vendor-')) return 'vendor';
  if (nodeId.startsWith('team-')) return 'team';
  return null;
}

export function extractOriginEntityId(nodeId: string): string | null {
  const prefixes = ['acq-', 'vendor-', 'team-'];
  for (const prefix of prefixes) {
    if (nodeId.startsWith(prefix)) {
      return nodeId.slice(prefix.length);
    }
  }
  return null;
}
```

#### 2. Generic TreeItemList Component

Replace duplicated tree section internals with a generic component:

```typescript
// frontend/src/features/navigation/components/TreeItemList.tsx
interface TreeItemListProps<T extends { id: string; name: string }> {
  items: T[];
  selectedId: string | null;
  searchPlaceholder: string;
  emptyMessage: string;
  icon: string;

  // Optional features
  currentView?: View | null;
  isInView?: (item: T, view: View) => boolean;
  getCustomColor?: (item: T, view: View) => string | undefined;
  renderLabel?: (item: T) => React.ReactNode;

  // Callbacks
  onSelect?: (id: string) => void;
  onContextMenu: (e: React.MouseEvent, item: T) => void;

  // Drag-drop
  dragDataKey: string;
  isDraggable?: (item: T) => boolean;

  // Inline editing (optional)
  editingState?: EditingState | null;
  setEditingState?: (state: EditingState | null) => void;
  onRenameSubmit?: () => void;
  editInputRef?: React.RefObject<HTMLInputElement | null>;

  // Search fields (for filtering)
  searchFields: (keyof T)[];
}
```

Entity-specific sections become thin wrappers:

```typescript
// AcquiredEntitiesSection becomes:
<TreeItemList
  items={acquiredEntities}
  selectedId={selectedEntityId}
  searchPlaceholder="Search acquired entities..."
  emptyMessage="No acquired entities"
  icon="🏢"
  dragDataKey="acquiredEntityId"
  searchFields={['name', 'notes']}
  renderLabel={(e) => `${e.name}${formatYear(e.acquisitionDate)}`}
  onSelect={onEntitySelect}
  onContextMenu={onEntityContextMenu}
/>
```

#### 3. Unified Selection Routing in DetailSection

Extend `MainLayout.tsx` to route origin entity selections:

```typescript
function renderDetailContent(props: DetailSectionProps): React.ReactNode {
  const { selectedNodeId, selectedEdgeId, selectedCapabilityId } = props;

  if (selectedNodeId) {
    const entityType = getEntityType(selectedNodeId);
    switch (entityType) {
      case 'acquired':
        return <AcquiredEntityDetailsPanel entityId={getEntityId(selectedNodeId)} />;
      case 'vendor':
        return <VendorDetailsPanel entityId={getEntityId(selectedNodeId)} />;
      case 'team':
        return <InternalTeamDetailsPanel entityId={getEntityId(selectedNodeId)} />;
      case 'capability':
        return <CapabilityDetails onRemoveFromView={props.onRemoveCapabilityFromView} />;
      default:
        return <ComponentDetails onEdit={props.onEditComponent} onRemoveFromView={props.onRemoveFromView} />;
    }
  }

  if (selectedEdgeId) {
    if (selectedEdgeId.startsWith('origin-')) {
      return <OriginRelationshipDetails />;
    }
    if (isRealizationEdge(selectedEdgeId)) {
      return <RealizationDetails />;
    }
    if (isRelationEdge(selectedEdgeId)) {
      return <RelationDetails onEdit={props.onEditRelation} />;
    }
  }

  return null;
}
```

#### 4. Origin Entity Context Menus

Extend `TreeContextMenus.tsx` to support origin entities:

```typescript
// Add to types
interface OriginEntityContextMenuState {
  x: number;
  y: number;
  entity: AcquiredEntity | Vendor | InternalTeam;
  entityType: 'acquired' | 'vendor' | 'team';
}

// In useTreeContextMenus hook, add handlers for origin entities
const handleOriginEntityContextMenu = (
  e: React.MouseEvent,
  entity: AcquiredEntity | Vendor | InternalTeam,
  entityType: 'acquired' | 'vendor' | 'team'
) => { ... };

const getOriginEntityContextMenuItems = (menu: OriginEntityContextMenuState): ContextMenuItem[] => {
  const items: ContextMenuItem[] = [];
  if (hasLink(menu.entity, 'edit')) {
    items.push({ label: 'Edit', onClick: () => onEditOriginEntity(menu) });
  }
  if (hasLink(menu.entity, 'delete')) {
    items.push({ label: 'Delete from Model', onClick: () => onDeleteOriginEntity(menu), isDanger: true });
  }
  return items;
};
```

#### 5. Canvas Node Context Menu for Origin Entities

Extend `NodeContextMenu.tsx` to handle origin entity nodes:

```typescript
// NodeContextMenu already checks nodeType, extend the type:
type NodeType = 'component' | 'capability' | 'originEntity';

// For origin entities, only show "Delete from Model" (they don't have view membership)
if (menu.nodeType === 'originEntity') {
  if (canDeleteFromModel) {
    items.push({
      label: 'Delete from Model',
      onClick: () => onRequestDelete({
        type: 'origin-entity-from-model',
        id: menu.nodeId,
        name: menu.nodeName,
      }),
      isDanger: true,
    });
  }
  return items;
}
```

#### 6. Edge Context Menu for Origin Relationships

Extend `EdgeContextMenu.tsx` to handle origin relationship edges:

```typescript
// Add new edge type check
if (menu.edgeId.startsWith('origin-')) {
  if (!canDelete) return [];

  return [{
    label: 'Delete Relationship',
    onClick: () => {
      onRequestDelete({
        type: 'origin-relationship',
        id: menu.edgeId.replace('origin-', ''),
        name: menu.edgeName,
      });
      onClose();
    },
    isDanger: true,
  }];
}
```

#### 7. Selection and Panning for Origin Entities

Wire up selection in NavigationTree:

```typescript
// In NavigationTree, add onOriginEntitySelect callback
<AcquiredEntitiesSection
  onEntitySelect={(entityId) => {
    onOriginEntitySelect?.(`acq-${entityId}`);
  }}
/>
```

In the parent (App.tsx or MainLayout), handle panning:

```typescript
const handleOriginEntitySelect = (nodeId: string) => {
  appStore.selectNode(nodeId);
  canvasRef.current?.centerOnNode(nodeId);
};
```

### Migration Path

1. **Phase 1**: Create utility functions (`nodeFactory.ts`) ✅
2. **Phase 2**: Extend DetailSection routing for origin entities ✅
3. **Phase 3**: Wire up origin entity selection and panning ✅
4. **Phase 4**: Extend context menus (tree and canvas) ✅
5. **Phase 5**: Add origin relationship edge handling ✅
6. **Phase 6**: (Future) Refactor tree sections to use TreeItemList

Note: Phase 6 is optional for this spec but recommended for maintainability.

### Implementation Notes

**Files Modified:**
- `frontend/src/features/canvas/utils/nodeFactory.ts` - Entity type utilities
- `frontend/src/features/canvas/hooks/useContextMenu.ts` - Origin entity/relationship context menu support
- `frontend/src/features/canvas/hooks/useDeleteConfirmation.ts` - Origin entity/relationship deletion handlers
- `frontend/src/features/canvas/components/context-menus/NodeContextMenu.tsx` - Origin entity node context menu
- `frontend/src/features/canvas/components/context-menus/EdgeContextMenu.tsx` - Origin relationship edge context menu
- `frontend/src/features/navigation/hooks/useTreeContextMenus.ts` - Tree context menu for origin entities
- `frontend/src/features/navigation/components/sections/AcquiredEntitiesSection.tsx` - Selection and context menu
- `frontend/src/features/navigation/components/sections/VendorsSection.tsx` - Selection and context menu
- `frontend/src/features/navigation/components/sections/InternalTeamsSection.tsx` - Selection and context menu
- `frontend/src/layouts/MainLayout.tsx` - DetailSection routing
- `frontend/src/layouts/DockviewLayout.tsx` - DetailSection routing

---

## Checklist

- [x] Specification approved

**Frontend Architecture:**
- [x] Entity type detection utilities (`frontend/src/features/canvas/utils/nodeFactory.ts`)
- [x] DetailSection routing for origin entities
- [x] Origin entity selection and panning wired up
- [x] Tree context menu for origin entities
- [x] Canvas node context menu for origin entities
- [x] Edge context menu for origin relationships
- [x] Origin relationship details panel

**Acquired Entities:**
- [x] AcquiredEntity aggregate (domain model, events)
- [x] AcquiredEntity CRUD command handlers
- [x] AcquiredEntity API endpoints
- [x] AcquiredEntity UI (tree section exists)
- [x] AcquiredEntity detail panel integration (routing + panel)

**Vendors:**
- [x] Vendor aggregate (domain model, events)
- [x] Vendor CRUD command handlers
- [x] Vendor API endpoints
- [x] Vendor UI (tree section exists)
- [x] Vendor detail panel integration (routing + panel)

**Internal Teams:**
- [x] InternalTeam aggregate (domain model, events)
- [x] InternalTeam CRUD command handlers
- [x] InternalTeam API endpoints
- [x] InternalTeam UI (tree section exists)
- [x] InternalTeam detail panel integration (routing + panel)

**Origin Relationships:**
- [x] AcquiredViaRelationship aggregate (domain model, events)
- [x] PurchasedFromRelationship aggregate (domain model, events)
- [x] BuiltByRelationship aggregate (domain model, events)
- [x] Origin relationship command handlers
- [x] Origin relationship API endpoints (path-based filtering)
- [x] Canvas: draw relationships from origin entities to components
- [x] Auto-detect relationship type based on source entity
- [x] Component detail panel: show origins section
- [x] Component origins convenience endpoint

**Domain Architect:**
- [x] Domain architect field (domain model, events)
- [x] Domain architect API (update business domain endpoint)
- [x] Users query API (filter by Architect/Admin role)
- [x] Domain architect UI (dropdown from eligible users)

**General:**
- [x] Cascade deletion for origin entities
- [x] Test-data added to scripts\seed-test-data.ts
- [x] Tests passing
- [x] User sign-off
