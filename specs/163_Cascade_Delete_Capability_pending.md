# Cascade Delete Capability

## Description
Delete a capability and all its descendants (child capabilities and their realizations). The user sees an impact analysis before confirming. Applications that only realise capabilities within the deleted subtree can optionally be deleted. Applications realising capabilities outside the subtree are always retained.

## Purpose
Enable users to remove entire capability subtrees without manually deleting each child. Maintain referential integrity across bounded contexts through event-driven cleanup.

## Bounded Contexts Involved

### Capability Mapping (Owner)
- Owns the Capability aggregate, hierarchy, CapabilityRealization, CapabilityDependency, BusinessDomainAssignment, StrategyImportance
- Publishes `CapabilityDeleted`, `SystemRealizationDeleted`, `CapabilityDependencyDeleted` events
- Orchestrates the cascade delete within a single HTTP request

### Downstream Contexts (React to CapabilityDeleted)
- **Enterprise Architecture** — `DomainCapabilityMetadataProjector` removes cached metadata
- **View Layouts** — `CapabilityDeletedHandler` removes element positions
- **Access Delegation** — `ArtifactDeletionProjector` revokes edit grants
- **Value Streams** — `CapabilityProjector` + `CapabilityDeletedHandler` remove cache entries and stage references

### Architecture Modeling (Conditionally Affected)
- If user opts to delete realising applications, deletion commands cross into this context via loopback HTTP (same pattern as Arch Assistant)

## Business Rules

### Cascade Deletion
- Deleting a capability with descendants requires explicit `cascade: true` confirmation
- Descendants are deleted bottom-up (L4 → L3 → L2 → L1) so existing `CanDelete` checks pass naturally
- Each deleted capability emits its own `CapabilityDeleted` event, triggering all existing downstream handlers
- Realizations on deleted capabilities are removed before the capabilities themselves
- Dependencies involving any capability in the subtree are removed

### Application Deletion Rule
- Applications are only deleted if the user explicitly opts in (`deleteRealisingApplications: true`)
- An application is only eligible for deletion if ALL its realizations point to capabilities within the deletion set
- If an application realises ANY capability outside the deletion set, it is retained (only its realizations to the deleted capabilities are removed)
- Cross-context application deletion uses loopback HTTP to `DELETE /api/v1/components/{id}`, respecting RBAC

### HATEOAS / RBAC
- The `delete` and `x-delete-impact` links appear on capabilities only when `actor.CanDelete("capabilities")` is true
- Impact analysis endpoint (`GET`) is gated by `capabilities:read` — any reader can preview impact
- The `x-confirm-delete` link in the impact response is only emitted when the actor has delete permission
- If the actor lacks `components:delete` permission, application deletion gracefully skips those components

## API Design

### GET /capabilities/{id}/delete-impact

Impact analysis — returns what would be affected. Pure read, never modifies state.

**Route registration:** Read permission group in `registerCapabilityRoutes`

**Response — 200 OK:**
```json
{
  "capabilityId": "uuid",
  "capabilityName": "Customer Management",
  "hasDescendants": true,
  "affectedCapabilities": [
    {
      "id": "uuid",
      "name": "Customer Onboarding",
      "level": "L2",
      "parentId": "uuid",
      "_links": { "self": { "href": "/capabilities/{id}", "method": "GET" } }
    }
  ],
  "realizationsOnDeletedCapabilities": [
    {
      "id": "uuid",
      "componentId": "uuid",
      "componentName": "CRM System",
      "capabilityId": "uuid",
      "capabilityName": "Customer Onboarding",
      "realizationLevel": "Full",
      "origin": "Direct",
      "_links": {
        "self": { "href": "/capability-realizations/{id}", "method": "GET" },
        "x-component": { "href": "/components/{componentId}", "method": "GET" }
      }
    }
  ],
  "realizationsOnRetainedCapabilities": [
    {
      "id": "uuid",
      "componentId": "uuid",
      "componentName": "CRM System",
      "capabilityId": "uuid",
      "capabilityName": "Sales Tracking",
      "realizationLevel": "Full",
      "origin": "Direct"
    }
  ],
  "_links": {
    "self": { "href": "/capabilities/{id}/delete-impact", "method": "GET" },
    "x-capability": { "href": "/capabilities/{id}", "method": "GET" },
    "x-confirm-delete": { "href": "/capabilities/{id}", "method": "DELETE" }
  }
}
```

**Field semantics:**
- `affectedCapabilities`: All descendants (not including the target itself)
- `realizationsOnDeletedCapabilities`: Realizations where the component ONLY realises capabilities in the deletion set — candidates for application deletion
- `realizationsOnRetainedCapabilities`: Realizations where the component also realises capabilities OUTSIDE the deletion set — always retained

**Swagger annotation:** Relative path `/capabilities/{id}/delete-impact` (no `/api/v1/` prefix)

**Status codes:**
| Code | Condition |
|------|-----------|
| 200 | Success |
| 401 | No valid session |
| 403 | Actor lacks `capabilities:read` |
| 404 | Capability not found |

### DELETE /capabilities/{id} (Enhanced)

Replaces the existing delete handler with cascade support.

**Request body (JSON):**
```json
{
  "cascade": true,
  "deleteRealisingApplications": false
}
```

- `cascade` (bool, default `false`): Must be `true` if capability has descendants. Returns 409 if `false` and children exist.
- `deleteRealisingApplications` (bool, default `false`): When `true`, deletes realizations and eligible applications.

Missing body is treated as `{ cascade: false, deleteRealisingApplications: false }`.

**Response — 204 No Content** on success.

**Response — 409 Conflict** when `cascade: false` but children exist:
```json
{
  "error": "Conflict",
  "message": "Capability has descendants. Set cascade:true to confirm cascade deletion.",
  "_links": {
    "x-delete-impact": { "href": "/capabilities/{id}/delete-impact", "method": "GET" }
  }
}
```

**Status codes:**
| Code | Condition |
|------|-----------|
| 204 | Success |
| 400 | Malformed body |
| 401 | No valid session |
| 403 | Actor lacks `capabilities:delete` |
| 404 | Capability not found |
| 409 | Has descendants but `cascade: false` |

### HATEOAS Link Changes

In `capabilityBaseForActor` (`links.go`):
```go
if actor.CanDelete("capabilities") {
    links["delete"] = h.Del(p)
    links["x-delete-impact"] = h.Get(p + "/delete-impact")
}
```

The impact response includes `x-confirm-delete` pointing to the DELETE endpoint (only when actor can delete). This creates a two-step HATEOAS discovery chain.

## DDD Architecture

### No Saga Needed
All affected aggregates are within Capability Mapping. Each aggregate save is a separate event store transaction. The existing `DeleteApplicationComponentHandler` pattern (sequential command dispatch within one HTTP handler) is the precedent. Partial failure leaves a consistent state — the user can retry.

### New Command
```
DeleteCapabilityWithCascade {
  ID: CapabilityID
  Cascade: bool
  DeleteRealisingApplications: bool
}
```

### Value Object: DeletionScope
Internal to the handler — represents the set of capabilities being deleted:
```go
type DeletionScope struct {
    rootID        CapabilityID
    capabilityIDs []CapabilityID
    idSet         map[string]bool
}
func (s DeletionScope) Contains(id string) bool
func (s DeletionScope) BottomUp() []CapabilityID  // L4 → L3 → L2 → L1
```

### Impact Read Model (Computed, Not Persisted)
Assembled on-the-fly from existing read models — no new database table:
1. `CapabilityHierarchyService.GetDescendants()` → subtree IDs
2. `RealizationReadModel.GetByCapabilityID()` per capability → all realizations
3. For each component: `RealizationReadModel.GetByComponentID()` → check if all realizations are within the deletion set
4. Components with all realizations in set → `realizationsOnDeletedCapabilities`
5. Components with realizations outside set → `realizationsOnRetainedCapabilities`

### Cascade Handler Flow
```
DeleteCapabilityWithCascadeHandler:
  1. Load capability, return 404 if not found
  2. Build DeletionScope via GetDescendants
  3. If hasDescendants && !cascade → return 409 with x-delete-impact link
  4. Delete all realizations for capabilities in scope
  5. Delete all dependencies involving capabilities in scope
  6. Delete capabilities bottom-up (BottomUp ordering)
     Each raises CapabilityDeleted → downstream handlers fire normally
  7. If deleteRealisingApplications:
     For each component exclusively in scope:
       Loopback HTTP DELETE /api/v1/components/{id}
  8. Return 204
```

### New Domain Error
```go
var ErrCascadeRequiredForChildCapabilities = errors.New("cascade deletion required for capability with descendants")
```
Registered in `error_registration.go` as 409 Conflict.

## Frontend Design

### Context Menu Entry Points

All three surfaces already gate "Delete from Model" on `hasLink(capability, 'delete')`. The change is minimal — they open the new impact-aware dialog instead of the old simple one.

| Surface | File | Change |
|---------|------|--------|
| Business Domain Grid | `business-domains/hooks/useCapabilityContextMenu.ts` | Pass `domainId` to dialog |
| Tree View | `navigation/hooks/useTreeContextMenus.ts` | No hook changes; dialog handles impact |
| Canvas | `canvas/hooks/useDeleteConfirmation.ts` | Route `capability-from-model` through dialog instead of direct mutation |

### Dialog Flow
```
DeleteCapabilityDialog opens
  ↓
useDeleteImpact(capability.id) fires
  ├── Loading → Skeleton UI
  ├── Error → Alert with retry
  └── Success →
      ├── No children → Simple confirm: "Delete {name}?" → [Cancel] [Delete]
      └── Has children → Cascade view:
            ├── Warning: "This will delete {name} and {N} child capabilities"
            ├── ImpactSection: child capabilities list (collapsible, grouped by level)
            ├── ImpactSection: affected applications
            │     "Will lose realizations" (retained apps)
            │     "Can be deleted" (exclusively-realising apps)
            ├── Checkbox: "Also delete {N} applications that only realise these capabilities"
            └── [Cancel] [Delete {name} and {N} children]
```

### New/Modified Frontend Files

| File | Change |
|------|--------|
| `api/types.ts` | Add `CapabilityDeleteImpact`, `DeleteImpactChild`, `DeleteImpactApplication` types |
| `capabilities/api/capabilitiesApi.ts` | Add `getDeleteImpact()`, `cascadeDelete()` |
| `capabilities/queryKeys.ts` | Add `deleteImpact` key |
| `capabilities/hooks/useCapabilities.ts` | Add `useDeleteImpact()` query, `useCascadeDeleteCapability()` mutation |
| `capabilities/hooks/useDeleteCapabilityFlow.ts` | NEW: orchestrates simple vs cascade |
| `capabilities/components/DeleteCapabilityDialog.tsx` | REWRITE: impact-aware dialog |
| `capabilities/components/ImpactSection.tsx` | NEW: collapsible impact list |
| `capabilities/mutationEffects.ts` | Add `cascadeDelete` effect |
| `canvas/hooks/useDeleteConfirmation.ts` | Route capability deletes through dialog |
| `test/mocks/handlers.ts` | Add `GET /capabilities/:id/delete-impact` handler |

### Cache Invalidation (mutationEffects)

```typescript
cascadeDelete: (context) => [
  capabilitiesQueryKeys.lists(),
  capabilitiesQueryKeys.details(),
  capabilitiesQueryKeys.children(context.parentId),
  capabilitiesQueryKeys.realizationsByComponents(),
  businessDomainsQueryKeys.lists(),
  businessDomainsQueryKeys.details(),
  maturityAnalysisQueryKeys.unlinked(),
  valueStreamsQueryKeys.all,
  artifactCreatorsQueryKeys.all,
  auditQueryKeys.history(context.id),
  // Only if deleteApplications:
  ...(context.deleteApplications ? [
    componentsQueryKeys.lists(),
    componentsQueryKeys.details(),
  ] : []),
]
```

## Backend Files to Change

### New Files
| File | Purpose |
|------|---------|
| `capabilitymapping/application/commands/cascade_delete_capability.go` | Command struct |
| `capabilitymapping/application/handlers/cascade_delete_capability_handler.go` | Orchestrating handler |
| `capabilitymapping/infrastructure/api/delete_impact_handler.go` | Impact analysis HTTP handler |

### Modified Files
| File | Change |
|------|--------|
| `infrastructure/api/links.go` | Add `x-delete-impact` link |
| `infrastructure/api/routes.go` | Register `GET /{id}/delete-impact`, wire cascade handler |
| `infrastructure/api/capability_handlers.go` | Add `DeleteImpactResponse` types, modify delete handler to accept body |
| `infrastructure/api/error_registration.go` | Register `ErrCascadeRequiredForChildCapabilities` |
| `domain/services/capability_deletion_service.go` | Add new error var |

## Implementation Order
1. Backend: impact analysis read model + `GET /capabilities/{id}/delete-impact` endpoint
2. Backend: `DeleteCapabilityWithCascade` command + handler (bottom-up deletion)
3. Backend: enhanced `DELETE /capabilities/{id}` with body parsing
4. Backend: HATEOAS link additions
5. Frontend: API types + API layer + query keys
6. Frontend: `useDeleteImpact` + `useCascadeDeleteCapability` hooks
7. Frontend: rewrite `DeleteCapabilityDialog` with impact analysis UI
8. Frontend: update canvas `useDeleteConfirmation` to route through dialog
9. Frontend: mutation effects for cache invalidation
10. Frontend: MSW handlers + tests

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Single request vs saga | Single request | All aggregates in same context; `DeleteApplicationComponentHandler` precedent |
| New command vs modify existing | New `DeleteCapabilityWithCascade` | Keeps simple delete path unchanged |
| Cross-context app deletion | Loopback HTTP | Respects BC boundaries and RBAC; Arch Assistant precedent |
| Impact as persisted read model | Computed on-the-fly | Rare use, data changes constantly; existing read models sufficient |
| DELETE body vs query params | Request body | More expressive for structured options; precedent in import sessions |
| Bottom-up ordering | L4 → L3 → L2 → L1 | Avoids modifying existing `CanDelete` validation |

## Checklist
- [x] Backend: `GET /capabilities/{id}/delete-impact` endpoint
- [x] Backend: `DeleteCapabilityWithCascade` command and handler
- [x] Backend: Enhanced `DELETE /capabilities/{id}` with cascade body
- [x] Backend: HATEOAS links (`x-delete-impact`, `x-confirm-delete`)
- [x] Backend: Error registration for cascade-required
- [x] Backend: Unit tests for cascade handler
- [ ] Backend: Integration tests for cascade delete
- [x] Frontend: API types and API layer
- [x] Frontend: Query keys and hooks
- [x] Frontend: Impact-aware DeleteCapabilityDialog
- [x] Frontend: ImpactSection component (inlined into DeleteCapabilityDialog)
- [x] Frontend: Canvas delete confirmation routing
- [x] Frontend: Mutation effects for cache invalidation
- [x] Frontend: MSW handlers
- [ ] Frontend: Component tests
- [x] Frontend: Build passes (`npm run build`)
- [ ] User sign-off
