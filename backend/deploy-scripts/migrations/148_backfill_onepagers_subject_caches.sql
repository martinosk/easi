-- Migration: Backfill the OnePagers subject caches (spec 209)
-- Description: Seeds the caches added by 147 from the supplier tables so a one-pager renders
--              completely the moment the backend starts, without waiting for any event.
--              Cross-schema reads are permitted here because the filename marks this as a backfill.
--
--   built_in_fields carries the COMPLETE set of attributes each supplier publishes on its subject
--   Created/Updated events, keyed by the published (camelCase) attribute name. The catalogue entry
--   to attribute mapping lives only in OnePagers' BuiltInFieldSource, so adding a built-in field
--   over an attribute that is already published needs no projector change and no migration.

-- ============================================================================
-- Built-in attributes
-- ============================================================================

UPDATE onepagers.one_pager_subject_index idx
SET built_in_fields = idx.built_in_fields || jsonb_build_object(
        'description', s.description,
        'parentId', s.parent_id,
        'level', s.level,
        'strategyPillar', s.strategy_pillar,
        'pillarWeight', s.pillar_weight,
        'maturityValue', COALESCE(s.maturity_value, 0),
        'ownershipModel', s.ownership_model,
        'primaryOwner', s.primary_owner,
        'eaOwner', s.ea_owner,
        'status', s.status,
        'experts', COALESCE((
            SELECT jsonb_agg(jsonb_build_object(
                       'expertName', e.expert_name,
                       'expertRole', e.expert_role,
                       'contactInfo', e.contact_info
                   ) ORDER BY e.added_at)
            FROM capabilitymapping.capability_experts e
            WHERE e.tenant_id = s.tenant_id AND e.capability_id = s.id
        ), '[]'::jsonb)
    )
FROM capabilitymapping.capabilities s
WHERE idx.tenant_id = s.tenant_id AND idx.subject_type = 'capability' AND idx.subject_id = s.id;

UPDATE onepagers.one_pager_subject_index idx
SET built_in_fields = idx.built_in_fields || jsonb_build_object(
        'description', s.description,
        'category', s.category,
        'active', s.active,
        'targetMaturity', s.target_maturity
    )
FROM enterprisearchitecture.enterprise_capabilities s
WHERE idx.tenant_id = s.tenant_id AND idx.subject_type = 'enterprise-capability' AND idx.subject_id = s.id;

UPDATE onepagers.one_pager_subject_index idx
SET built_in_fields = idx.built_in_fields || jsonb_build_object(
        'description', s.description,
        'experts', COALESCE((
            SELECT jsonb_agg(jsonb_build_object(
                       'expertName', e.expert_name,
                       'expertRole', e.expert_role,
                       'contactInfo', e.contact_info
                   ) ORDER BY e.added_at)
            FROM architecturemodeling.application_component_experts e
            WHERE e.tenant_id = s.tenant_id AND e.component_id = s.id
        ), '[]'::jsonb)
    )
FROM architecturemodeling.application_components s
WHERE idx.tenant_id = s.tenant_id AND idx.subject_type = 'application' AND idx.subject_id = s.id;

UPDATE onepagers.one_pager_subject_index idx
SET built_in_fields = idx.built_in_fields || jsonb_build_object(
        'acquisitionDate', to_char(s.acquisition_date::timestamp, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
        'integrationStatus', s.integration_status,
        'notes', s.notes
    )
FROM architecturemodeling.acquired_entities s
WHERE idx.tenant_id = s.tenant_id AND idx.subject_type = 'acquired-entity' AND idx.subject_id = s.id;

UPDATE onepagers.one_pager_subject_index idx
SET built_in_fields = idx.built_in_fields || jsonb_build_object(
        'implementationPartner', s.implementation_partner,
        'notes', s.notes
    )
FROM architecturemodeling.vendors s
WHERE idx.tenant_id = s.tenant_id AND idx.subject_type = 'vendor' AND idx.subject_id = s.id;

UPDATE onepagers.one_pager_subject_index idx
SET built_in_fields = idx.built_in_fields || jsonb_build_object(
        'department', s.department,
        'contactPerson', s.contact_person,
        'notes', s.notes
    )
FROM architecturemodeling.internal_teams s
WHERE idx.tenant_id = s.tenant_id AND idx.subject_type = 'internal-team' AND idx.subject_id = s.id;

-- ============================================================================
-- Business domain names (the only relation target that is not a one-pager subject)
-- ============================================================================

INSERT INTO onepagers.business_domain_name_cache (tenant_id, business_domain_id, name)
SELECT d.tenant_id, d.id, COALESCE(d.name, '')
FROM capabilitymapping.business_domains d
ON CONFLICT (tenant_id, business_domain_id) DO UPDATE SET name = EXCLUDED.name;

-- ============================================================================
-- Relations
-- ============================================================================

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'capability', r.capability_id, 'realizing-applications', 'application', r.component_id, '',
       COALESCE(NULLIF(r.source_realization_id, ''), r.id)
FROM capabilitymapping.capability_realizations r
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'application', r.component_id, 'realized-capabilities', 'capability', r.capability_id, '',
       COALESCE(NULLIF(r.source_realization_id, ''), r.id)
FROM capabilitymapping.capability_realizations r
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT DISTINCT ON (d.tenant_id, d.source_capability_id, d.target_capability_id)
       d.tenant_id, 'capability', d.source_capability_id, 'depends-on', 'capability', d.target_capability_id, '', d.id
FROM capabilitymapping.capability_dependencies d
ORDER BY d.tenant_id, d.source_capability_id, d.target_capability_id, d.created_at
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT a.tenant_id, 'capability', a.capability_id, 'business-domains', '', a.business_domain_id,
       COALESCE(d.name, a.business_domain_name, ''), a.assignment_id
FROM capabilitymapping.domain_capability_assignments a
LEFT JOIN capabilitymapping.business_domains d
       ON d.tenant_id = a.tenant_id AND d.id = a.business_domain_id
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT c.tenant_id, 'capability', c.id, 'parent-capability', 'capability', c.parent_id, '', ''
FROM capabilitymapping.capabilities c
WHERE c.parent_id IS NOT NULL AND c.parent_id <> ''
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT c.tenant_id, 'capability', c.parent_id, 'child-capabilities', 'capability', c.id, '', ''
FROM capabilitymapping.capabilities c
WHERE c.parent_id IS NOT NULL AND c.parent_id <> ''
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'application', r.component_id, 'built-by', 'internal-team', r.internal_team_id, '', r.id
FROM architecturemodeling.built_by_relationships r
WHERE COALESCE(r.is_deleted, FALSE) = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'internal-team', r.internal_team_id, 'built-applications', 'application', r.component_id, '', r.id
FROM architecturemodeling.built_by_relationships r
WHERE COALESCE(r.is_deleted, FALSE) = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'application', r.component_id, 'purchased-from', 'vendor', r.vendor_id, '', r.id
FROM architecturemodeling.purchased_from_relationships r
WHERE COALESCE(r.is_deleted, FALSE) = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'vendor', r.vendor_id, 'purchased-applications', 'application', r.component_id, '', r.id
FROM architecturemodeling.purchased_from_relationships r
WHERE COALESCE(r.is_deleted, FALSE) = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'application', r.component_id, 'acquired-via', 'acquired-entity', r.acquired_entity_id, '', r.id
FROM architecturemodeling.acquired_via_relationships r
WHERE COALESCE(r.is_deleted, FALSE) = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT r.tenant_id, 'acquired-entity', r.acquired_entity_id, 'acquired-applications', 'application', r.component_id, '', r.id
FROM architecturemodeling.acquired_via_relationships r
WHERE COALESCE(r.is_deleted, FALSE) = FALSE
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

INSERT INTO onepagers.subject_relation_cache
    (tenant_id, subject_type, subject_id, entry_id, related_type, related_id, related_name, edge_id)
SELECT DISTINCT ON (cr.tenant_id, cr.source_component_id, cr.target_component_id)
       cr.tenant_id, 'application', cr.source_component_id, 'component-relations', 'application', cr.target_component_id, '', cr.id
FROM architecturemodeling.component_relations cr
WHERE COALESCE(cr.is_deleted, FALSE) = FALSE
ORDER BY cr.tenant_id, cr.source_component_id, cr.target_component_id, cr.created_at
ON CONFLICT (tenant_id, subject_type, subject_id, entry_id, related_id) DO UPDATE
SET related_type = EXCLUDED.related_type, related_name = EXCLUDED.related_name, edge_id = EXCLUDED.edge_id;

-- ============================================================================
-- Maturity scale
-- ============================================================================

INSERT INTO onepagers.maturity_scale_cache (tenant_id, sections)
SELECT DISTINCT ON (c.tenant_id) c.tenant_id, COALESCE(c.sections, '[]'::jsonb)
FROM metamodel.meta_model_configurations c
ORDER BY c.tenant_id, c.version DESC
ON CONFLICT (tenant_id) DO UPDATE SET sections = EXCLUDED.sections;
