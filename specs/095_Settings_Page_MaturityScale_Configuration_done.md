# Settings Page with Maturity Scale Configuration

## Description
Introduce a tenant-level Settings page accessible to administrators for managing maturity scale configuration. The page should be designed to accommodate future settings sections.

## Purpose
Enable tenant administrators to customize the maturity scale section names and boundaries through a dedicated settings interface, making the tool's terminology match their organization's vocabulary.

## Dependencies
- Spec 090: MetaModel Bounded Context
- Spec 091: Maturity Scale Configuration Aggregate
- Spec 092: Maturity Scale REST API

## User Need
As a tenant administrator, I need to customize how maturity sections are named and divided so the evolution scale terminology matches my organization's vocabulary and mental models.

## Success Criteria
- Administrators can access a Settings page from main navigation
- Administrators can view current maturity scale configuration
- Administrators can rename any of the 4 sections
- Administrators can adjust section boundaries while maintaining coverage of 0-99
- Administrators can reset to default Wardley mapping scale
- Non-administrators cannot access the Settings page
- Changes are persisted and immediately reflected across the application

## Navigation Integration

### Route
`/settings` - single page initially, structure for future expansion when additional settings are added.

### Permission
Requires `settings:manage` permission (new permission to be added to admin role). Follows existing permission naming pattern (e.g., `users:manage`).

### Navigation Item
Add "Settings" to main navigation after Users/Invitations. Requires adding `'settings'` to the `AppView` type. Visible only to users with `settings:manage` permission.

## Page Structure

### Initial Layout
Single settings page showing maturity scale configuration. Design should allow adding a sidebar with categories when future settings are needed (Strategy Pillars, Element Types).

### Maturity Scale Section

#### View Mode
- Horizontal bar visualization showing 4 sections spanning 0-99
- Each section displays: name, range (e.g., "0-24")
- Visual indicator when using default configuration
- "Edit" button to enter edit mode
- "Reset to Defaults" button (hidden when already using defaults)

#### Edit Mode
- Inline editable section names
- Draggable boundary handles between sections
- Real-time preview as user adjusts boundaries
- Inline validation feedback for invalid states
- "Save" and "Cancel" buttons
- Optimistic locking using version field from API

## Validation Rules

Domain validation is authoritative (backend). Frontend provides immediate feedback for basic rules to improve UX.

### Client-Side Validation (Immediate Feedback)
- Section names: non-empty, max 50 characters
- Boundaries: values 0-99, contiguous ranges

### Server-Side Validation (Authoritative)
- All invariants enforced by domain model
- API returns detailed error messages for invalid states

### Validation Feedback
- Inline error messages next to invalid fields
- Save button disabled while validation errors exist
- API 400 errors displayed inline with field-specific messages

## API Integration

Uses endpoints defined in Spec 092:
- `GET /api/v1/meta-model/maturity-scale` - fetch current configuration
- `PUT /api/v1/meta-model/maturity-scale` - update configuration (include version for optimistic locking)
- `POST /api/v1/meta-model/maturity-scale/reset` - reset to defaults (requires confirmation dialog)

## State Management

### Query Hook
Create `useMaturityScale()` hook (follows existing naming pattern: `useMaturityLevels`, `useStatuses`).

### Cache Invalidation
On successful update/reset, invalidate:
- Maturity scale config query
- Maturity levels metadata query (used by capability editing and canvas)

Add query key to `queryKeys` in `queryClient.ts` following existing pattern.

## Error Handling

### Version Conflict (409)
Display message: "Configuration was modified by another user. Please refresh and try again." with Refresh button.

### Validation Errors (400)
Display field-specific errors inline from API response.

### Network Errors
Display toast error with retry option (follow existing `react-hot-toast` pattern).

## Accessibility

- All form inputs have associated labels
- Keyboard navigation for boundary editor (arrow keys to adjust values, Tab between sections, Escape to cancel)
- ARIA labels for boundary handles describing current value and valid range
- Focus management: focus first editable field when entering edit mode
- Live region for validation errors (announced to screen readers)

## Acceptance Criteria

### Navigation
- [ ] Settings nav item appears for users with `meta-model:manage` permission
- [ ] Settings nav item is hidden for users without permission
- [ ] Clicking Settings navigates to `/settings`
- [ ] Default redirect to `/settings/maturity-scale`

### View Configuration
- [ ] Page loads and displays current maturity scale configuration
- [ ] Shows visual bar representation of 4 sections
- [ ] Displays section names and ranges
- [ ] Shows "Default" badge when using default configuration
- [ ] Loading state shown while fetching

### Edit Configuration
- [ ] Edit button enters edit mode
- [ ] Section names are editable inline
- [ ] Boundary handles are draggable
- [ ] Changes preview in real-time
- [ ] Cancel discards changes and exits edit mode
- [ ] Save persists changes and exits edit mode
- [ ] Success toast shown on save

### Reset to Defaults
- [ ] Reset button shown when not using defaults
- [ ] Confirmation dialog before reset
- [ ] Reset restores Genesis/Custom Built/Product/Commodity with equal ranges
- [ ] Success toast shown on reset

### Validation
- [ ] Empty section name shows error
- [ ] Section name over 50 chars shows error
- [ ] Non-contiguous boundaries show error
- [ ] Save button disabled with validation errors

### Error Handling
- [ ] Version conflict shows refresh prompt
- [ ] Network error shows retry option
- [ ] API validation errors display inline

## Out of Scope
- Strategy Pillars configuration (future section)
- Element Types configuration (future section)
- Relationship Rules configuration (future section)
- Bulk import/export of settings
- Settings change history/audit log UI

## Checklist
- [ ] Specification ready
- [ ] Permission `settings:manage` added to backend (admin role)
- [ ] `'settings'` added to `AppView` type
- [ ] Settings route added to router
- [ ] Settings nav item added with permission check
- [ ] Settings page component created
- [ ] Maturity scale view mode component
- [ ] Maturity scale edit mode with boundary editor
- [ ] `useMaturityScale()` hook created
- [ ] Query key added to `queryKeys` in `queryClient.ts`
- [ ] Cache invalidation configured
- [ ] Client-side validation implemented
- [ ] Error handling with toast notifications
- [ ] Loading states implemented
- [ ] Keyboard navigation and accessibility
- [ ] Unit tests for components
- [ ] Integration tests for settings flow
- [ ] User sign-off
