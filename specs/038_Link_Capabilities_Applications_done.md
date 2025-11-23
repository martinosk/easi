# Link Capabilities to Applications

## Description
Enable connecting capabilities to applications on the canvas to model which applications realize which capabilities. Realization edges are visually distinct from parent and dependency edges. Supports multiple capabilities per application and enforces hierarchical realization rule.

## User Need
As a user, I need to link capabilities to applications so that I can document which systems realize which business capabilities.

## Visual Design

### Realization Edge
```
┌────────────┐
│ L3: Billing│ ← Capability
└─────┬──────┘
      │ ╌╌╌╌╌╌╌ ← Dashed green line
      ▼
┌────────────┐
│ SAP System │ ← Application
└────────────┘
```

**Edge Styling:**
- Line style: Dashed
- Color: Green (#10B981)
- Width: 2px
- Label: "Realizes"
- Arrow: Standard arrowhead

### Inherited Realization
When auto-created from hierarchical rule:
- Same styling but with 60% opacity
- Label: "Realizes (inherited)"
- **Visibility rule: Only displayed when the directly realized capability is not visible on canvas**
- If both direct and inherited capabilities are visible, only the direct realization edge is shown

## Functional Requirements

### Create Realization
1. Connect capability to application (or vice versa)
2. Backend `POST /api/v1/capabilities/:id/systems` called
3. Realization edge displayed with green dashed styling
4. Default coverage: 100%

### Hierarchical Realization Rule
When application realizes a capability:
1. System automatically creates realizations for all ancestor capabilities
2. If app realizes L3, also create realizations for L2 and L1 parents
3. **Visual display follows visibility rule:**
   - If direct realization source is visible on canvas, show only the direct edge
   - If direct realization source is not visible, show inherited edge to closest visible ancestor
   - Inherited edges displayed with reduced opacity (60%) and "(inherited)" label

Example scenarios:
- Canvas shows L3 and L2: Only L3→App edge visible
- Canvas shows only L2 (L3 not present): L2→App edge visible with inherited styling
- Canvas shows L1, L2, L3: Only L3→App edge visible
- Canvas shows only L1 (L2, L3 not present): L1→App edge visible with inherited styling

## Domain Model

### RealizationOrigin Value Object
Distinguishes realizations by their provenance:
- **Direct**: Explicitly created by user action
- **Inherited**: Automatically derived from hierarchical rule

### SourceRealizationID Value Object
Optional reference to the causative direct realization:
- Null for Direct realizations
- Required for Inherited realizations - references the Direct realization that triggered creation

### Invariants
1. **Origin Immutability**: A realization's origin cannot change after creation
2. **Source Validity**: Inherited realizations must reference a valid Direct realization; if source is deleted, inherited must cascade delete
3. **No Duplicate Direct Realizations**: Only one Direct realization per capability-component pair
4. **Promotion Rule**: Creating a Direct realization for an existing Inherited pair replaces the Inherited with Direct
5. **Deletion Cascade**: Deleting a Direct realization cascades to all Inherited realizations with that source (unless ancestor has another Direct from same component)

### Edit Realization
Dialog fields:
- Capability name (read-only)
- Application name (read-only)
- Coverage Percent (0-100)
- Notes (optional text)

### Delete Realization
1. Right-click edge → Delete
2. Confirmation dialog
3. Backend `DELETE /api/v1/capability-realizations/:id` called
4. Edge removed from canvas

### Multiple Capabilities
- Application can realize multiple capabilities
- Each realization shown as separate edge
- All visible on canvas simultaneously

## Details Panel Updates

### Application Selected
```
Application: SAP System
━━━━━━━━━━━━━━━━━━━━━━━
Realizes Capabilities:
• L3: Billing (100%) ← direct
• L2: Finance (100%) ← inherited
• L1: Core Operations (100%) ← inherited
```
Note: Details panel always shows all realizations (both direct and inherited) regardless of canvas visibility. The visibility rule only applies to edge rendering on canvas.

### Capability Selected
```
Capability: L3: Billing
━━━━━━━━━━━━━━━━━━━━━━━
Realized By:
• SAP System (100%)
• Legacy Billing App (75%)
```

## Edge Context Menu
Right-click on realization edge:
```
┌──────────────┐
│ Edit         │
│ Delete       │
└──────────────┘
```

## Acceptance Criteria

### Domain Model
- [x] RealizationOrigin value object with Direct/Inherited variants
- [x] SourceRealizationID value object for tracking inheritance source
- [x] Invariants enforced: origin immutability, source validity, no duplicate directs
- [x] Deletion cascade: removing Direct realization removes its Inherited realizations

### Backend
- [x] Backend API called to create linkage
- [x] Hierarchical rule enforced: Direct realization creates Inherited ancestor realizations
- [x] Each realization stores origin (Direct/Inherited) and sourceRealizationID

### Frontend
- [x] Can connect capability to application (either direction)
- [x] Realization edge created with distinct styling (green dashed)
- [x] Inherited realizations only shown when directly realized capability is not visible on canvas
- [x] When both direct and inherited capabilities are visible, only direct edge is displayed
- [x] Inherited realizations shown with different styling (60% opacity) when displayed
- [x] Edge visibility updates dynamically when capabilities are added/removed from canvas
- [x] Multiple capabilities can link to same application
- [x] Coverage percentage can be edited (0-100)
- [x] Notes can be added to realization
- [x] Realization edges can be deleted via context menu
- [x] Application details panel shows realized capabilities (distinguishing direct vs inherited)
- [x] Capability details panel shows realizing applications
- [x] Existing realizations loaded and displayed when canvas loads

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] Documentation updated if needed
- [x] User sign-off
