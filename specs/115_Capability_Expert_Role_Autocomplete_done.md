# Capability Expert Role Autocomplete and Removal

## Description
Add role autocomplete support and expert removal for capability experts, matching the pattern implemented for application component experts.

## Purpose
Enable role suggestions when adding experts to capabilities, and allow removal of experts. Improves UX consistency across entity types. Role types build up organically from previously used roles within the capability context.

## Dependencies
- Spec 024: Capability Metadata (existing Expert implementation)
- Spec 114: Application Component Experts (pattern reference)

## Architecture

### Domain Model Extension

**Capability Aggregate:**
- Add `RemoveExpert(expertName)` method raising `CapabilityExpertRemoved` event

**New Event:** `CapabilityExpertRemoved`
- Fields: CapabilityID, ExpertName, RemovedAt

### Read Model Extension

**Extend `CapabilityReadModel`:**
- Add `GetDistinctExpertRoles(ctx) []string` method
- Add `RemoveExpert(ctx, capabilityID, expertName)` method
- Query: `SELECT DISTINCT expert_role FROM capability_experts WHERE tenant_id = $1 ORDER BY expert_role`

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/capabilities/expert-roles` | Get distinct roles for autocomplete |
| `DELETE` | `/capabilities/{id}/experts/{name}` | Remove expert from capability |

### Frontend

Replace `TextInput` with `Autocomplete` component in `AddExpertDialog.tsx` for capabilities.

## Behaviour

### Role Autocomplete Suggests Previously Used Roles

**Given** a capability has an expert with role "Product Owner"
**And** another capability has an expert with role "Domain Expert"
**When** I start adding an expert to any capability
**Then** both "Product Owner" and "Domain Expert" are suggested in role autocomplete

### Custom Roles Can Be Created

**Given** no expert with role "Security Champion" exists on any capability
**When** I add an expert with role "Security Champion"
**Then** the expert is saved with role "Security Champion"
**And** "Security Champion" becomes available for autocomplete on future capability experts

### Role Vocabularies Are Separate Per Entity Type

**Given** an application component has an expert with role "Tech Lead"
**And** no capability has an expert with that role
**When** I add an expert to a capability
**Then** "Tech Lead" is NOT suggested in role autocomplete

### Removing an Expert

**Given** capability "Customer Management" has expert "Alice Smith"
**When** I remove the expert "Alice Smith"
**Then** the expert no longer appears in the capability's expert list
**But** the role "Product Owner" remains available for autocomplete (from other usages or history)

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] API Documentation updated in OpenAPI specification
- [x] User sign-off
