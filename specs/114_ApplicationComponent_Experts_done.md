# Application Component Experts

## Description
Add subject matter experts as metadata to application components, with autocomplete support for role types.

## Purpose
Enable tracking of who has domain knowledge about specific application components. Role types build up organically within each entity type (application components and capabilities maintain separate role vocabularies).

## Dependencies
- Spec 024: Capability Metadata (pattern reference for Expert entity)

## Architecture

### Domain Model

**Bounded Context:** `architecturemodeling`

**Expert Value Object:** `backend/internal/architecturemodeling/domain/entities/expert.go`
- Immutable value object with `name`, `role`, `contact`, `addedAt`
- Constructor validation for all required fields
- Separate from capability's Expert (no shared kernel)

**ApplicationComponent Aggregate Extension:**
- Add `experts []*entities.Expert` field
- Add `AddExpert(expert)` method raising `ApplicationComponentExpertAdded` event
- Add `RemoveExpert(expertName)` method raising `ApplicationComponentExpertRemoved` event

### Event Sourcing Elements

| Element | Name |
|---------|------|
| Command | `AddApplicationComponentExpert` |
| Command | `RemoveApplicationComponentExpert` |
| Event | `ApplicationComponentExpertAdded` |
| Event | `ApplicationComponentExpertRemoved` |

### Read Model

**New Table: `application_component_experts`**
- Primary key: `(tenant_id, component_id, expert_name)`
- Index on `(tenant_id, component_id)` for listing
- Index on `(tenant_id, expert_role)` for role autocomplete

**Role Autocomplete:** Query-based via `SELECT DISTINCT expert_role` - roles build up organically from expert assignments.

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/components/{id}/experts` | Add expert to component |
| `DELETE` | `/components/{id}/experts/{name}` | Remove expert from component |
| `GET` | `/components/expert-roles` | Get distinct roles for autocomplete |

Experts are included in component details via `GET /components/{id}`.

## Behaviour

### Adding an Expert to an Application Component

**Given** an application component "Customer Portal" exists
**When** I add an expert with name "Alice Smith", role "Product Owner", contact "alice@example.com"
**Then** the expert appears in the component's expert list
**And** the role "Product Owner" becomes available for autocomplete on future application component experts

### Role Autocomplete Suggests Previously Used Roles

**Given** an application component has an expert with role "Product Owner"
**And** another application component has an expert with role "Tech Lead"
**When** I start adding an expert to any application component
**Then** both "Product Owner" and "Tech Lead" are suggested in role autocomplete

### Custom Roles Can Be Created

**Given** no expert with role "Security Champion" exists on any application component
**When** I add an expert with role "Security Champion"
**Then** the expert is saved with role "Security Champion"
**And** "Security Champion" becomes available for autocomplete on future application component experts

### Role Vocabularies Are Separate Per Entity Type

**Given** a capability has an expert with role "Domain Expert"
**And** no application component has an expert with that role
**When** I add an expert to an application component
**Then** "Domain Expert" is NOT suggested in role autocomplete
**But** if I add an expert to a capability, "Domain Expert" IS suggested

### Removing an Expert

**Given** application component "Customer Portal" has expert "Alice Smith"
**When** I remove the expert "Alice Smith"
**Then** the expert no longer appears in the component's expert list
**But** the role "Product Owner" remains available for autocomplete (from other usages or history)

### Listing Experts

**Given** application component "Customer Portal" has experts:
  | Name         | Role          | Contact              |
  | Alice Smith  | Product Owner | alice@example.com    |
  | Bob Johnson  | Tech Lead     | bob@example.com      |
**When** I view the component details
**Then** I see both experts with their roles and contact information

### Validation

**Given** I am adding an expert to an application component
**When** I submit with an empty name
**Then** the operation fails with error "expert name cannot be empty"

**Given** I am adding an expert to an application component
**When** I submit with an empty role
**Then** the operation fails with error "expert role cannot be empty"

**Given** I am adding an expert to an application component
**When** I submit with an empty contact
**Then** the operation fails with error "expert contact cannot be empty"

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API Documentation updated in OpenAPI specification
- [x] User sign-off
