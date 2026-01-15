# Import Metadata Pre-fill

## Description
Allow users to pre-fill metadata fields when importing architecture models, applying the selected values to all imported elements of each type.

## Purpose
When importing large architecture models, users need a way to assign common metadata (like EA Owner) to all imported capabilities and application components without having to edit each one individually after import.

## Dependencies
- Spec 061: ImportOpenExchange (existing import flow)
- Spec 024: Capability Metadata (EA Owner field)

## Architecture

### Import Session Extension

**Bounded Context:** `importing`

**ImportSession Aggregate Extension:**
- Add `capabilityEAOwner` field (optional string - user ID)

### API Changes

**POST /imports** (Create Import Session)
- Add optional `capabilityEAOwner` field to request body

### Orchestrator Changes

When `capabilityEAOwner` is provided:
- After creating each capability via `CreateCapability` command
- Issue `UpdateCapabilityMetadata` command with the pre-filled EA Owner

### UI Changes

**ImportUploadStep Component:**
- Add optional EA Owner selection dropdowns after file upload
- Use existing `useEAOwnerCandidates()` hook for user options
- Fields are clearable (user can leave blank for no pre-fill)

## Behaviour

### Pre-filling EA Owner for Capabilities

**Given** I am importing an ArchiMate model with 5 capabilities
**And** the import upload step shows an "EA Owner for Capabilities" dropdown
**When** I select "Alice Smith" as the EA Owner
**And** I complete the import
**Then** all 5 imported capabilities have "Alice Smith" set as their EA Owner

### Leaving EA Owner Blank

**Given** I am importing an ArchiMate model with 3 capabilities
**When** I leave the "EA Owner for Capabilities" dropdown blank
**And** I complete the import
**Then** all 3 imported capabilities have no EA Owner set
**And** users can assign EA Owners individually later

### EA Owner Dropdown Shows Available Users

**Given** the system has users:
  | Name         | Role      |
  | Alice Smith  | architect |
  | Bob Johnson  | admin     |
  | Carol Davis  | viewer    |
**When** I open the "EA Owner for Capabilities" dropdown during import
**Then** I see "Alice Smith" and "Bob Johnson" as options
**But** I do not see "Carol Davis" (viewers cannot be EA Owners)

### Import Preview Shows Pre-fill Selection

**Given** I have uploaded an ArchiMate file
**And** I have selected "Alice Smith" as EA Owner for Capabilities
**When** I proceed to the preview step
**Then** the preview indicates that imported capabilities will have "Alice Smith" as EA Owner

### Pre-fill Persists Through Import Confirmation

**Given** I have configured EA Owner pre-fill to "Alice Smith"
**And** I am on the import preview step
**When** I click "Confirm Import"
**Then** the import processes with the configured EA Owner
**And** the progress step shows the metadata assignment phase

### Validation: Invalid User ID

**Given** I submit an import request with an invalid user ID for EA Owner
**When** the system validates the request
**Then** the import fails with error "invalid EA Owner: user not found"

## UI Considerations

- EA Owner dropdowns should appear in a collapsible "Import Options" section below the file upload
- Dropdowns should be searchable for tenants with many users
- Clear visual indication of optional nature (e.g., placeholder text "Select EA Owner (optional)")
- Selected values should be visible in the preview step summary

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] API Documentation updated in OpenAPI specification
- [ ] User sign-off
