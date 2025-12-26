# Enhanced Capability Maturity Editing

## Description
Replace the current maturity level dropdown with a 0-99 slider that provides full precision while displaying section context. The slider shows section boundaries as visual markers, allowing users to position capabilities precisely on the evolution scale.

## Purpose
Enable architects to express nuanced maturity positions for capabilities beyond the 4 discrete section values, while maintaining visual connection to the familiar section names.

## Dependencies
- Spec 091: Maturity Scale Configuration Aggregate
- Spec 092: Maturity Scale REST API (maturity-levels endpoint with ranges)
- Spec 093: Capability Maturity Data Migration
- Spec 095: Settings Page (for configuration source)

## Implementation Approach
Use Mantine's `Slider` component as the base, customized with section markers and labels. This maintains consistency with the existing Mantine-based UI.

## User Need
As an architect, I need to set granular maturity values for capabilities so I can express precise positions on the evolution scale rather than being limited to 4 discrete categories.

## Success Criteria
- Architects can set any value from 0-99 for capability maturity
- The slider displays current tenant's section names and boundaries
- Visual feedback shows which section the current value falls within
- The precise numeric value is visible during editing
- Existing capability edit dialogs use the new slider
- Create capability dialog uses the new slider

## Slider Component Design

### Visual Elements
- Horizontal track spanning full width (0-99)
- Section name labels above each section region
- Vertical boundary markers at section edges
- Draggable thumb showing current position
- Numeric value display (adjacent to thumb or as tooltip)
- Visual highlight of the active section

### Interactions
- Drag thumb to set value
- Click anywhere on track to jump to that value
- Arrow keys for fine adjustment (+/- 1)
- Value constrained to 0-99 range

### Responsive Behavior
- On mobile/tablet: touch-friendly thumb size, labels may stack vertically if needed
- Minimum touch target size per accessibility guidelines

## Component Integration

### EditCapabilityDialog
Replace the maturity level Select with MaturitySlider component. Initialize from capability's `maturityValue` field.

### CreateCapabilityDialog
Add MaturitySlider to the create form. Default value: midpoint of first section (calculated from current configuration, not hardcoded).

### CapabilityDetails (Side Panel)
Display maturity as: "Section Name (value)" format, e.g., "Custom Built (42)".

## API Changes

### Update Capability Metadata Request
Per Spec 093, the API accepts both `maturityLevel` (string, legacy) and `maturityValue` (number, 0-99). Frontend sends `maturityValue` only.

### Capability Response
Response includes `maturityValue` (number) and `maturitySection` object with name, order, and range. Section name is derived by backend.

### Metadata Query
Update `useMaturityLevels()` to include range information (min/max values per section) as defined in Spec 092.

## State Management

### Form State Changes
- Add `maturityValue: number` to form state in both EditCapabilityDialog and CreateCapabilityDialog
- String-based `maturityLevel` no longer stored in form state (derived for display)
- Default value for new capabilities: midpoint of first section from config

### Hooks
- `useMaturityScale()` - fetch tenant's maturity scale configuration (from Spec 095)
- Updated `useMaturityLevels()` - return section data with ranges for backward compatibility

### Loading State
While configuration is loading, show slider in disabled state with default section boundaries as fallback.

## Backward Compatibility

### Pre-Migration Capabilities
Capabilities without `maturityValue` (pre-migration) derive value from `maturityLevel` string using section midpoints. This logic should be encapsulated in a utility function, not duplicated across components.

### Configuration Fallback
If section configuration unavailable, use default Wardley mapping boundaries (Genesis 0-24, Custom Built 25-49, Product 50-74, Commodity 75-99).

## Accessibility

- Slider keyboard accessible via arrow keys (follows Mantine Slider behavior)
- ARIA attributes: `aria-valuemin`, `aria-valuemax`, `aria-valuenow`, `aria-valuetext` (e.g., "42 - Custom Built")
- Screen reader announces value and section name on change
- Focus visible state on thumb
- High contrast between section regions
- Live region for value changes during drag

## Acceptance Criteria

### MaturitySlider Component
- [ ] Renders horizontal slider track
- [ ] Displays 4 section regions with names
- [ ] Shows boundary markers at section edges
- [ ] Thumb is draggable across full range (0-99)
- [ ] Click on track moves thumb to position
- [ ] Arrow keys adjust value by 1
- [ ] Shift+Arrow adjusts value by 10
- [ ] Current numeric value is displayed
- [ ] Active section is visually highlighted

### EditCapabilityDialog Integration
- [ ] Slider replaces dropdown for maturity
- [ ] Initial value loads from capability's maturityValue
- [ ] Slider updates form state on change
- [ ] Save sends maturityValue to API
- [ ] Section name displays in confirmation/success

### CreateCapabilityDialog Integration
- [ ] Slider appears in create form
- [ ] Default value is 12
- [ ] Value included in create request

### CapabilityDetails Display
- [ ] Shows "Section Name (value)" format
- [ ] Example: "Custom Built (42)"

### Configuration Integration
- [ ] Slider boundaries match tenant's configuration
- [ ] Section names match tenant's configuration
- [ ] Updates when configuration changes

### Backward Compatibility
- [ ] Pre-migration capabilities show derived values
- [ ] Saving updates to numeric format
- [ ] No data loss during transition

## Out of Scope
- Bulk maturity updates for multiple capabilities
- Maturity history/change tracking UI
- Maturity comparison between capabilities
- Maturity suggestions/recommendations

## Checklist
- [ ] Specification ready
- [ ] MaturitySlider component created (using Mantine Slider)
- [ ] Section boundaries loaded from configuration
- [ ] Keyboard navigation and accessibility attributes
- [ ] Responsive behavior for touch devices
- [ ] EditCapabilityDialog updated
- [ ] CreateCapabilityDialog updated
- [ ] CapabilityDetails display updated
- [ ] Default value calculated from config (not hardcoded)
- [ ] Legacy value derivation utility created
- [ ] Fallback to default boundaries when config unavailable
- [ ] Capability type updated with `maturityValue` and `maturitySection`
- [ ] Unit tests for MaturitySlider
- [ ] Integration tests for edit flow
- [ ] Integration tests for create flow
- [ ] User sign-off
