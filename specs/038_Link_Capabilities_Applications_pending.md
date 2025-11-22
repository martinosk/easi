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
3. Ancestor realizations shown with reduced opacity if visible on canvas
4. Notes field indicates "Auto-created via hierarchical realization rule"

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
• L3: Billing (100%)
• L2: Finance (100%) [inherited]
• L1: Core Operations (100%) [inherited]
```

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
- [ ] Can connect capability to application (either direction)
- [ ] Realization edge created with distinct styling (green dashed)
- [ ] Backend API called to create linkage
- [ ] Hierarchical rule enforced: child realization creates ancestor realizations
- [ ] Inherited realizations shown with different styling (reduced opacity)
- [ ] Multiple capabilities can link to same application
- [ ] Coverage percentage can be edited (0-100)
- [ ] Notes can be added to realization
- [ ] Realization edges can be deleted via context menu
- [ ] Application details panel shows realized capabilities
- [ ] Capability details panel shows realizing applications
- [ ] Existing realizations loaded and displayed when canvas loads

## Checklist
- [ ] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] Documentation updated if needed
- [ ] User sign-off
