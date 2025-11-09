# Graphical Component Modeller - Frontend Feature

## Description
A visual canvas-based interface for modelling application components and their relations. Users can graphically create, position, and link components on a canvas, providing an intuitive way to design and visualise system architecture.

## Purpose
Enable architects to create an interactive diagramming tool that:
- Allows users to visually create and arrange application components on a canvas
- Enables users to draw relationships between components

## Integration Requirements
- **API-First Approach**: Frontend must consume ONLY the backend API endpoints as documented in the OpenAPI specification at `frontend/openapi.json`
- **OpenAPI Contract**: The OpenAPI specification is generated during backend build to `frontend/openapi.json`
- **HATEOAS Navigation**: Utilize hypermedia links provided in API responses for navigation

## Functional Requirements

### Canvas Operations
- Users should be able to add new application components to the canvas
- Users should be able to position components by dragging them
- Users should be able to create relations between components by drawing connections
- Component positions should persist. This should be implemented as a new bounded context in the backed called "Architecture Views"
- Users should be able to view component details
- Users should be able to view relation details
- Component and relation details must contain a "type" with link to the Archimate documentation as found in hatoas links.

### Component Visualization
- Components should be visually represented with their name and optionally description
- Components should be distinguishable and clearly visible on the canvas
- The visual design of components is left to frontend engineering discretion

### Relation Visualization
- Relations should be visually represented as connections between components
- Relations should clearly indicate direction (from source to target)
- Relation type (Triggers/Serves) should be visually distinguishable
- The visual representation (arrows, lines, curves) is left to frontend engineering discretion

### User Interactions
- Adding a component should prompt for name and optional description
- Creating a relation should prompt for relation type, optional name and description
- Users should be able to select components and relations to view details
- Error handling should provide clear feedback to users

## Technical Considerations
- Choose appropriate canvas/diagramming library (e.g., React Flow, Konva, D3, etc.)
- Implement proper state management for canvas state
- Handle API communication with proper error handling
- Consider performance for diagrams with many components
- Implement responsive design considerations

## Getting Started

The OpenAPI specification will be available at `frontend/openapi.json`
- Contains: All API endpoints, request/response models, validation rules, HATEOAS links

## Checklist

### Backend Prerequisites
- [ ] OpenAPI specification generation script created (`backend/scripts/generate-openapi.sh`)
- [ ] Documentation created for OpenAPI usage (`docs/OpenAPI-Access.md`)
- [ ] Backend for creating and retrieving architecture views (partially implemented - value objects, events created; aggregate, commands, handlers, API endpoints pending)

### Frontend Implementation
- [ ] Project setup with chosen framework and libraries (React + TypeScript + Vite + React Flow)
- [ ] API client generated or implemented based on OpenAPI spec (src/api/client.ts)
- [ ] Canvas component implemented with basic rendering (ComponentCanvas.tsx)
- [ ] Component creation functionality implemented (CreateComponentDialog.tsx)
- [ ] Component positioning/dragging implemented (React Flow with localStorage persistence - backend Architecture Views pending)
- [ ] Relation creation functionality implemented (CreateRelationDialog.tsx)
- [ ] Relation visualization implemented (React Flow edges with color coding)
- [ ] Component details view implemented (ComponentDetails.tsx with Archimate documentation links)
- [ ] Relation details view implemented (RelationDetails.tsx with Archimate documentation links)
- [ ] Error handling and user feedback implemented (error states and dialogs)
- [ ] Basic styling and UX polish (inline styles, intuitive interactions)

### Testing
- [ ] Unit tests for state management logic
- [ ] Unit tests for API communication layer (client.test.ts with vitest)
- [ ] Component rendering tests
- [ ] User interaction tests (component creation)
- [ ] User interaction tests (component dragging)
- [ ] User interaction tests (relation creation)
- [ ] End-to-end test: Create and position component
- [ ] End-to-end test: Create relation between components
- [ ] End-to-end test: Load existing components and relations
- [ ] Error handling test: Invalid component creation
- [ ] Error handling test: Invalid relation creation
- [ ] Error handling test: Network failure scenarios

### Documentation
- [ ] Frontend setup instructions (frontend/README.md)
- [ ] Development server setup (frontend/README.md)
- [ ] API integration documentation (docs/Frontend-Setup.md)
- [ ] Component architecture documentation (frontend/README.md, docs/Frontend-Setup.md)
- [ ] Testing approach documentation (frontend/README.md)

### Final
- [ ] User acceptance testing completed
- [ ] Performance verified with realistic data set
- [ ] User sign-off
