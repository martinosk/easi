# Color Scheme Backend Support and Custom Colors

## User Need
As an enterprise architect using the Architecture Modeler, I need:
1. My color scheme selection to persist when I reload the view
2. The ability to assign custom colors to individual elements for visual differentiation
3. Color choices to be view-specific (same element can have different colors in different views)

## Success Criteria
- Color scheme selection persists per view in the backend
- Users can assign custom hex colors to individual capabilities and components
- Custom colors only render when "custom" color scheme is active
- Custom colors are preserved when switching between schemes (not deleted)
- All color operations follow REST Level 3 maturity with HATEOAS

---

## Context

**Prerequisite:** Spec 044 (UI Enhancements) implemented the frontend color scheme selector, but the backend endpoint doesn't exist. The frontend currently calls `PATCH /api/v1/views/{viewId}/color-scheme` which returns 404.

**Architecture:** The architectureviews bounded context uses a hybrid pattern:
- **Event Sourced**: View membership, metadata, lifecycle
- **CRUD**: Layout/visual properties (positions, edge type, layout direction, color scheme)

This spec follows the existing CRUD pattern for visual properties via `ViewLayoutRepository`.

---

## Vertical Slices

### Slice 1: Backend Color Scheme Persistence

Implement the missing backend endpoint for color scheme persistence.

**API Endpoint:**
```http
PATCH /api/v1/views/{viewId}/color-scheme
Body: { "colorScheme": "maturity" | "archimate" | "archimate-classic" | "custom" }

Response: 200 OK
{
  "colorScheme": "archimate",
  "_links": {
    "self": "/api/v1/views/{viewId}/color-scheme",
    "view": "/api/v1/views/{viewId}"
  }
}
```

**Implementation:**
- Database: Add `color_scheme VARCHAR(20)` to `view_preferences` table
- Value Object: `ColorScheme` at `backend/internal/architectureviews/domain/valueobjects/color_scheme.go`
- Command: `UpdateViewColorScheme`
- Handler: `UpdateViewColorSchemeHandler` (uses `ViewLayoutRepository`)
- Repository: `ViewLayoutRepository.UpdateColorScheme(ctx, viewID, colorScheme)`
- Read Model: Include `colorScheme` in GET view responses

**Pattern Reference:** Follow existing `EdgeType` and `LayoutDirection` implementations

**Acceptance Criteria:**
- [x] Migration adds `color_scheme` column to `view_preferences`
- [x] `ColorScheme` value object validates allowed values (maturity, archimate, archimate-classic, custom)
- [x] `UpdateViewColorScheme` command and handler implemented
- [x] `ViewLayoutRepository.UpdateColorScheme` method added
- [x] PATCH endpoint returns 200 OK with body and HATEOAS links
- [x] GET `/api/v1/views/{viewId}` returns `colorScheme` field at view level
- [x] Handler registered in command bus
- [x] Route registered in router
- [x] Integration tests for color scheme CRUD
- [x] Unit tests for ColorScheme value object

---

### Slice 2: Custom Color Picker UI

Add react-color library and color picker to details panels.

**User Flow:**
1. User sets view color scheme to "custom"
2. User selects a capability or component
3. Details panel shows color picker (enabled only when scheme is "custom")
4. User picks a color → optimistic update → API call
5. Element renders with custom color immediately

**Frontend Implementation:**
- Add `react-color` dependency to `package.json`
- Create `ColorPicker` component (wrapper around SketchPicker or CompactPicker)
- Add color picker to `CapabilityDetails.tsx` and `ComponentDetails.tsx`
- Disable picker when `colorScheme !== "custom"` with tooltip: "Switch to custom color scheme to assign colors"
- Store colors as hex strings: `#RRGGBB`
- Optimistic update with rollback on API failure

**Acceptance Criteria:**
- [x] react-color dependency added to package.json (using react-colorful)
- [x] ColorPicker component created with proper styling
- [x] Color picker appears in CapabilityDetails panel
- [x] Color picker appears in ComponentDetails panel
- [x] Picker disabled when colorScheme is not "custom" (with explanatory tooltip)
- [x] Selecting color calls appropriate API endpoint
- [x] Color picker displays current custom color (if set)
- [x] Optimistic update + rollback on error
- [x] Unit tests for ColorPicker component (14/14 passing)

---

### Slice 3: Backend Custom Color Persistence

Store custom colors per-element in view positions.

**Architectural Note:** Colors are stored in `view_element_positions` (not on capability/component entities) because colors are view-specific presentation concerns.

**API Endpoints:**
```http
PATCH /api/v1/views/{viewId}/components/{componentId}/color
Body: { "color": "#FF5733" }
Response: 204 No Content

PATCH /api/v1/views/{viewId}/capabilities/{capabilityId}/color
Body: { "color": "#FF5733" }
Response: 204 No Content

DELETE /api/v1/views/{viewId}/components/{componentId}/color
Response: 204 No Content

DELETE /api/v1/views/{viewId}/capabilities/{capabilityId}/color
Response: 204 No Content
```

**Implementation:**
- Database: Add `custom_color VARCHAR(7)` to `view_element_positions` table
- Value Object: `HexColor` with regex validation `^#[0-9A-Fa-f]{6}$`
- Commands: `UpdateElementColor`, `ClearElementColor`
- Handlers: Use `ViewLayoutRepository`
- Repository Methods: `UpdateElementColor(ctx, viewID, elementID, elementType, color)`, `ClearElementColor(...)`
- Read Model: Include `customColor` (nullable) in component/capability arrays

**Enhanced GET View Response:**
```json
{
  "id": "view-123",
  "name": "Architecture Overview",
  "colorScheme": "custom",
  "components": [
    {
      "id": "comp-1",
      "name": "Order Service",
      "x": 100,
      "y": 200,
      "customColor": "#FF5733",
      "_links": {
        "self": "/api/v1/components/comp-1",
        "updateColor": "/api/v1/views/view-123/components/comp-1/color",
        "clearColor": "/api/v1/views/view-123/components/comp-1/color"
      }
    }
  ],
  "capabilities": [
    {
      "id": "cap-1",
      "name": "Order Processing",
      "x": 150,
      "y": 250,
      "customColor": null,
      "_links": {
        "self": "/api/v1/capabilities/cap-1",
        "updateColor": "/api/v1/views/view-123/capabilities/cap-1/color",
        "clearColor": "/api/v1/views/view-123/capabilities/cap-1/color"
      }
    }
  ],
  "_links": {
    "self": "/api/v1/views/view-123",
    "updateColorScheme": "/api/v1/views/view-123/color-scheme"
  }
}
```

**Acceptance Criteria:**
- [x] Migration adds `custom_color` column to `view_element_positions`
- [x] `HexColor` value object validates hex format
- [x] `UpdateElementColor` command, handler, and repository method
- [x] `ClearElementColor` command, handler, and repository method
- [x] PATCH endpoints return 204 No Content (consistent with position updates)
- [x] DELETE endpoints return 204 No Content
- [x] GET view endpoints return `customColor` field (nullable) on elements
- [x] HATEOAS links for updateColor and clearColor (added to DTOs)
- [x] Handlers registered in command bus
- [x] Routes registered in router
- [x] Integration tests for color CRUD operations
- [x] Unit tests for HexColor value object (20/20 passing)

---

### Slice 4: Custom Color Rendering

Render elements with custom colors when scheme is "custom".

**Frontend Implementation:**
- Update `CapabilityNode.tsx`: Use `customColor` from view position data when `colorScheme === "custom"`
- Update `ComponentNode.tsx`: Use `customColor` from view position data when `colorScheme === "custom"`
- Update `NavigationTree.tsx`: Show custom color indicator when `colorScheme === "custom"`
- Fallback: If `colorScheme === "custom"` but element has no `customColor`, use neutral gray

**Color Priority:**
1. If `colorScheme !== "custom"`: Use scheme-based color (ignore customColor)
2. If `colorScheme === "custom"` and `customColor` exists: Use customColor
3. If `colorScheme === "custom"` and `customColor` is null: Use neutral default (#E0E0E0)

**Acceptance Criteria:**
- [x] CapabilityNode renders custom background color from position data
- [x] ComponentNode renders custom background color from position data
- [x] NavigationTree shows custom color indicators
- [x] Elements without custom color use neutral default (#E0E0E0) when scheme is "custom"
- [x] Elements use scheme-based colors when scheme is not "custom" (ignore customColor)
- [x] Color changes reflect immediately (no page reload)
- [x] Unit tests for color rendering logic

---

## Architecture Decision: Color Consistency

**When color scheme changes from/to "custom", custom colors are PRESERVED (not deleted).**

**Rationale:**
- Users might experiment with schemes and switch back
- Prevents loss of hours of customization work
- Storage cost is negligible
- Enables "undo-friendly" UX

**Implementation:** Color scheme and custom colors are independent. Scheme determines rendering strategy; colors are data.

**UX Enhancement (optional):** Show indicator like "23 custom colors defined (currently hidden)" when scheme is not "custom" but colors exist.

---

## Out of Scope
- Bulk color operations (set multiple colors at once)
- Color picker for edges/relations
- Color gradients or transparency
- Color palettes/presets beyond the four schemes
- Persisting color scheme globally (remains per-view)
- Event sourcing for color changes (continues CRUD pattern)

## Checklist
- [x] Specification ready
- [x] Backend implementation (Slices 1 & 3)
- [x] Frontend implementation (Slices 2 & 4) - COMPLETE
- [x] Database migrations applied (migrations 019 & 020)
- [x] Unit tests implemented and passing (421 frontend tests + all backend tests)
- [x] Integration tests implemented and passing (backend tests created)
- [x] API documentation updated (OpenAPI/Swagger)
- [x] User sign-off

---

## Related Specs
- **044**: UI Enhancements (done) - Frontend color scheme selector already implemented

## Files Referenced
- `backend/internal/architectureviews/infrastructure/repositories/view_layout_repository.go`
- `backend/internal/architectureviews/domain/valueobjects/edge_type.go` (pattern reference)
- `backend/internal/architectureviews/application/handlers/update_view_edge_type_handler.go` (pattern reference)
- `frontend/src/components/ColorSchemeSelector.tsx` (already exists from spec 044)
- `frontend/src/store/slices/layoutSlice.ts` (already calls missing endpoint)
