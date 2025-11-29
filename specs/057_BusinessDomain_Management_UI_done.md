# Business Domain Management UI

## Description
React frontend for CRUD operations on business domains and managing capability associations.

## Purpose
Enable enterprise architects to create, update, and delete business domains and associate L1 capabilities with them.

## Dependencies
- Spec 056: Business Domain REST API

## Pages

### BusinessDomainsPage
Main page listing all business domains with actions.

**URL:** `/business-domains`

**Features:**
- Display all business domains in grid/list view
- Show domain name, description, and capability count
- Create new domain button
- Edit domain button per item
- Delete domain button per item
- Navigate to domain detail for capability management
- Navigate to visualization page
- Empty state when no domains exist

**State Management:**
- Use React Query `useBusinessDomains()` hook
- Optimistic updates for delete operations
- Error handling with toast notifications

### DomainDetailPage
Page for editing domain and managing capability associations.

**URL:** `/business-domains/{id}`

**Features:**
- Edit domain name and description inline
- Show associated capabilities with remove action
- Add capabilities via modal selector
- Display orphaned L1 warning if applicable
- Breadcrumb navigation back to list
- Save/cancel actions for edits

**State Management:**
- Use React Query `useBusinessDomain(selfLink)` hook
- Use `useDomainCapabilities(capabilitiesLink)` hook
- Optimistic updates for associations

## Components

### DomainList
Displays business domains in grid or list format.

**Props:**
- `domains`: Array of BusinessDomain objects
- `onEdit`: Callback when edit clicked
- `onDelete`: Callback when delete clicked
- `onView`: Callback when domain clicked

**Behavior:**
- Render domain cards with name, description, capability count
- Show action buttons (edit, delete, view)
- Hide delete button if domain has capabilities
- Handle loading and error states

### DomainCard
Individual domain display card.

**Props:**
- `domain`: BusinessDomain object
- `onEdit`: Edit callback
- `onDelete`: Delete callback
- `onView`: View callback

**Behavior:**
- Display domain name as heading
- Show description (truncated if long)
- Show capability count badge
- Render HATEOAS action buttons based on available links
- Confirm before delete

### DomainForm
Form for creating or editing a domain.

**Props:**
- `domain`: BusinessDomain object (optional for create)
- `mode`: "create" | "edit"
- `onSubmit`: Callback with form data
- `onCancel`: Cancel callback

**Behavior:**
- Text input for name (required, max 100 chars)
- Textarea for description (optional, max 500 chars)
- Client-side validation before submit
- Display validation errors from API
- Disable submit while saving

**Validation:**
- Name must not be empty
- Name max 100 characters
- Description max 500 characters
- Show field-level errors

### CapabilityAssociationManager
Manages capability associations for a domain.

**Props:**
- `domainId`: Business domain ID
- `capabilitiesLink`: HATEOAS link to domain capabilities

**Behavior:**
- Display currently associated capabilities as tags
- Remove button per capability
- "Add Capabilities" button opens modal selector
- Empty state when no capabilities
- Confirm before remove

**State Management:**
- Use `useDomainCapabilities(capabilitiesLink)` hook
- Use `useAssociateCapability()` mutation
- Use `useDissociateCapability()` mutation

### CapabilitySelectorModal
Modal for selecting capabilities to associate.

**Props:**
- `isOpen`: Boolean
- `onClose`: Close callback
- `domainId`: Business domain ID
- `currentAssociations`: Array of capability IDs

**Behavior:**
- Display L1 capabilities in hierarchical tree
- Checkbox per L1 capability
- Show which capabilities are already associated (disabled checkboxes)
- Highlight orphaned capabilities
- Save button applies changes
- Cancel button discards changes

**State Management:**
- Local state for selection
- Use `useCapabilityTree()` hook for full hierarchy
- Bulk associate/dissociate on save

### CapabilityTree
Hierarchical tree view of capabilities for selection.

**Props:**
- `tree`: Array of CapabilityTreeNode
- `selectedIds`: Set of selected capability IDs
- `onToggle`: Callback when checkbox toggled
- `disabledIds`: Set of disabled capability IDs (optional)

**Behavior:**
- Render L1 capabilities with checkboxes
- Show child capabilities (L2/L3) as nested items (read-only)
- Highlight orphaned L1 capabilities with warning icon
- Expand/collapse tree nodes
- Only L1 capabilities are selectable

### CapabilityTagList
Display associated capabilities as removable tags.

**Props:**
- `capabilities`: Array of Capability objects
- `onRemove`: Callback with capability to remove

**Behavior:**
- Render capability code and name as tag
- Show remove icon on hover
- Confirm before remove
- Empty state if no capabilities

## Routing

Add to React Router configuration:
```
/business-domains -> BusinessDomainsPage
/business-domains/new -> DomainDetailPage (create mode)
/business-domains/:id -> DomainDetailPage (edit mode)
/business-domains/visualization -> DomainVisualizationPage (see spec 058)
```

## API Integration

### Custom Hooks

**useBusinessDomains()**
- Fetches all business domains
- Returns: `{ data, isLoading, error, refetch }`
- Query key: `['businessDomains']`

**useBusinessDomain(selfLink)**
- Fetches single business domain
- Returns: `{ data, isLoading, error }`
- Query key: `['businessDomain', selfLink]`

**useCreateDomain()**
- Creates new business domain
- Invalidates `['businessDomains']` on success
- Returns: `{ mutate, isLoading, error }`

**useUpdateDomain()**
- Updates existing domain
- Invalidates domain queries on success
- Returns: `{ mutate, isLoading, error }`

**useDeleteDomain()**
- Deletes business domain
- Invalidates `['businessDomains']` on success
- Returns: `{ mutate, isLoading, error }`

**useDomainCapabilities(capabilitiesLink)**
- Fetches capabilities for a domain
- Returns: `{ data, isLoading, error }`
- Query key: `['domainCapabilities', capabilitiesLink]`

**useAssociateCapability()**
- Associates capability with domain
- Invalidates capability queries on success
- Returns: `{ mutate, isLoading, error }`

**useDissociateCapability()**
- Removes capability from domain
- Invalidates capability queries on success
- Returns: `{ mutate, isLoading, error }`

## TypeScript Interfaces

**BusinessDomain**
```typescript
interface BusinessDomain {
  id: string;
  name: string;
  description: string;
  capabilityCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}
```

**HATEOASLinks**
```typescript
interface HATEOASLinks {
  self: string;
  capabilities?: string;
  update?: string;
  delete?: string;
  collection?: string;
}
```

**Capability (in domain context)**
```typescript
interface Capability {
  id: string;
  code: string;
  name: string;
  description: string;
  level: "L1" | "L2" | "L3" | "L4";
  assignedAt?: string;
  _links: HATEOASLinks;
}
```

## User Interactions

**Creating a Domain:**
1. Click "Create Domain" button
2. Fill in form (name required, description optional)
3. Submit creates domain via POST
4. Navigate to detail page on success
5. Show error toast on failure

**Editing a Domain:**
1. Click domain or edit button
2. Modify name/description in form
3. Save updates domain via PUT
4. Show success toast
5. Show error toast on conflict/validation

**Deleting a Domain:**
1. Click delete button (only shown if no capabilities)
2. Confirm deletion in dialog
3. Delete via DELETE request
4. Remove from list on success
5. Show error toast on failure

**Adding Capabilities:**
1. Click "Add Capabilities" button
2. Modal opens with capability tree
3. Select L1 capabilities via checkboxes
4. Save associates capabilities via POST requests
5. Modal closes and list refreshes

**Removing Capabilities:**
1. Click remove icon on capability tag
2. Confirm removal in dialog
3. Dissociate via DELETE request
4. Remove from list on success

## Implementation Notes
- Follow HATEOAS links from API responses for all actions
- Use React Query for server state management
- Optimistic updates for better UX
- Toast notifications for all success/error feedback
- Confirm dialogs for destructive actions
- Loading states for all async operations
- Client-side validation before API calls
- Display API validation errors on form fields
