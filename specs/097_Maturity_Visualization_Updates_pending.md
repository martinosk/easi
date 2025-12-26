# Maturity Visualization Updates

## Description
Update all canvas and visualization components to correctly display capabilities using the new 0-99 maturity scale and tenant-customized section names. Ensure color coding, labels, and tooltips reflect the configured maturity scale.

## Purpose
Ensure stakeholders see capabilities visualized with their organization's customized maturity terminology and that the numeric precision is reflected in color gradients.

## Dependencies
- Spec 091: Maturity Scale Configuration Aggregate
- Spec 092: Maturity Scale REST API
- Spec 093: Capability Maturity Data Migration
- Spec 095: Settings Page (configuration source)
- Spec 096: Capability Maturity Slider

## User Need
As a stakeholder viewing architecture visualizations, I need to see capabilities colored and labeled according to our customized maturity scale so the visualizations are meaningful to our organization.

## Success Criteria
- Canvas nodes display colors based on 0-99 maturity values
- Color gradients reflect position within sections
- Tooltips show tenant-customized section names
- Navigation tree color indicators use new scale
- Color scheme "Maturity" correctly maps values to colors
- All visualizations update when configuration changes

## Color Mapping Strategy

### Section-Based Gradient
Each section has a base color. Values within a section produce a gradient:
- Lower values in section = lighter shade
- Higher values in section = darker/more saturated shade

### Default Color Palette
| Section | Order | Light (min) | Saturated (max) |
|---------|-------|-------------|-----------------|
| Genesis | 1 | #FEE2E2 | #EF4444 |
| Custom Built | 2 | #FFEDD5 | #F97316 |
| Product | 3 | #FEF9C3 | #EAB308 |
| Commodity | 4 | #D1FAE5 | #10B981 |

### Color Calculation
1. Determine which section contains the value
2. Calculate position within section (0-1 scale)
3. Interpolate between light and saturated variants using HSL interpolation (produces better gradients than RGB)

### Performance Optimization
Pre-calculate color lookup table for all 100 values (0-99) when configuration loads. Memoize to avoid recalculation on every render.

## Affected Components

### CapabilityNode (Canvas)
Update to map `maturityValue` (0-99) to gradient color instead of mapping `maturityLevel` string. Add section name to tooltip.

### ComponentNode (Canvas)
If displaying linked capability maturity, update to use new scale.

### Navigation Panel
Update color indicators to use gradient colors based on maturity value. (Clarify: identify the specific component that shows capability list with color indicators.)

### CapabilityDetails Panel
Display maturity as "Section Name (value)" with colored badge matching the gradient.

### nodeFactory (Canvas Utils)
Ensure node data includes `maturityValue` field for color calculation.

### ColorSchemeSelector
No changes needed - "Maturity" scheme already exists. Implementation uses new color mapping.

## Hooks and Utilities

### useMaturityColorScale Hook
Provides color calculation functions. Returns methods for:
- Getting color for a maturity value (0-99)
- Getting section name for a maturity value
- Getting base color for a section order (1-4)

Uses memoized lookup table internally. Depends on `useMaturityScale()` hook from Spec 095.

### Color Utilities
Utility functions for color calculation (can be pure functions, not hooks):
- Calculate position within section (0-1 scale)
- Interpolate between two colors using HSL
- Get final color for a maturity value given section configuration

Move existing maturity color constants from CapabilityNode.tsx to a shared constants file.

## Configuration Integration

### Cache Strategy
- Maturity scale config cached with React Query
- Cache invalidated when settings updated (from Spec 095)
- Visualizations re-render automatically on cache update

### Fallback Behavior
If configuration unavailable:
- Use default section boundaries
- Use default section colors
- Log warning for debugging

## Tooltip Updates

### Capability Node Tooltip
Current: "Maturity: Genesis"
New: "Maturity: Genesis (12)"

### Navigation Tree Hover
Show same format: "Section Name (value)"

## Acceptance Criteria

### Color Mapping
- [ ] Values 0-24 render in Genesis color range (reds)
- [ ] Values 25-49 render in Custom Built color range (oranges)
- [ ] Values 50-74 render in Product color range (yellows)
- [ ] Values 75-99 render in Commodity color range (greens)
- [ ] Lower values in section render lighter
- [ ] Higher values in section render more saturated

### CapabilityNode
- [ ] Node fill color based on maturity value
- [ ] Color reflects gradient position within section
- [ ] Tooltip shows "Section Name (value)"
- [ ] Works with custom color scheme (ignores maturity)
- [ ] Works with classic color scheme (ignores maturity)

### NavigationTree
- [ ] Color indicators use new maturity colors
- [ ] Gradient reflects position within section
- [ ] Hover shows section name and value

### CapabilityDetails
- [ ] Badge shows section name
- [ ] Numeric value displayed in parentheses
- [ ] Badge color matches section

### Configuration Updates
- [ ] Visualizations update when config changes
- [ ] No page refresh required
- [ ] Section names update in tooltips
- [ ] Colors remain consistent with sections

### Custom Section Names
- [ ] Renamed sections appear in tooltips
- [ ] Renamed sections appear in details panel
- [ ] Color mapping unaffected by name changes

### Boundary Changes
- [ ] Adjusted boundaries affect color thresholds
- [ ] Capabilities re-color based on new boundaries
- [ ] No capability data changes required

## Edge Cases

### Pre-Migration Capabilities
Capabilities without `maturityValue` derive value from `maturityLevel` string using midpoints (same logic as Spec 096).

### Missing Configuration
If API unavailable, use hardcoded default boundaries and colors. Visualizations remain functional.

### Invalid Maturity Values
Clamp values outside 0-99 to valid range for color calculation.

## Performance Considerations

- Pre-calculate lookup table for all 100 values when config loads
- Cache configuration (not fetched per-node)
- Color lookup is O(1) via lookup table
- Re-renders minimized - only when config or capability data changes

## Accessibility Considerations

- Tooltips provide text information (section name + value) so color is not the only indicator
- Color palette chosen for reasonable distinction between sections
- Future enhancement: consider patterns/textures for colorblind users

## Out of Scope
- Custom color scheme configuration by users
- High contrast accessibility mode
- Animated color transitions on value change
- Maturity trend indicators
- Comparative maturity views

## Checklist
- [ ] Specification ready
- [ ] Color constants moved to shared file
- [ ] Color interpolation utilities created (HSL-based)
- [ ] Color lookup table with memoization
- [ ] `useMaturityColorScale()` hook created
- [ ] CapabilityNode updated for gradient coloring
- [ ] Navigation panel color indicators updated
- [ ] CapabilityDetails badge updated
- [ ] Tooltips show "Section Name (value)"
- [ ] nodeFactory includes maturityValue
- [ ] Fallback to default colors when config unavailable
- [ ] Pre-migration capability handling (shared with Spec 096)
- [ ] Capability type includes `maturityValue` and `maturitySection`
- [ ] Unit tests for color utilities
- [ ] Performance verified on large canvas
- [ ] User sign-off
