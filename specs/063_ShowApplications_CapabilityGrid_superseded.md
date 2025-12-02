# Show Applications in Business Domain Capability Grid

## Status
**Ongoing**

## User Need
Enterprise architects visualizing business domains need to see which applications realise each capability directly within the grid view. When the depth filter hides a capability, architects should optionally see those applications "bubble up" to the nearest visible parent capability.

## Dependencies
- Spec 058: Business Domain Visualization UI (grid visualization)
- Spec 026: Capability System Realization (application-capability links)

## Superseded By
- **Spec 064: Business Domain Realizations API** - The API design in this spec (passing capabilityIds as query parameters) is superseded by spec 064, which uses a semantic endpoint accepting domainId instead.

---

## User Stories

1. As an architect, I can toggle "Show Applications" to display realising applications as nested boxes within capabilities
2. As an architect, I can see which applications realise a specific capability at a glance
3. As an architect, I can toggle "Show inherited realisations" to see applications from hidden child capabilities appear in visible parent capabilities
4. As an architect, I can click on an application box to navigate to its detail view
5. As an architect, I can distinguish between direct realisations and inherited realisations

## Success Criteria

- Toggle control "Show Applications" in the grid toolbar (off by default)
- When enabled, application boxes appear inside the capabilities they realise
- Application boxes are styled distinctly from capability boxes (different shape/color)
- Application name displayed; truncated with tooltip for overflow
- Realization level indicator (Full/Partial/Planned) shown via visual style
- "Show inherited realisations" checkbox appears only when "Show Applications" is enabled
- Inherited applications show visual indicator distinguishing them from direct realisations
- Click on application box opens application detail

---

## Vertical Slices

### Slice 1: Show Applications Toggle + Direct Realisations

Display applications within the capabilities they directly realise.

**Frontend:**
- [x] Add "Show Applications" toggle button to grid toolbar (next to depth selector)
- [x] Create `useCapabilityRealisations(capabilityIds)` hook to batch-fetch realisations for visible capabilities
- [x] Create `ApplicationChip` component for displaying application within capability
- [x] Integrate chips into `NestedCapabilityGrid` capability cells
- [x] Style chips: smaller than capabilities, rounded rectangle, neutral color (gray/slate)
- [x] Show application name with text truncation and title tooltip
- [x] Show realisation level via border/badge: Full (solid), Partial (dashed), Planned (dotted)

**API:**
- [x] Create batch endpoint: `GET /api/v1/capability-realizations?capabilityIds=id1,id2,...`
- [x] Response includes component name for display (denormalized for performance)
- [x] Response includes `origin` field ("Direct" or "Inherited") and `sourceCapabilityName` for inherited realisations

### Slice 2: Application Details on Click

Enable navigation to application details when clicking an application chip.

- [x] Add click handler to `ApplicationChip`
- [x] Open side panel with application details (similar to capability detail panel,use existing component detail infrastructure)

### Slice 3: Inherited Realisations

Show applications that realise child capabilities (inherited realisations are already computed by the backend).

**Backend Context:**
The realization projector already creates inherited realizations automatically. When an app realizes an L3 capability, the projector creates:
- Direct realization: L3 ← App (origin="Direct")
- Inherited realization: L2 ← App (origin="Inherited", sourceCapabilityId=L3)
- Inherited realization: L1 ← App (origin="Inherited", sourceCapabilityId=L3)

See: `backend/internal/capabilitymapping/application/projectors/realization_projector.go`

**Frontend:**
- [x] Add "Show inherited" toggle button (hidden unless "Show Applications" is enabled)
- [x] Persist preference in localStorage alongside depth preference
- [x] Filter realisations by `origin` field:
  - Toggle OFF: Show only realisations where `origin === "Direct"`
  - Toggle ON: Show all realisations (both Direct and Inherited)
- [x] Display inherited applications with distinct visual indicator (icon badge)
- [x] Tooltip on inherited apps shows: "Realises [Source Capability Name]" (from `sourceCapabilityName` field)

### Slice 4: Performance Optimization

Ensure acceptable performance with realistic data volumes.

- [x] Batch API call fetches all realisations for visible capabilities in one request
- [x] Frontend caches realisations, invalidates only on relevant mutations
- [x] Lazy load realisations only when "Show Applications" is toggled on
- [x] Consider limiting displayed apps per capability with "+N more" indicator if > 5

---

## Technical Requirements

### Frontend

**New Components:**
- `ShowApplicationsToggle`: Toolbar toggle with toggle button for inherited mode
- `ApplicationChip`: Small component displaying single application
- `ApplicationChipList`: Container managing overflow and layout within capability cell

**Hooks:**
- `useCapabilityRealisations(capabilityIds: CapabilityId[], enabled: boolean)`: Fetches and caches realisations for visible capabilities

**State:**
- `showApplications: boolean` - localStorage persisted
- `showInheritedRealisations: boolean` - localStorage persisted (only relevant when showApplications=true)

**Layout:**
- Application chips flow horizontally within capability cell
- Wrap to multiple rows if needed
- Minimum capability cell size accommodates 2-3 visible chips
- Overflow indicator if more apps than fit

### Backend

**New Endpoint:**
```
GET /api/v1/capability-realizations?capabilityIds=id1,id2,id3
```

This follows REST collection filtering conventions and is extensible for future filters.

Response:
```json
{
  "data": [
    {
      "id": "realization-123",
      "capabilityId": "cap-456",
      "componentId": "comp-789",
      "componentName": "Order Service",
      "realizationLevel": "Full",
      "origin": "Direct",
      "sourceCapabilityId": null,
      "sourceCapabilityName": null,
      "notes": "Primary implementation",
      "linkedAt": "2025-01-15T10:00:00Z",
      "_links": {
        "self": "/api/v1/capability-realizations/realization-123",
        "capability": "/api/v1/capabilities/cap-456",
        "component": "/api/v1/components/comp-789"
      }
    },
    {
      "id": "realization-456",
      "capabilityId": "cap-123",
      "componentId": "comp-789",
      "componentName": "Order Service",
      "realizationLevel": "Full",
      "origin": "Inherited",
      "sourceCapabilityId": "cap-456",
      "sourceCapabilityName": "Order Processing",
      "notes": null,
      "linkedAt": "2025-01-15T10:00:00Z",
      "_links": {
        "self": "/api/v1/capability-realizations/realization-456",
        "capability": "/api/v1/capabilities/cap-123",
        "component": "/api/v1/components/comp-789",
        "sourceCapability": "/api/v1/capabilities/cap-456"
      }
    }
  ],
  "_links": {
    "self": "/api/v1/capability-realizations?capabilityIds=cap-456,cap-123"
  }
}
```

**Read Model Extension:**
- Extend `RealizationDTO` to include `componentName` (denormalized)
- Extend `RealizationDTO` to include `sourceCapabilityName` for inherited realisations (denormalized)
- Update read model projector to denormalize component name on realization events
- Handle component renamed events to update denormalized names
- Handle capability renamed events to update `sourceCapabilityName`

### Visual Design

**Application Chip:**
- Background: `#e2e8f0` (slate-200)
- Border radius: `4px`
- Padding: `2px 8px`
- Font size: smaller than capability name
- Max width: constrained to prevent overflow

**Realisation Level Indicators:**
| Level   | Border Style | Color      |
|---------|--------------|------------|
| Full    | Solid 2px    | #22c55e (green) |
| Partial | Dashed 2px   | #eab308 (yellow) |
| Planned | Dotted 2px   | #94a3b8 (gray) |

**Inherited Indicator:**
- Lighter background: `#f1f5f9` (slate-100)
- Small icon badge (arrow-up or hierarchy icon)
- Tooltip: "Inherited from [Child Capability Name]"

---

## Acceptance Criteria Summary

- [x] "Show Applications" toggle visible in grid toolbar
- [x] Applications display as chips within capabilities when enabled
- [x] Realisation level visually distinguished (Full/Partial/Planned)
- [x] Click on application opens detail panel
- [x] "Show inherited" toggle visible when "Show Applications" enabled
- [x] Inherited realisations filtered by `origin` field from API response
- [x] Inherited applications visually distinct from direct realisations
- [x] Tooltip shows source capability name for inherited realisations
- [x] Settings persist across page refreshes
- [x] Batch API endpoint `GET /api/v1/capability-realizations?capabilityIds=...` exists
- [x] Component name and source capability name denormalized in response

## Checklist
- [x] Specification ready
- [ ] User sign-off on spec
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] Final user sign-off
