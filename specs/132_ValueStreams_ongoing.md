# Value Streams

## Status
**ongoing** — Slice 1 & 2 complete, Slices 3–4 pending

---

## Problem Statement

Enterprise architects need to model how their organization delivers value end-to-end. Today, EASI captures what the business *can* do (capabilities) and what IT systems exist (components), but there is no way to express the *flow* through which these capabilities combine to produce business outcomes.

Without value streams, architects cannot answer questions like:
- "Which capabilities are involved in onboarding a new customer, and in what sequence?"
- "If we improve the Payment Processing capability, which value streams benefit?"
- "Which stages in our value delivery have no IT system support (gaps)?"

Value streams bridge the gap between the static capability map and dynamic business value delivery. They give stakeholders a shared language for discussing cross-cutting business processes from the customer's (or stakeholder's) perspective.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Define value streams, compose them from capabilities, analyze coverage and gaps |
| **Business Analyst** | Understand which capabilities participate in which value streams, assess impact of changes |
| **Technology Lead** | Trace from a value stream stage back to the realizing IT systems via capabilities |

---

## Core Concepts

### Value Stream
A value stream is a named, ordered sequence of stages that together deliver a specific business outcome. Examples: "Customer Onboarding", "Order-to-Cash", "Employee Hiring".

A value stream exists independently of capabilities -- it describes *what value is delivered and through what stages*, not *how* it is delivered. Capabilities are then mapped to stages to show the "how".

### Value Stream Stage
A stage is one step within a value stream. Stages have a defined order (sequence position) and represent a meaningful phase in the value delivery. Example stages within "Order-to-Cash": "Order Placement", "Payment Processing", "Fulfillment", "Delivery", "Post-Sale Support".

### Stage-Capability Mapping
A stage can be realized by one or more capabilities. A capability can participate in multiple stages across different value streams. This many-to-many relationship is the key analytical link -- it connects the "flow" view (value streams) to the "inventory" view (capability map).

---

## Business Rules & Invariants

### Value Stream Rules
1. **Name required and unique** -- Value stream names must be non-empty and unique within the tenant (max 100 characters).
2. **Description optional** -- Free-text description (max 500 characters).
3. **At least one stage** -- A value stream must have at least one stage to be meaningful, but the system allows creating a value stream with zero stages (stages are added incrementally).
4. **Independent lifecycle** -- Creating, updating, or deleting a value stream has no effect on the capabilities themselves.

### Stage Rules
5. **Name required** -- Stage names must be non-empty (max 100 characters).
6. **Description optional** -- Free-text description (max 500 characters).
7. **Ordered sequence** -- Each stage has a position (1-based integer) within its value stream. Positions must be contiguous (no gaps).
8. **Unique name within value stream** -- No two stages in the same value stream may share the same name.
9. **Reordering** -- Stages can be reordered. Reordering updates position values for all affected stages to maintain contiguity.

### Stage-Capability Mapping Rules
10. **Capability must exist** -- Cannot map a non-existent capability to a stage.
11. **No duplicates** -- The same capability cannot be mapped to the same stage twice.
12. **Any level allowed** -- Capabilities of any level (L1-L4) can be mapped to stages. The appropriate level depends on the granularity the architect is modeling.
13. **Cascade on stage deletion** -- When a stage is deleted, all its capability mappings are removed.
14. **Cascade on capability deletion** -- When a capability is deleted (from the Capability Mapping context), all mappings referencing it must be removed.
15. **Cascade on value stream deletion** -- When a value stream is deleted, all its stages and their capability mappings are removed.

### Cross-Context Rules
16. **Read-only consumption of capabilities** -- Value streams reference capabilities by ID only. The Value Stream context does not modify capabilities.
17. **Event-driven consistency** -- When a capability is deleted in Capability Mapping, the Value Stream context must react by removing all references to that capability.

---

## User Stories (Vertical Slices)

### Slice 1: Create and List Value Streams

**User Need:** As an enterprise architect, I need to define value streams so I can start modeling how my organization delivers value.

**Acceptance Criteria:**
- [x] User can navigate to a "Value Streams" page via the top navigation bar
- [x] The page displays a list of all value streams for the tenant (name, description, stage count, creation date)
- [x] User can create a new value stream by providing a name and optional description
- [x] Duplicate names are rejected with a clear error message
- [x] Empty state is shown when no value streams exist, with guidance to create the first one
- [x] Each value stream in the list is clickable to navigate to its detail view
- [x] User can edit the name and description of an existing value stream
- [x] User can delete a value stream (with confirmation dialog)
- [x] Deleting a value stream removes all its stages and mappings

**Edge Cases:**
- Deleting a value stream that has stages with mapped capabilities triggers cascade deletion
- Value stream names are trimmed of leading/trailing whitespace
- Maximum of ~200 value streams per tenant (no explicit limit enforced, but performance assumed for this scale)

---

### Slice 2: Interactive Flow Diagram — Stage & Capability Modeling

**User Need:** As an enterprise architect, I need a visual, interactive flow diagram where I can define stages, map capabilities to them, and see the complete value stream — all in one place.

The flow diagram is the primary modeling surface for value streams. It combines stage management, capability mapping, and visualization into a single interactive experience.

**Acceptance Criteria — Stage Management:**
- [x] Value stream detail page renders stages as columns in a horizontal left-to-right flow
- [x] User can add a new stage via a "+" button at the end of the flow (appended at the end by default)
- [x] User can insert a stage between existing stages (via "+" insert point between columns)
- [x] User can reorder stages via drag-and-drop of stage columns
- [x] User can edit a stage name and description inline (click to edit) or via context menu
- [x] User can delete a stage via context menu (with confirmation if it has capability mappings)
- [x] Deleting a stage automatically adjusts positions of subsequent stages
- [x] Each stage column displays its name, description (on hover/expand), and mapped capabilities

**Acceptance Criteria — Capability Mapping:**
- [x] A capability panel is available on the side of the flow diagram, showing the full L1-L4 capability tree
- [x] User can drag capabilities from the panel onto a stage column to create a mapping
- [x] User can remove a capability mapping by clicking a remove action on the capability chip within a stage
- [x] Already-mapped capabilities are visually indicated in the capability panel (e.g., dimmed or badged with stage count)
- [x] A capability can be mapped to multiple stages (same or different value streams)
- [x] The capability name and level (L1-L4) are displayed as chips/cards within the stage column

**Acceptance Criteria — Visualization & Analysis:**
- [ ] Capabilities with no realization (no IT system linked) are visually highlighted as gaps (e.g., warning icon or color) *(deferred — requires cross-context enterprise architecture data)*
- [ ] Clicking a capability chip navigates to or shows capability details *(deferred to Slice 4)*
- [x] The flow is responsive to the number of stages (horizontal scroll for many stages)
- [x] A summary bar shows: total stages, total unique capabilities mapped *(gap count deferred — requires cross-context data)*
- [x] Empty state (zero stages) prompts the user to add their first stage

**Edge Cases:**
- Reordering a single stage to the same position is a no-op
- Deleting the only stage in a value stream is allowed (value stream becomes empty)
- Stage names must be unique within the value stream
- Attempting to map the same capability to the same stage twice is rejected (visual feedback)
- If a capability is deleted in the Capability Mapping context, it disappears from all stages (handled via cross-context event)
- Stages with many capabilities (10+) should remain usable (scrollable within the column)
- Very long stage names are truncated with tooltip

---

### Slice 3: Value Stream in Navigation Tree Sidebar

**User Need:** As an enterprise architect, I want to see value streams in the sidebar explorer so I can quickly navigate between value streams alongside other artifacts.

**Acceptance Criteria:**
- [ ] The navigation tree sidebar includes a "Value Streams" section (collapsible, like Capabilities, Views, etc.)
- [ ] Value streams are listed by name, sorted alphabetically
- [ ] Clicking a value stream in the sidebar navigates to its detail page
- [ ] The count of value streams is shown in the section header
- [ ] Right-click context menu provides "Edit" and "Delete" actions
- [ ] A "+" button in the section header allows creating a new value stream

**Edge Cases:**
- If value streams feature is empty, the section shows "No value streams" empty state
- Sidebar section is collapsed by default to avoid cluttering for users who have not adopted value streams yet

---

### Slice 4: Cross-Capability Impact Analysis

**User Need:** As an enterprise architect, I need to see which value streams a given capability participates in, so I can assess the impact of changing or retiring a capability.

**Acceptance Criteria:**
- [ ] The capability detail panel (existing) shows a "Value Streams" section listing all value streams and stages where this capability is mapped
- [ ] Each entry shows: value stream name, stage name
- [ ] Clicking a value stream name navigates to the value stream detail page
- [ ] If a capability is not mapped to any value stream, the section shows "Not part of any value stream"

**Edge Cases:**
- A capability mapped to multiple stages in the same value stream appears once per stage
- This section loads independently and does not block the rest of the capability detail panel

---

## Bounded Context Considerations

Value streams introduce a new concern that touches capabilities but is conceptually distinct. Two reasonable options:

**Option A: New "ValueStream" bounded context** -- Clean separation. Value streams own their lifecycle, stages, and mappings. They consume capability data read-only and listen for capability deletion events. This follows the same pattern as Enterprise Architecture consuming capabilities.

**Option B: Extend Capability Mapping context** -- Value streams are closely tied to capabilities and could be considered part of the same strategic mapping activity. This avoids cross-context overhead for what is essentially an additional "view" over capabilities.

**Recommendation:** Option A (new bounded context) aligns better with the project's DDD philosophy. Value streams have their own aggregate lifecycle, their own invariants, and their own UI. The coupling to capabilities is limited to read-only references and a single deletion event. This is the same integration pattern used by Enterprise Architecture.

---

## Design Decisions

1. **No sub-stages.** Stages are a flat ordered list. The capability hierarchy (L1-L4) already provides granularity — architects map fine-grained L3/L4 capabilities to stages instead of nesting stages.

2. **No stage types.** Stages have no phase classification (e.g., "Triggering", "Value-Adding"). They are simply ordered steps with a name and description.

3. **Binary capability mappings.** Stage-capability mappings have no contribution level. A capability either enables a stage or it doesn't.

4. **Cross-value-stream overlap via Slice 4.** The capability detail panel (Slice 4) shows which value streams a capability participates in.

5. **Dedicated permission model.** Value stream management uses its own `valuestreams:write` permission, separate from `capabilities:write`. Value streams serve different stakeholders (process owners, CX teams) who may not manage the capability taxonomy.

6. **Interactive flow diagram as modeling surface.** The value stream detail page uses a purpose-built interactive flow diagram (Slice 2) as the primary modeling surface. This respects the sequential nature of value streams, which doesn't map to the free-form spatial canvas.

---

## Architecture

New bounded context: `backend/internal/valuestreams/`. Follows all existing conventions (standard folder layout, event sourcing, CQRS, HATEOAS). Reference `capabilitymapping` and `architectureviews` for patterns.

---

### Aggregate Design

**Single `ValueStream` aggregate root** containing `Stage` entities and `[]CapabilityRef` per stage.

```
ValueStream (Aggregate Root)
  |-- name:        ValueStreamName       (value object)
  |-- description: Description           (value object)
  |-- stages:      []Stage               (ordered entity collection)
  |-- createdAt, isDeleted

Stage (Entity)
  |-- id:              StageID           (UUID)
  |-- name:            StageName
  |-- description:     Description
  |-- position:        StagePosition     (1-based integer)
  |-- capabilityRefs:  []CapabilityRef   (capability ID wrapper)
```

**Why single aggregate:**
1. Stage position contiguity requires atomic cross-stage updates
2. Stages have no independent lifecycle
3. Cascade deletion is one operation
4. Precedent: `ArchitectureView` manages child memberships the same way
5. Event streams stay small (5-15 stages typical)

---

### Domain Events

All prefixed with `ValueStream`, all keyed on the value stream's aggregate ID.

| Event | Key Fields |
|---|---|
| `ValueStreamCreated` | `id`, `name`, `description` |
| `ValueStreamUpdated` | `id`, `name`, `description` |
| `ValueStreamDeleted` | `id` |
| `ValueStreamStageAdded` | `valueStreamId`, `stageId`, `name`, `position` |
| `ValueStreamStageUpdated` | `valueStreamId`, `stageId`, `name` |
| `ValueStreamStageRemoved` | `valueStreamId`, `stageId` |
| `ValueStreamStageReordered` | `valueStreamId`, `stageId`, `oldPosition`, `newPosition` |
| `ValueStreamStageCapabilityAdded` | `valueStreamId`, `stageId`, `capabilityId` |
| `ValueStreamStageCapabilityRemoved` | `valueStreamId`, `stageId`, `capabilityId` |

---

### API Endpoints

All relative to `/api/v1/`. Sub-resource nesting for stages and capabilities.

| Method | Path | Permission | Status |
|---|---|---|---|
| `GET` | `/value-streams` | `valuestreams:read` | 200 |
| `POST` | `/value-streams` | `valuestreams:write` | 201 |
| `GET` | `/value-streams/{id}` | `valuestreams:read` | 200 |
| `PUT` | `/value-streams/{id}` | `valuestreams:write` | 200 |
| `DELETE` | `/value-streams/{id}` | `valuestreams:delete` | 204 |
| `GET` | `/value-streams/{id}/capabilities` | `valuestreams:read` | 200 |
| `POST` | `/value-streams/{id}/stages` | `valuestreams:write` | 201 |
| `PUT` | `/value-streams/{id}/stages/{stageId}` | `valuestreams:write` | 200 |
| `DELETE` | `/value-streams/{id}/stages/{stageId}` | `valuestreams:delete` | 204 |
| `PATCH` | `/value-streams/{id}/stages/{stageId}/position` | `valuestreams:write` | 200 |
| `PUT` | `/value-streams/{id}/stages/positions` | `valuestreams:write` | 200 |
| `POST` | `/value-streams/{id}/stages/{stageId}/capabilities` | `valuestreams:write` | 201 |
| `DELETE` | `/value-streams/{id}/stages/{stageId}/capabilities/{capabilityId}` | `valuestreams:delete` | 204 |
| `GET` | `/capabilities/{id}/value-streams` | `valuestreams:read` | 200 |

`GET /value-streams/{id}/capabilities` returns all unique capabilities mapped across all stages in the value stream, with the stage(s) each capability belongs to. Supports the analytics use case without requiring clients to traverse stages individually.

`PUT /value-streams/{id}/stages/positions` accepts a batch reorder payload `{ "positions": [{ "stageId": "...", "position": 1 }, ...] }` for drag-and-drop reordering. The individual `PATCH .../position` endpoint remains for single-stage moves.

---

### HATEOAS Links

All links are permission-gated per actor.

**Value Stream resource:**

| Rel | Method | Path | Condition |
|---|---|---|---|
| `self` | GET | `/value-streams/{id}` | always |
| `collection` | GET | `/value-streams` | always |
| `edit` | PUT | `/value-streams/{id}` | `valuestreams:write` |
| `delete` | DELETE | `/value-streams/{id}` | `valuestreams:delete` |
| `x-stages` | GET | `/value-streams/{id}` (stages embedded) | always |
| `x-add-stage` | POST | `/value-streams/{id}/stages` | `valuestreams:write` |
| `x-capabilities` | GET | `/value-streams/{id}/capabilities` | always |
| `x-reorder` | PUT | `/value-streams/{id}/stages/positions` | `valuestreams:write` |

**Stage resource (embedded in value stream response):**

| Rel | Method | Path | Condition |
|---|---|---|---|
| `self` | GET | `/value-streams/{id}` (stage within response) | always |
| `up` | GET | `/value-streams/{id}` | always |
| `edit` | PUT | `/value-streams/{id}/stages/{stageId}` | `valuestreams:write` |
| `delete` | DELETE | `/value-streams/{id}/stages/{stageId}` | `valuestreams:delete` |
| `x-reposition` | PATCH | `/value-streams/{id}/stages/{stageId}/position` | `valuestreams:write` |
| `x-add-capability` | POST | `/value-streams/{id}/stages/{stageId}/capabilities` | `valuestreams:write` |

**Capability mapping (embedded in stage):**

| Rel | Method | Path | Condition |
|---|---|---|---|
| `delete` | DELETE | `.../stages/{stageId}/capabilities/{capabilityId}` | `valuestreams:delete` |
| `x-capability` | GET | `/capabilities/{capabilityId}` | `capabilities:read` |

---

### Error Responses

All endpoints return errors using the standard `ErrorResponse` format. Domain model owns all validation -- the API layer translates domain exceptions to HTTP status codes.

**Common error codes across all endpoints:**

| Status | Condition |
|---|---|
| 401 | Authentication required |
| 403 | Missing required permission |
| 409 | Concurrency conflict (optimistic lock on aggregate version) -- client should retry |
| 500 | Internal server error |

**Value stream endpoints:**

| Endpoint | Status | Condition |
|---|---|---|
| `POST /value-streams` | 400 | Name empty or exceeds 100 chars |
| `POST /value-streams` | 409 | Name already exists within tenant |
| `PUT /value-streams/{id}` | 400 | Name empty or exceeds 100 chars |
| `PUT /value-streams/{id}` | 404 | Value stream not found |
| `PUT /value-streams/{id}` | 409 | New name conflicts with existing |
| `DELETE /value-streams/{id}` | 404 | Value stream not found |

**Stage endpoints:**

| Endpoint | Status | Condition |
|---|---|---|
| `POST .../stages` | 400 | Name empty or exceeds 100 chars |
| `POST .../stages` | 404 | Value stream not found |
| `POST .../stages` | 409 | Stage name already exists in this value stream |
| `PUT .../stages/{stageId}` | 400 | Name empty or exceeds 100 chars |
| `PUT .../stages/{stageId}` | 404 | Value stream or stage not found |
| `PUT .../stages/{stageId}` | 409 | New name conflicts with existing stage |
| `DELETE .../stages/{stageId}` | 404 | Value stream or stage not found |
| `PATCH .../stages/{stageId}/position` | 400 | Invalid position |
| `PATCH .../stages/{stageId}/position` | 404 | Value stream or stage not found |
| `PUT .../stages/positions` | 400 | Missing stages, invalid positions, non-contiguous |
| `PUT .../stages/positions` | 404 | Value stream not found |

**Capability mapping endpoints:**

| Endpoint | Status | Condition |
|---|---|---|
| `POST .../capabilities` | 404 | Value stream, stage, or capability not found |
| `POST .../capabilities` | 409 | Capability already mapped to this stage |
| `DELETE .../capabilities/{capabilityId}` | 404 | Mapping not found |

---

### Permissions

| Role | Permissions |
|---|---|
| `admin` | `valuestreams:read`, `valuestreams:write`, `valuestreams:delete` |
| `architect` | `valuestreams:read`, `valuestreams:write`, `valuestreams:delete` |
| `stakeholder` | `valuestreams:read` |

**All three permission registries must be updated in sync:**
1. `auth/domain/valueobjects/permission.go` -- add `PermValueStreamsRead`, `PermValueStreamsWrite`, `PermValueStreamsDelete` constants to `validPermissions`
2. `auth/domain/valueobjects/role.go` -- add permissions to `rolePermissionsList` for admin, architect, stakeholder
3. `shared/context/actor_context.go` -- add permissions to `rolePermissions` map

All DELETE operations (`value-streams`, `stages`, `capability mappings`) require `valuestreams:delete`. This is consistent -- any removal of data from the value stream aggregate requires delete permission.

---

### Cross-Context Integration

| Source Event | Handler | Action |
|---|---|---|
| `CapabilityDeleted` | `OnCapabilityDeletedHandler` | Query read model for affected `(valueStreamId, stageId)` pairs, dispatch `RemoveStageCapability` commands (preserves audit trail) |
| `CapabilityUpdated` | `CapabilityNameUpdateProjector` | Update denormalized `capability_name`/`capability_level` in read model |

Read-only capability data via `CapabilityGateway` interface (anti-corruption layer querying the `capabilities` read model table).

**Partial failure handling:** The `OnCapabilityDeletedHandler` iterates over affected mappings and dispatches individual `RemoveStageCapability` commands. If an individual command fails (e.g., concurrency conflict), the handler logs a warning and continues processing remaining mappings. This follows the established pattern in `capabilitymapping/OnCapabilityDeletedHandler`. Orphaned references (stage mappings pointing to a deleted capability) are tolerable -- the read model projector should handle missing capabilities gracefully (e.g., display "Unknown capability" or omit). Consider adding a reconciliation health-check query that detects orphaned capability references for operational monitoring.

---

### Database Migration

Three read model tables, no foreign keys, RLS on all three. `value_stream_id` denormalized into `value_stream_stage_capabilities` for efficient queries. All business invariants (name uniqueness, stage name uniqueness, duplicate mapping prevention) are enforced exclusively by the domain model -- the database serves as a read model projection only.

| Table | PK | Notable Indexes |
|---|---|---|
| `value_streams` | `(tenant_id, id)` | Unique on `(tenant_id, name)` for query performance |
| `value_stream_stages` | `(tenant_id, id)` | On `(tenant_id, value_stream_id, position)` |
| `value_stream_stage_capabilities` | `(tenant_id, stage_id, capability_id)` | On `(tenant_id, capability_id)`, `(tenant_id, value_stream_id)` |

**RLS is mandatory on all three tables.** Each table must have `ROW LEVEL SECURITY` enabled with a `tenant_isolation_policy` using `USING (tenant_id = current_setting('app.current_tenant', true)::uuid)` and a matching `WITH CHECK` clause, scoped `TO easi_app`. This must be in the same migration file that creates the tables. Add an integration test verifying that querying with a different tenant context returns zero rows.

---

### Frontend

- New feature folder: `frontend/src/features/value-streams/` (standard layout: `api/`, `hooks/`, `components/`, `queryKeys.ts`, `mutationEffects.ts`)
- Branded types: `ValueStreamId`, `ValueStreamStageId` in `api/types.ts`
- Routes: `/value-streams` and `/value-streams/:valueStreamId`
- New `AppView`: `'value-streams'` with nav button between Business Domains and Enterprise Architecture
- Sidebar: `ValueStreamsSection.tsx` (flat list pattern like `VendorsSection.tsx`), wired into `NavigationTreeContent.tsx`
- Capability detail panel: add "Value Stream Participation" section using `GET /capabilities/{id}/value-streams`
- Cross-feature cache invalidation: add `valueStreamsQueryKeys.lists()` to `capabilitiesMutationEffects.delete`
- MSW handlers for tests

---

### Trade-offs

| Decision | Trade-off | Mitigation |
|---|---|---|
| Single aggregate for ValueStream + Stages | All stage ops load full event stream | Streams stay small (5-15 stages). Separate aggregate would need eventual consistency for position ordering. |
| Denormalized capability names in mapping table | Must sync on rename | Projector subscribes to `CapabilityUpdated`, proven pattern. |
| Command-based cleanup on capability deletion | More complex than direct RM delete | Preserves event-sourced audit trail. |
| Denormalized `value_stream_id` in stage capabilities | Slight redundancy | Enables efficient queries without joining through stages. |

---

## Checklist
- [x] Specification ready
- [ ] Implementation done *(Slice 1 & 2 complete; Slices 3–4 pending)*
- [x] Unit tests implemented and passing *(73+ tests: 56+ backend, 17+ frontend)*
- [ ] Integration tests implemented if relevant
- [ ] API Documentation updated in OpenAPI specification
- [ ] User sign-off

## Slice 1 Implementation Notes

**Backend** — new bounded context `backend/internal/valuestreams/`:
- Domain layer: `ValueStream` aggregate, value objects (`ValueStreamName`, `Description`, `ValueStreamID`), 3 domain events
- Application layer: 3 command handlers (create/update/delete), read model, projector
- Infrastructure layer: event-sourced repository, HATEOAS links, API handlers, routes, error registration
- Published language: event constants for cross-context integration
- Database: migration `094_add_value_streams.sql` with RLS
- Permissions: `valuestreams:read/write/delete` added to all 3 permission registries

**Frontend** — new feature folder `frontend/src/features/value-streams/`:
- API client, query keys, mutation effects, hooks (`useValueStreams`, `useValueStreamsQuery`, `useValueStream`)
- `ValueStreamsPage` with list, create/edit modal, delete confirmation, empty state
- Navigation: top nav button (between Business Domains and Enterprise Architecture), routing
- MSW handlers for tests

**Tests:**
- Backend: value object tests, aggregate tests, handler tests (36 tests)
- Frontend: hook tests (7 tests)

## Slice 2 Implementation Notes

**Backend** — extended `valuestreams` bounded context:

*Domain layer:*
- New value objects: `StageID`, `StageName`, `StagePosition`, `CapabilityRef`
- New entity: `Stage` (immutable with `With*` methods)
- 6 new domain events: `ValueStreamStageAdded`, `ValueStreamStageUpdated`, `ValueStreamStageRemoved`, `ValueStreamStagesReordered`, `ValueStreamStageCapabilityAdded`, `ValueStreamStageCapabilityRemoved`
- Extended `ValueStream` aggregate with stage management (AddStage, UpdateStage, RemoveStage, ReorderStages, AddCapabilityToStage, RemoveCapabilityFromStage)

*Application layer:*
- 6 new commands and handlers
- `CapabilityGateway` interface for cross-context capability existence check
- Expanded read model with stage/capability DTOs, CRUD methods, detail query
- Projector handles 6 new events with generic `unmarshalEvent` helper

*Infrastructure layer:*
- Migration `095_add_value_stream_stages.sql` (2 tables + RLS)
- `CapabilityGatewayImpl` querying capability read model
- 7 new API handlers in `stage_handlers.go`
- HATEOAS links for stages and capability mappings
- 6 new routes with permission middleware
- Error registration for stage/capability domain errors

*Published language:*
- 6 new event constants for cross-context integration

**Frontend** — expanded `value-streams` feature:

*Types & API:*
- `StageId` branded type, `ValueStreamStage`, `StageCapabilityMapping`, `ValueStreamDetail`
- 7 new API methods (addStage, updateStage, deleteStage, reorderStages, addStageCapability, removeStageCapability, getById returns detail)

*Hooks:*
- `useValueStreamDetail`, `useAddStage`, `useUpdateStage`, `useDeleteStage`, `useReorderStages`, `useAddStageCapability`, `useRemoveStageCapability`
- `useStageOperations` — extracted form/mutation logic for clean page component

*Components:*
- `ValueStreamDetailPage` — detail page with header, back nav, summary bar, form overlay
- `StageFlowDiagram` — horizontal flow with drag-and-drop stage reorder and capability drop
- `StageColumn` — individual stage card with capabilities, edit/delete actions
- `CapabilityChip` — capability badge with remove button
- `AddStageButton` — dashed "+" button
- `SummaryBar` — stage count and capability count
- `CapabilitySidebar` — filterable L1-L4 capability tree with drag support
- `StageFormOverlay` — extracted add/edit stage form

*Routing:*
- Added `VALUE_STREAM_DETAIL` route, nested routing in `ValueStreamsRouter`
- Value stream cards are clickable to navigate to detail

*MSW:*
- 7 new endpoint handlers for tests

*Mutation effects:*
- 6 new cache invalidation effects (addStage, updateStage, deleteStage, reorderStages, addStageCapability, removeStageCapability)

**Tests:**
- Backend: value object tests (StageID, StageName, StagePosition, CapabilityRef), aggregate stage tests (20 tests), handler tests (6 files with success/error cases)
- Frontend: hook tests (existing 7 tests updated), 975 total tests passing
