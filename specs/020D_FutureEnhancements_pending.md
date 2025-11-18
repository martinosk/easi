# Future Enhancements - Delete Operations

## Description
Future enhancements for delete operations including OpenAPI documentation, HATEOAS links, relation deletion from views, and additional accessibility features.

## Purpose
This spec tracks enhancements that were identified during implementation of specs 020A, 020B, and 020C but were deemed not critical for the initial release.

## Dependencies
- Spec 020A: Delete Operations - Backend Domain Model (completed)
- Spec 020B: Cascade Deletion - Cross-Context Integration (completed)
- Spec 020C: Context Menu Operations - Frontend (completed)

## Enhancements

### 1. OpenAPI Specification and Documentation

**From Spec 020A:**

#### DELETE /api/v1/components/{id} Documentation
- Add endpoint definition to OpenAPI spec
- Document request parameters
- Document response codes (204, 404, 500)
- Include example requests and responses
- Document cascade deletion behavior in description

#### DELETE /api/v1/relations/{id} Documentation
- Add endpoint definition to OpenAPI spec
- Document request parameters
- Document response codes (204, 404, 500)
- Include example requests and responses

#### HATEOAS Links Documentation
- Update component schema to include delete link
- Update relation schema to include delete link
- Document link relations in OpenAPI spec
- Add examples of HATEOAS responses

#### OpenAPI Generation
- Update OpenAPI spec generation script
- Ensure DELETE endpoints are included in generated spec
- Validate generated spec against OpenAPI 3.0 standards

**Why Deferred:**
- Core delete functionality works without OpenAPI documentation
- HATEOAS links can be added incrementally
- OpenAPI documentation is valuable but not blocking user functionality

### 2. Remove Relation from View Operations

**From Spec 020B:**

#### Backend: RemoveRelationFromView Command
- Purpose: Remove relation from specific view only (not from model)
- Input: View ID, Relation ID
- Result: RelationRemovedFromView event raised
- Used when user explicitly removes relation from view (not from model)
- Command handler implementation in ArchitectureViews context

#### Backend: RelationRemovedFromView Event
- Purpose: Signal that relation was removed from specific view
- Data: View ID, Relation ID, timestamp
- Raised when user removes relation from view (not model)
- Read model projector updates view state

#### Backend: Read Model Projector
- Projector handles RelationRemovedFromView event
- Update view state to exclude removed relation
- Maintain relation in other views and model

#### API Endpoint: DELETE /api/v1/views/{viewId}/relations/{relationId}
- Purpose: Remove relation from specific view only (not from model)
- Request Parameters:
  - viewId (path parameter, UUID, required): View identifier
  - relationId (path parameter, UUID, required): Relation identifier
- Success Response: 204 No Content
- Error Responses:
  - 404 Not Found: View or relation not found in view
  - 500 Internal Server Error
- Behavior: Issues RemoveRelationFromView command
- HATEOAS: Relation DTOs in view context include removeFromView link

**From Spec 020C:**

#### Frontend: API Client
- Add deleteRelationFromView method to API client
- Method signature: `deleteRelationFromView(viewId: string, relationId: string): Promise<void>`

#### Frontend: Canvas Context Menu
- Add "Delete from View" option to relation context menu
- Immediate removal without confirmation (non-destructive)
- Toast notification confirms removal
- Uses DELETE /api/v1/views/{viewId}/relations/{relationId} endpoint

**Why Deferred:**
- Current implementation allows deleting relations from the model
- Relations are automatically removed from views when components are removed
- "Delete from view" for relations is a nice-to-have for specific use cases
- Users can work around this by deleting the relation from the model if needed

### 3. Additional Accessibility Features

**From Spec 020C:**

#### Keyboard Shortcut to Open Context Menu
- Implement Shift+F10 keyboard shortcut
- Alternative: Context Menu key (if available)
- Opens context menu on currently focused item
- Works in both tree view and canvas

#### Color Contrast Standards
- Audit all context menu colors for WCAG AA compliance
- Ensure sufficient contrast ratios:
  - Normal text: 4.5:1 minimum
  - Large text: 3:1 minimum
  - UI components: 3:1 minimum
- Update CSS variables if needed

#### Screen Reader Testing
- Complete manual testing with NVDA (Windows)
- Complete manual testing with JAWS (Windows)
- Complete manual testing with VoiceOver (macOS)
- Document any issues found
- Implement fixes for screen reader compatibility

**Why Deferred:**
- Basic accessibility (ARIA roles, keyboard navigation) is already implemented
- Additional features enhance UX but are not blockers
- Can be implemented incrementally based on user feedback

## Business Value

### High Value
- OpenAPI documentation enables third-party integrations
- HATEOAS links make the API more discoverable and self-documenting

### Medium Value
- "Delete relation from view" provides more granular control
- Enhanced accessibility reaches more users

### Low Value
- Screen reader testing is important but can be done iteratively
- Color contrast adjustments are refinements to existing functionality

## Implementation Priority

### Phase 1: OpenAPI and HATEOAS
- Highest ROI for external developers and API consumers
- Relatively low effort to implement
- Improves API discoverability

### Phase 2: Remove Relation from View
- Medium complexity implementation
- Requires backend command, event, and API endpoint
- Requires frontend integration
- Provides additional flexibility for power users

### Phase 3: Enhanced Accessibility
- Ongoing effort requiring manual testing
- Incremental improvements over time
- Important for inclusivity but not blocking core functionality

## Checklist

### OpenAPI Documentation
- [ ] Document DELETE /api/v1/components/{id} endpoint
- [ ] Document DELETE /api/v1/relations/{id} endpoint
- [ ] Document response schemas
- [ ] Add HATEOAS links to component schema
- [ ] Add HATEOAS links to relation schema
- [ ] Update OpenAPI spec generation script
- [ ] Validate generated spec
- [ ] Publish updated API documentation

### Remove Relation from View
- [ ] Create RemoveRelationFromView command in ArchitectureViews
- [ ] Create RemoveRelationFromView command handler
- [ ] Create RelationRemovedFromView event
- [ ] Update read model projector to handle event
- [ ] Add DELETE /api/v1/views/{viewId}/relations/{relationId} endpoint
- [ ] Add HATEOAS removeFromView link to relation DTOs
- [ ] Add deleteRelationFromView method to frontend API client
- [ ] Add "Delete from View" option to canvas relation context menu
- [ ] Add toast notifications
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Write E2E tests
- [ ] Update documentation

### Enhanced Accessibility
- [ ] Implement Shift+F10 keyboard shortcut
- [ ] Audit color contrast in context menus
- [ ] Update CSS for WCAG AA compliance
- [ ] Complete NVDA screen reader testing
- [ ] Complete JAWS screen reader testing
- [ ] Complete VoiceOver screen reader testing
- [ ] Document screen reader usage patterns
- [ ] Fix any identified accessibility issues
- [ ] Create accessibility testing guide

### Final
- [ ] All enhancements tested and working
- [ ] Documentation updated
- [ ] User acceptance testing completed
- [ ] User sign-off
