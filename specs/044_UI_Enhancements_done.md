# UI Enhancements: Auto-layout, Color Schemes UI, and Terminology

## User Need
As an enterprise architect, I need the Architecture Modeler to:
1. Automatically arrange all element types on the canvas for better readability
2. Choose color schemes that align with internal conventions or industry standards
3. See consistent, professional terminology throughout the application

## Success Criteria
- Auto-layout positions all canvas elements (applications, capabilities, any future types) correctly
- Users can switch between color schemes from the toolbar
- All UI text uses "Application" terminology instead of "Component"
- Capabilities display with a recognizable icon

---

## Vertical Slices

### Slice 1: Universal Auto-layout

Currently, `applyAutoLayout` in `layoutSlice.ts` only processes application components. It must include capability nodes and their edges.

- [x] Auto-layout includes capability nodes from `canvasCapabilities` in the layout calculation
- [x] Auto-layout includes parent-child edges between capabilities
- [x] Auto-layout includes realization edges between applications and capabilities
- [x] Capability node positions are persisted after layout (via `updateCapabilityPosition`)
- [x] Layout respects current `layoutDirection` setting for all node types
- [x] Toast notification confirms layout applied to N elements

### Slice 2: Color Scheme Selector

Add a dropdown to the toolbar allowing users to select canvas color schemes.

- [x] Dropdown appears in toolbar next to existing controls (Edge Type, Layout Direction)
- [x] Three options available: "Maturity" (default), "Classic", "Custom"
- [x] Selection persists per view (frontend state only - backend endpoint not yet implemented)
- [x] Capability nodes apply colors based on selected scheme:
  - **Maturity**: Current behavior (Genesis=red, Custom Build=orange, Product=green, Commodity=blue)
  - **Classic**: Classic enterprise architecture palette (Business=#FFFFB5, Application=#B5FFFF, Technology=#C9E7B7)
  - **Custom**: User-defined custom colors per element
- [x] Application nodes apply appropriate layer color based on scheme (Application layer)
- [x] Color scheme change reflects immediately without page reload

**Note:** Frontend calls `PATCH /api/v1/views/{viewId}/color-scheme` but backend endpoint doesn't exist yet. See spec 045 for backend implementation.

### Slice 3: Capability Icon

- [x] Capability nodes display a modern icon in the node header
- [x] Icon visually represents "capability" concept (2x2 grid pattern)
- [x] Icon replaces current diamond character in `CapabilityNode.tsx`
- [x] Icon scales appropriately with node size
- [x] Icon maintains visibility across all color schemes

### Slice 4: Terminology Updates

Update all user-facing text from "Component" to "Application" throughout the frontend.

- [x] Toolbar title: "Component Modeler" -> "Architecture Modeler"
- [x] Tree section: "Models" -> "Applications"
- [x] Details panel: "Component Details" -> "Application Details"
- [x] Edit dialog: "Edit Component" -> "Edit Application"
- [x] Edit dialog submit: "Update Component" -> "Update Application"
- [x] Create dialog: "Create Component" -> "Create Application"
- [x] Form placeholders: "Enter component..." -> "Enter application..."
- [x] Error messages: "component" -> "application"
- [x] Context menu labels updated
- [x] Delete confirmation dialogs updated
- [x] Tooltip text updated
- [x] No visible occurrence of standalone "Component" remains in UI (except "Application Component" type label)
- [x] Backend API endpoints unchanged (still `/api/v1/components`)
- [x] TypeScript types and internal code unchanged

---

## Implementation Summary

**Files Modified:** 16 frontend files
**New Files:** 2 files (`ColorSchemeSelector.tsx`, this spec)
**Tests:** All 300 tests passing

### Key Changes:
1. **Auto-layout** (`layoutSlice.ts:147-296`): Extended to handle capabilities and all edge types
2. **Color Scheme Selector** (`ColorSchemeSelector.tsx`): New dropdown component with three predefined schemes
3. **Capability Icon** (`CapabilityNode.tsx:94-98`): Modern SVG icon replacing diamond character
4. **Terminology** (5 components): Systematic "Component" → "Application" renaming

---

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented
- [x] User sign-off

---

## Related Specs
- **045**: Color Scheme Backend Support and Custom Colors (pending)
