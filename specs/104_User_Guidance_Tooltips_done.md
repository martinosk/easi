# User Guidance and Tooltips

**Status**: done

## User Value

> "As a new user, I want to understand what each visual indicator means without having to consult external documentation, so I can immediately interpret dashboards and make informed decisions."

> "As an enterprise architect, I want contextual help explaining domain concepts like maturity sections and investment priorities, so I can efficiently use the analysis tools."

## Dependencies

None

---

## Scope

Add contextual help throughout the application to explain:
- Color coding schemes (maturity sections, investment priorities)
- Domain-specific terminology (maturity gap, strategic fit, implementations)
- Icons and visual indicators
- Action hints (drag-and-drop targets, available interactions)

---

## UI Components

### Tooltip Component

Create a reusable `HelpTooltip` component using Mantine's Tooltip:

```
HelpTooltip
├── icon (question mark or info icon)
├── label (short text for inline display)
└── tooltip content (detailed explanation on hover)
```

Usage patterns:
- Inline with labels: "Maturity Gap [?]"
- Icon-only for space-constrained areas
- Rich content with examples where helpful

### Help Icon

Small info icon (ⓘ) that triggers tooltip on hover. Consistent styling across the app.

---

## Areas Requiring Guidance

### Maturity Analysis Tab

**Summary Stats Section**
- "Capabilities" → "Enterprise capabilities with 2+ linked implementations that can be analyzed for maturity variance"
- "Implementations" → "Total domain capabilities linked to these enterprise capabilities"
- "Avg Gap" → "Average difference between implementation maturity and target (or highest implementation)"

**Maturity Distribution Bar Legend**
Add a legend explaining:
- Purple (Genesis, 0-24): Early-stage, experimental capabilities
- Blue (Custom Build, 25-49): Internally developed, customized solutions
- Green (Product, 50-74): Commercial or standardized products
- Gray (Commodity, 75-99): Utility services, fully commoditized

**Candidate Card Fields**
- "Target Maturity" → "The desired maturity level all implementations should reach"
- "Implementations" → "Number of domain capabilities linked to this enterprise capability"
- "Domains" → "Number of distinct business domains containing implementations"
- "Max Gap" → "Largest maturity difference from target among all implementations"

### Maturity Gap Detail Panel

**Investment Priority Sections**
- "High Priority (Gap > 40)" → "Significant investment needed to bring these implementations to target maturity"
- "Medium Priority (Gap 15-40)" → "Moderate investment required for maturity improvement"
- "Low Priority (Gap 1-14)" → "Minor improvements needed to reach target"
- "On Target" → "These implementations meet or exceed the target maturity"

**Target Marker on Bars**
- Explain the vertical line marker represents the target maturity level

### Enterprise Capabilities Tab

**Enterprise Capability Cards**
- "Links" → "Number of domain capabilities linked to this enterprise capability"
- "Domains" → "Number of business domains containing linked capabilities"
- Drag-drop hint: "Drag domain capabilities from the right panel to link them"

### Unlinked Capabilities Tab

**Purpose explanation**
- Header tooltip: "Domain capabilities not yet associated with any enterprise capability. Link them to enable maturity analysis."

### Strategy Pillars (when implemented)

**Importance Rating**
- Star ratings → "How critical this capability is to achieving the pillar's strategic goals (1-5)"

**Fit Score**
- Dot ratings → "How well the application supports this pillar's requirements (1-5)"

**Strategic Liability**
- Warning indicator → "The application's fit is significantly lower than the capability's importance"

---

## Implementation Approach

### Mantine Tooltip Usage

Use Mantine's `Tooltip` component with consistent configuration:
- Position: top or right (context-dependent)
- Width: 250-300px for detailed explanations
- Multiline enabled for longer content

### HelpTooltip Component

```tsx
interface HelpTooltipProps {
  content: React.ReactNode;
  label?: string;
  iconOnly?: boolean;
}
```

Place in `frontend/src/components/shared/HelpTooltip.tsx`

### Styling

- Icon color: subtle gray (#9CA3AF) that darkens on hover
- Icon size: 14-16px to not distract from primary content
- Consistent spacing from associated label

---

## Specific Additions

### MaturityAnalysisTab.tsx

1. Add legend component below summary stats explaining maturity section colors
2. Add help icons to stat labels

### MaturityGapDetailPanel.tsx

1. Add tooltip explaining the bar chart visualization
2. Add help icons to priority section headers

### EnterpriseCapabilityCard.tsx

1. Add drag-drop hint when no linked capabilities exist
2. Add help icons to Links/Domains labels

### UnlinkedCapabilitiesTab.tsx

1. Add explanatory header with tooltip

### Settings Pages

1. Add help icons to maturity scale configuration
2. Add help icons to strategy pillar configuration

---

## Checklist

- [x] Create HelpTooltip component with Mantine Tooltip
- [x] Add maturity section color legend to Maturity Analysis tab
- [x] Add tooltips to summary stats in Maturity Analysis
- [x] Add tooltips to candidate card fields
- [x] Add tooltips to priority section headers in Gap Detail
- [x] Add drag-drop guidance to Enterprise Capability cards
- [x] Add explanatory header to Unlinked Capabilities tab
- [x] Add help icons to Settings pages where applicable
- [x] Tests passing
