# Enterprise Capability Mapping Tool

## Description

A web-based platform for modelling, visualising, and managing DFDS’s enterprise capabilities, their ownership, maturity, system realization, dependencies, and alignment with digitalisation pillars. The tool enables Enterprise Architects and stakeholders to collaborate on a unified capability map that reflects the current and target state of DFDS’s digital enterprise landscape.

## Purpose

Provide a single, interactive source of truth for capability mapping and visualization that:

· Supports a multi-level capability hierarchy (L1–L4)

· Displays ownership, maturity, and realization

· Shows dependencies and relationships between capabilities

· Allows perspectives for different stakeholder views

· Enables versioning and change tracking over time

· Maps capabilities to strategic pillars (Always On, Grow, Transform)

## Integration Requirements

· **API-First Approach:** Frontend must consume only backend API endpoints documented in `capability-map/openapi.json`.

· **OpenAPI Contract:** The OpenAPI specification is generated during backend build and stored in `capability-map/openapi.json`.

· **HATEOAS Navigation:** Frontend must utilize hypermedia links from the API for navigation between related entities (capabilities, systems, owners, etc.).

· **Authentication:** Integrate with Azure AD SSO for authentication and role-based access (EA edit, Tribe edit, stakeholder view).

## Functional Requirements

### Capability Model

Supports 3–4 hierarchical levels: - L1: Business Domain (e.g., Customer Engagement, Operations, Finance) - L2: Capability Area (e.g., Digital Experience, Route Planning) - L3: Capability (e.g., Identity & Login, Route Optimization) - L4: Sub-capability/Component (optional) Users can expand, collapse, and navigate hierarchies. Aggregates maturity, ownership, and pillar alignment at higher levels.


#### Capability Metadata

· Name and Description

· Parent Hierarchy (L1–L4)

· Strategic Pillar Alignment (Always On, Grow, Transform) with optional weighting (%)

· Maturity Level (Initial, Developing, Defined, Managed, Optimizing)

· Ownership Model (Tribe-owned, Team-owned, Shared, Enterprise Service)

· Primary Owner (Tribe/Team + named individual)

· Enterprise Architect Owner

· Capability Experts (list of SMEs with role/contact)

· System Realization (linked system(s) supporting the capability)

· Status (Active, Planned, Deprecated, Decommissioned) /*ved ikke med de to sidste*/

· Last Updated (timestamp + editor)

· Custom Tags (e.g. 'Legacy', 'API-first', 'Cloud-native')

## Reference & Dependency Mapping

Users can define dependencies between capabilities (e.g., 'Identity & Login' depends on 'Customer Data Management'). Visualize cross-domain relationships. View dependencies as graph overlays or matrix views.

## Perspectives & Saved Views

Users can create and save custom perspectives. Saved views are shareable and exportable (read-only link). Example perspectives include 'Tribe X Portfolio', 'Transform Initiatives', and 'Legacy Modernization Targets'.

## Visualization

Interactive map using graph library (e.g., D3.js or Cytoscape). Supports color, size, and shape encoding for maturity, ownership, and strategic pillar. Includes zoom, pan, filter, tooltip, and drilldown functionality.

## Capability Realization

Each capability can be linked to system(s) that realize it. Ownership alignment between system and capability is visualized, and missing realization is flagged as a gap. Optionally, system lifecycle can be shown.

## Maturity Tracking

Five-level maturity scale visualized as color or progress bar. View maturity distribution by Tribe, Domain, or Pillar, and compare across versions or over time.

## Versioning & Change Tracking

Create named snapshots of the capability map. Compare versions and highlight changes (added/removed/updated capabilities). Store metadata: version name, author, timestamp, notes