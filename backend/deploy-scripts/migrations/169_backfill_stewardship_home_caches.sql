-- Migration: 169_backfill_stewardship_home_caches
-- Purpose: spec 228 — one-time seed of the Stewardship home caches added by
-- migration 168 from the supplier tables, so GET /home is complete the moment
-- the backend starts; the cache projectors keep them current from then on.
-- Only direct realizations and active edit grants are seeded, as the projectors
-- would hold them. Cross-schema reads are allowed because this is a backfill
-- (spec 209). Idempotent.

INSERT INTO stewardship.capability_cache (tenant_id, capability_id, name, level, parent_id, status, ea_owner)
SELECT tenant_id, id, name, level, NULLIF(parent_id, ''), status, NULLIF(ea_owner, '')
FROM capabilitymapping.capabilities
ON CONFLICT (tenant_id, capability_id) DO UPDATE SET
    name = EXCLUDED.name,
    level = EXCLUDED.level,
    parent_id = EXCLUDED.parent_id,
    status = EXCLUDED.status,
    ea_owner = EXCLUDED.ea_owner;

INSERT INTO stewardship.domain_assignment_cache (tenant_id, capability_id, domain_id)
SELECT tenant_id, capability_id, business_domain_id
FROM capabilitymapping.domain_capability_assignments
ON CONFLICT (tenant_id, capability_id, domain_id) DO NOTHING;

INSERT INTO stewardship.application_cache (tenant_id, component_id, name, ownership_state, owner_kind, owner_id)
SELECT tenant_id, id, name, ownership_state, NULLIF(owner_kind, ''), NULLIF(owner_id, '')
FROM architecturemodeling.application_components
WHERE is_deleted = FALSE
ON CONFLICT (tenant_id, component_id) DO UPDATE SET
    name = EXCLUDED.name,
    ownership_state = EXCLUDED.ownership_state,
    owner_kind = EXCLUDED.owner_kind,
    owner_id = EXCLUDED.owner_id;

INSERT INTO stewardship.realization_cache (tenant_id, realization_id, capability_id, component_id)
SELECT tenant_id, id, capability_id, component_id
FROM capabilitymapping.capability_realizations
WHERE origin = 'Direct'
ON CONFLICT (tenant_id, realization_id) DO UPDATE SET
    capability_id = EXCLUDED.capability_id,
    component_id = EXCLUDED.component_id;

INSERT INTO stewardship.time_assessment_cache (tenant_id, capability_id, component_id, grade, assessed_at)
SELECT tenant_id, capability_id, component_id, grade, assessed_at
FROM architecturedirection.time_assessments
ON CONFLICT (tenant_id, capability_id, component_id) DO UPDATE SET
    grade = EXCLUDED.grade,
    assessed_at = EXCLUDED.assessed_at;

INSERT INTO stewardship.edit_grant_cache (tenant_id, grant_id, artifact_type, artifact_id, grantee_email, expires_at)
SELECT tenant_id, id::text, artifact_type, artifact_id::text, grantee_email, expires_at
FROM accessdelegation.edit_grants
WHERE status = 'active'
ON CONFLICT (tenant_id, grant_id) DO UPDATE SET
    artifact_type = EXCLUDED.artifact_type,
    artifact_id = EXCLUDED.artifact_id,
    grantee_email = EXCLUDED.grantee_email,
    expires_at = EXCLUDED.expires_at;
