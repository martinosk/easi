-- Create schemas for bounded context ownership
CREATE SCHEMA IF NOT EXISTS infrastructure;
CREATE SCHEMA IF NOT EXISTS shared;
CREATE SCHEMA IF NOT EXISTS architecturemodeling;
CREATE SCHEMA IF NOT EXISTS architectureviews;
CREATE SCHEMA IF NOT EXISTS capabilitymapping;
CREATE SCHEMA IF NOT EXISTS enterprisearchitecture;
CREATE SCHEMA IF NOT EXISTS viewlayouts;
CREATE SCHEMA IF NOT EXISTS importing;
CREATE SCHEMA IF NOT EXISTS platform;
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS accessdelegation;
CREATE SCHEMA IF NOT EXISTS metamodel;
CREATE SCHEMA IF NOT EXISTS releases;
CREATE SCHEMA IF NOT EXISTS valuestreams;

-- infrastructure
ALTER TABLE events SET SCHEMA infrastructure;
ALTER TABLE snapshots SET SCHEMA infrastructure;

-- shared
ALTER TABLE sessions SET SCHEMA shared;

-- architecturemodeling
ALTER TABLE application_components SET SCHEMA architecturemodeling;
ALTER TABLE component_relations SET SCHEMA architecturemodeling;
ALTER TABLE application_component_experts SET SCHEMA architecturemodeling;
ALTER TABLE acquired_entities SET SCHEMA architecturemodeling;
ALTER TABLE vendors SET SCHEMA architecturemodeling;
ALTER TABLE internal_teams SET SCHEMA architecturemodeling;
ALTER TABLE acquired_via_relationships SET SCHEMA architecturemodeling;
ALTER TABLE purchased_from_relationships SET SCHEMA architecturemodeling;
ALTER TABLE built_by_relationships SET SCHEMA architecturemodeling;

-- architectureviews
ALTER TABLE architecture_views SET SCHEMA architectureviews;
ALTER TABLE view_element_positions SET SCHEMA architectureviews;
ALTER TABLE view_preferences SET SCHEMA architectureviews;

-- capabilitymapping
ALTER TABLE capabilities SET SCHEMA capabilitymapping;
ALTER TABLE capability_dependencies SET SCHEMA capabilitymapping;
ALTER TABLE capability_realizations SET SCHEMA capabilitymapping;
ALTER TABLE capability_experts SET SCHEMA capabilitymapping;
ALTER TABLE capability_tags SET SCHEMA capabilitymapping;
ALTER TABLE capability_component_cache SET SCHEMA capabilitymapping;
ALTER TABLE domain_capability_assignments SET SCHEMA capabilitymapping;
ALTER TABLE effective_capability_importance SET SCHEMA capabilitymapping;
ALTER TABLE application_fit_scores SET SCHEMA capabilitymapping;
ALTER TABLE cm_strategy_pillar_cache SET SCHEMA capabilitymapping;
ALTER TABLE strategy_importance SET SCHEMA capabilitymapping;
ALTER TABLE domain_composition_view SET SCHEMA capabilitymapping;
ALTER TABLE business_domains SET SCHEMA capabilitymapping;
ALTER TABLE cm_effective_business_domain SET SCHEMA capabilitymapping;

-- enterprisearchitecture
ALTER TABLE enterprise_capabilities SET SCHEMA enterprisearchitecture;
ALTER TABLE enterprise_capability_links SET SCHEMA enterprisearchitecture;
ALTER TABLE enterprise_strategic_importance SET SCHEMA enterprisearchitecture;
ALTER TABLE domain_capability_metadata SET SCHEMA enterprisearchitecture;
ALTER TABLE capability_link_blocking SET SCHEMA enterprisearchitecture;
ALTER TABLE ea_strategy_pillar_cache SET SCHEMA enterprisearchitecture;
ALTER TABLE ea_realization_cache SET SCHEMA enterprisearchitecture;
ALTER TABLE ea_importance_cache SET SCHEMA enterprisearchitecture;
ALTER TABLE ea_fit_score_cache SET SCHEMA enterprisearchitecture;

-- viewlayouts
ALTER TABLE layout_containers SET SCHEMA viewlayouts;
ALTER TABLE element_positions SET SCHEMA viewlayouts;

-- importing
ALTER TABLE import_sessions SET SCHEMA importing;

-- platform
ALTER TABLE tenants SET SCHEMA platform;
ALTER TABLE tenant_domains SET SCHEMA platform;
ALTER TABLE tenant_oidc_configs SET SCHEMA platform;

-- auth
ALTER TABLE users SET SCHEMA auth;
ALTER TABLE invitations SET SCHEMA auth;

-- accessdelegation
ALTER TABLE edit_grants SET SCHEMA accessdelegation;

-- metamodel
ALTER TABLE meta_model_configurations SET SCHEMA metamodel;

-- releases
ALTER TABLE releases SET SCHEMA releases;

-- valuestreams
ALTER TABLE value_streams SET SCHEMA valuestreams;
ALTER TABLE value_stream_stages SET SCHEMA valuestreams;
ALTER TABLE value_stream_stage_capabilities SET SCHEMA valuestreams;
ALTER TABLE value_stream_capability_cache SET SCHEMA valuestreams;
