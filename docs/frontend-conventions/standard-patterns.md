# Frontend Standard Patterns

## HATEOAS-Driven UI

The frontend follows a HATEOAS driven approach where the backend controls what actions are available to users through hypermedia links in API responses.

### Core Principle

**Never hardcode action availability in the frontend.** The backend is the single source of truth for what a user can do with a resource. The frontend simply renders UI controls based on the presence or absence of HATEOAS links.

### Link Structure

API responses include a `_links` object with available actions:

```typescript
interface HATEOASLink {
  href: string;
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
}

interface HATEOASLinks {
  self?: HATEOASLink;
  edit?: HATEOASLink;
  delete?: HATEOASLink;
  collection?: HATEOASLink;
  [key: string]: HATEOASLink | undefined;
}
```

### HATEOAS Utility Functions

Use the utility functions from `src/utils/hateoas.ts`:

```typescript
import { hasLink, getLink, canEdit, canDelete } from '../utils/hateoas';

// Check if an action is available
if (hasLink(resource, 'edit')) {
  // Show edit button
}

// Get the URL for an action
const editUrl = getLink(resource, 'edit');

// Convenience helpers
if (canEdit(resource)) { /* ... */ }
if (canDelete(resource)) { /* ... */ }
```

### Standard Link Relations

| Relation | Purpose | HTTP Method |
|----------|---------|-------------|
| `self` | Current resource URL | GET |
| `edit` | Update the resource | PUT |
| `delete` | Delete the resource | DELETE |
| `collection` | Parent collection | GET |
| `create` | Create a new child resource | POST |
| `up` | Parent resource (hierarchy) | GET |

### Custom Link Relations

Custom relations are prefixed with `x-`:

| Relation | Purpose |
|----------|---------|
| `x-children` | Child resources |
| `x-remove` | Remove from a relationship |
| `x-create-link` | Create a link/association |

### Conditional UI Rendering

Always gate UI actions on link presence:

```tsx
// CORRECT - UI controlled by backend
function ResourceCard({ resource }: { resource: Resource }) {
  const canEditResource = resource._links?.edit !== undefined;
  const canDeleteResource = resource._links?.delete !== undefined;

  return (
    <div>
      <h3>{resource.name}</h3>
      {canEditResource && (
        <button onClick={handleEdit}>Edit</button>
      )}
      {canDeleteResource && (
        <button onClick={handleDelete}>Delete</button>
      )}
    </div>
  );
}

// INCORRECT - Hardcoded permission logic
function ResourceCard({ resource, userRole }: Props) {
  // DON'T DO THIS - duplicates backend logic
  const canEdit = userRole === 'admin' || resource.ownerId === currentUserId;
  // ...
}
```

### Drag and Drop with HATEOAS

For drag-and-drop operations, check the appropriate link before accepting:

```tsx
function DropTarget({ target, onDrop }: Props) {
  const canAcceptDrop = target._links?.['x-create-link'] !== undefined;

  const handleDragOver = (e: React.DragEvent) => {
    if (!canAcceptDrop) return;
    e.preventDefault();
  };

  const handleDrop = (e: React.DragEvent) => {
    if (!canAcceptDrop) return;
    e.preventDefault();
    onDrop(/* ... */);
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      style={{ opacity: canAcceptDrop ? 1 : 0.5 }}
    >
      {/* ... */}
    </div>
  );
}
```

### API Calls Using Links

When making API calls, prefer using the href from links:

```typescript
// CORRECT - Use the link href
async function updateResource(resource: Resource, data: UpdateData) {
  const editLink = resource._links?.edit;
  if (!editLink) throw new Error('Edit not permitted');

  return fetch(editLink.href, {
    method: editLink.method,
    body: JSON.stringify(data),
  });
}

// ACCEPTABLE - Construct URL when link not available
async function fetchResource(id: string) {
  return fetch(`/api/v1/resources/${id}`);
}
```

### Type Definitions

Resource types should include `_links`:

```typescript
interface Capability {
  id: CapabilityId;
  name: string;
  description?: string;
  _links: HATEOASLinks;  // Required for resources
}

interface ViewCapability {
  capabilityId: CapabilityId;
  x: number;
  y: number;
  _links?: HATEOASLinks;  // Optional for embedded resources
}
```

### Testing HATEOAS Behavior

Test that UI correctly responds to link presence/absence:

```typescript
describe('CapabilityDetails', () => {
  it('shows edit button when edit link is present', () => {
    const capability = {
      id: '123',
      name: 'Test',
      _links: {
        self: { href: '/api/v1/capabilities/123', method: 'GET' },
        edit: { href: '/api/v1/capabilities/123', method: 'PUT' },
      },
    };

    render(<CapabilityDetails capability={capability} />);
    expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument();
  });

  it('hides edit button when edit link is absent', () => {
    const capability = {
      id: '123',
      name: 'Test',
      _links: {
        self: { href: '/api/v1/capabilities/123', method: 'GET' },
        // No edit link
      },
    };

    render(<CapabilityDetails capability={capability} />);
    expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument();
  });
});
```
