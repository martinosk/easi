-- Migration: Backfill the Architecture Direction reference and realization caches (spec 209)
-- Description: Seeds architecturedirection.realization_cache from Capability Mapping's direct
--              realizations and reconciles architecturedirection.reference_name_cache with the
--              application components and business domains that still exist, so component and
--              business-domain existence checks are complete and correct the moment the backend
--              starts. Rows for entities deleted since migration 119 are dropped, because the
--              cache is now the authority on whether a referenced entity exists. Idempotent.

INSERT INTO architecturedirection.realization_cache (tenant_id, realization_id, capability_id, component_id)
SELECT tenant_id, id, capability_id, component_id
FROM capabilitymapping.capability_realizations
WHERE origin = 'Direct'
ON CONFLICT (tenant_id, realization_id) DO UPDATE SET
    capability_id = EXCLUDED.capability_id,
    component_id = EXCLUDED.component_id;

INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
SELECT tenant_id, 'application', id, name
FROM architecturemodeling.application_components
WHERE is_deleted = FALSE
ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name)
SELECT tenant_id, 'business_domain', id, name
FROM capabilitymapping.business_domains
ON CONFLICT (tenant_id, entity_type, entity_id) DO UPDATE SET name = EXCLUDED.name;

DELETE FROM architecturedirection.reference_name_cache rnc
WHERE rnc.entity_type = 'application'
  AND NOT EXISTS (
      SELECT 1 FROM architecturemodeling.application_components ac
      WHERE ac.tenant_id = rnc.tenant_id AND ac.id = rnc.entity_id AND ac.is_deleted = FALSE
  );

DELETE FROM architecturedirection.reference_name_cache rnc
WHERE rnc.entity_type = 'business_domain'
  AND NOT EXISTS (
      SELECT 1 FROM capabilitymapping.business_domains bd
      WHERE bd.tenant_id = rnc.tenant_id AND bd.id = rnc.entity_id
  );
