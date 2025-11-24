-- Migration: Seed Release Notes
-- Spec: 043_Release_Notes_done.md
-- Description: Seeds historical release notes data

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.1.0', '2025-11-10', '## What''s New in 0.1.0

### Major
- **Application Component Modeling**: Create and manage application components using ArchiMate standards
- **Component Relations**: Define relationships between components (Triggers, Serves)
- **Event Sourcing Backend**: Built on CQRS with event sourcing for full audit trail
- **Canvas-Based Modeler**: Visual drag-and-drop interface for designing system architecture
- **Navigation Tree**: Hierarchical view of all components and relations
- **Multiple Views**: Create and manage multiple architecture views with independent layouts

### API
- `POST /api/v1/components` - Create components
- `GET /api/v1/components` - List all components
- `POST /api/v1/relations` - Create relations
- `POST /api/v1/views` - Create architecture views', '2025-11-10 17:44:35'),

('0.2.0', '2025-11-16', '## What''s New in 0.2.0

### Major
- **Testing Strategy Phase 1**: Stabilized existing test suite with proper async handling
- **Testing Strategy Phase 2**: Refactored tests for better maintainability
- **Playwright E2E Tests**: Added end-to-end browser testing
- **CI/CD Pipeline**: Automated builds and deployments via Azure Pipelines
- **Database Migrations**: Structured migration system for schema changes
- **Production Deployment**: Kubernetes deployment with Traefik ingress

### Bugs
- Fixed connection pool issues in event store
- Fixed frontend routing for production paths', '2025-11-16 19:19:19'),

('0.3.0', '2025-11-18', '## What''s New in 0.3.0

### Major
- **Delete Components**: Remove components from the architecture model
- **Delete Relations**: Remove relationships between components
- **Context Menu**: Right-click context menu for quick actions on canvas
- **Automatic View Layout**: Views now auto-arrange components for better visualization
- **Cascade Deletion**: Deleting a component removes associated relations

### API
- `DELETE /api/v1/components/{id}` - Delete component
- `DELETE /api/v1/relations/{id}` - Delete relation', '2025-11-18 18:06:03'),

('0.4.0', '2025-11-21', '## What''s New in 0.4.0

### Major
- **Capability Model Backend**: New bounded context for business capability mapping
- **Hierarchical Capabilities**: Support for L1-L4 capability levels (Domain → Sub-capability)
- **Capability Metadata**: Track maturity levels, ownership, and strategy alignment
- **Standardized API Responses**: Consistent REST Level 3 responses with HATEOAS links

### API
- `POST /api/v1/capabilities` - Create capability
- `GET /api/v1/capabilities` - List all capabilities
- `GET /api/v1/capabilities/{id}` - Get capability details
- `PUT /api/v1/capabilities/{id}` - Update capability', '2025-11-21 14:17:55'),

('0.5.0', '2025-11-22', '## What''s New in 0.5.0

### Major
- **Capability Tree View**: Hierarchical navigation of business capabilities in sidebar
- **Create Capabilities from Tree**: Add new capabilities directly from the navigation tree
- **Edit Capabilities**: Modify capability name, description, and metadata
- **Delete Capabilities**: Remove capabilities with proper cleanup
- **Auto-Level on Parent Change**: Moving a capability automatically adjusts its level

### Bugs
- Fixed frontend pod permissions for production
- Optimized local Docker build process', '2025-11-22 23:32:54'),

('0.6.0', '2025-11-23', '## What''s New in 0.6.0

### Major
- **Capabilities on Canvas**: Visualize business capabilities alongside components
- **Link Capabilities to Applications**: Connect capabilities to the systems that realize them
- **Inherited Realizations**: Child capabilities inherit parent''s system realizations
- **Realization Levels**: Track Full, Partial, or Planned realization status

### API
- `POST /api/v1/capabilities/{id}/systems` - Link system to capability
- `GET /api/v1/capabilities/{id}/systems` - Get systems realizing capability
- `GET /api/v1/capability-realizations/by-component/{id}` - Get capabilities by component', '2025-11-23 17:15:37'),

('0.7.0', '2025-11-24', '## What''s New in 0.7.0

### Major
- **Release Notes System**: Stay informed about new features and changes
- **Release Notes Overlay**: See what''s new on first launch after updates
- **Release Notes Browser**: Access full release history from the toolbar
- **UI Consistency Improvements**: Polished interface across capability management

### Bugs
- Fixed inherited capability realization when adding new parent
- Fixed various UI inconsistencies in capability views

### API
- `GET /api/v1/version` - Get current application version
- `GET /api/v1/releases` - List all releases
- `GET /api/v1/releases/latest` - Get latest release notes
- `GET /api/v1/releases/{version}` - Get specific version notes', '2025-11-24 15:20:36')

ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
